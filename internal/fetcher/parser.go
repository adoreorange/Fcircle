package fetcher

import (
	"Fcircle/internal/model"
	"Fcircle/internal/utils"
	"fmt"
	"github.com/mmcdole/gofeed"
	"golang.org/x/net/html"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// FetchFriendArticles 请求并解析指定 friend 的 RSS，返回最新 maxCount 篇文章
func FetchFriendArticles(friend model.Friend, maxCount int) ([]model.Article, error) {

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			ResponseHeaderTimeout: 15 * time.Second,
			TLSHandshakeTimeout:   15 * time.Second,
		},
	}

	const maxRetry = 2

	var (
		resp *http.Response
		err  error
	)

	start := time.Now()

	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36 FcircleBot/1.0 (+https://github.com/TXM983/Fcircle;MuXiaoChen;+https://miraii.cn)"

	for attempt := 0; attempt <= maxRetry; attempt++ {
		req, e := http.NewRequest("GET", friend.RSS, nil)
		if e != nil {
			err = e
			break
		}

		req.Header.Set("User-Agent", userAgent)

		resp, err = client.Do(req)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	duration := time.Since(start)

	utils.Infof("抓取 [%s] RSS 用时: %v", friend.Name, duration)

	if err != nil {
		return nil, fmt.Errorf("RSS 请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RSS 请求失败，状态码: %d", resp.StatusCode)
	}

	parser := gofeed.NewParser()
	feed, err := parser.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("RSS 解析失败: %v", err)
	}

	articles := make([]model.Article, 0, maxCount)
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*60*60)
	}

	for i, item := range feed.Items {
		if i >= maxCount {
			break
		}

		pubTime := time.Now().In(loc)
		if item.PublishedParsed != nil {
			pubTime = item.PublishedParsed.In(loc)
		} else if item.UpdatedParsed != nil {
			pubTime = item.UpdatedParsed.In(loc)
		}

		formattedTime := pubTime.Format("2006-01-02 15:04:05")

		author := friend.Name
		if item.Author != nil && item.Author.Name != "" {
			author = item.Author.Name
		}

		content := getItemContent(item)

		cleanContent := ""
		if strings.TrimSpace(content) != "" {
			cleanContent = ExtractCleanHTML(content)       // 清理html标签
			cleanContent = safeTruncate(cleanContent, 250) // 字符截取
			cleanContent = insertNewlineAfterURL(cleanContent)
		}

		article := model.Article{
			Title:     item.Title,
			Link:      item.Link,
			Published: formattedTime,
			Author:    author,
			Avatar:    friend.Avatar,
			Content:   cleanContent,
			Url:       friend.URL,
		}
		articles = append(articles, article)
	}

	return articles, nil
}

func getItemContent(item *gofeed.Item) string {
	if item.Content != "" {
		return item.Content
	}
	if item.Description != "" {
		return item.Description
	}
	if exts, ok := item.Extensions["summary"]; ok {
		if vals, ok2 := exts[""]; ok2 && len(vals) > 0 {
			return vals[0].Value
		}
	}
	if exts, ok := item.Extensions["atom"]; ok {
		if vals, ok2 := exts["summary"]; ok2 && len(vals) > 0 {
			return vals[0].Value
		}
	}
	return ""
}

func insertNewlineAfterURL(text string) string {
	urlRegex := regexp.MustCompile(`https?://[^\s]*?\.shtml`)

	matches := urlRegex.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return text
	}

	var builder strings.Builder
	lastIndex := 0

	for _, match := range matches {
		start, end := match[0], match[1]

		builder.WriteString(text[lastIndex:start])
		builder.WriteString(text[start:end])

		if end < len(text) {
			nextChar := text[end]
			if nextChar != ' ' && nextChar != '\n' && nextChar != '\r' {
				builder.WriteByte(' ')
			}
		}

		lastIndex = end
	}

	if lastIndex < len(text) {
		builder.WriteString(text[lastIndex:])
	}

	return builder.String()
}

func safeTruncate(s string, maxChars int) string {
	cleaned := strings.ToValidUTF8(s, "")
	runes := []rune(cleaned)
	if len(runes) <= maxChars {
		return cleaned
	}

	truncated := string(runes[:maxChars])
	return fixBrokenHTML(truncated)
}

func fixBrokenHTML(s string) string {
	// 去掉最后不完整的标签
	if lastOpen := strings.LastIndex(s, "<"); lastOpen != -1 {
		if lastClose := strings.LastIndex(s, ">"); lastClose < lastOpen {
			s = s[:lastOpen]
		}
	}

	// 补全 a 标签
	aOpenCount := strings.Count(s, "<a")
	aCloseCount := strings.Count(s, "</a>")
	if aOpenCount > aCloseCount {
		s += "</a>"
	}

	return s
}

// ExtractCleanHTML 过滤和清理 HTML
func ExtractCleanHTML(htmlStr string) string {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return ""
	}

	var builder strings.Builder

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		switch n.Type {
		case html.TextNode:
			builder.WriteString(n.Data)
		case html.ElementNode:
			switch n.Data {
			case "a":
				builder.WriteString("<a")
				for _, attr := range n.Attr {
					if attr.Key == "href" && isSafeHref(attr.Val) {
						builder.WriteString(` href="` + html.EscapeString(attr.Val) + `"`)
					}
				}
				builder.WriteString(">")
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					traverse(c)
				}
				builder.WriteString("</a>")
			case "br":
				builder.WriteString("<br/>")
			default:
				// 其他标签忽略，只遍历子节点
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					traverse(c)
				}
			}

		default:
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				traverse(c)
			}
		}
	}

	traverse(doc)
	return strings.TrimSpace(builder.String())
}

// 简单判断 href 是否安全，防止 javascript: XSS 攻击
func isSafeHref(href string) bool {
	href = strings.ToLower(strings.TrimSpace(href))
	if strings.HasPrefix(href, "javascript:") || strings.HasPrefix(href, "data:") {
		return false
	}
	return true
}
