# Fcircle - 友链聚合文章抓取工具

`Fcircle` 是一个基于 Go 编写的轻量 RSS 聚合爬虫工具，用于定时抓取你指定的友链 RSS 信息，生成统一的 JSON 文件供前端展示使用。

## ✨ 功能简介

- 定时从远程友链配置文件中读取 RSS 地址
- 支持并发抓取每个友链的最新文章
- 自动生成 `feed_result.json` 文件
- 提供 HTTP 接口手动触发抓取或读取结果
- 可通过 Docker Compose 部署，支持环境变量灵活配置

---

## 🚀 快速启动

### 1. 编写 `docker-compose.yml`

```yaml
version: '3.8'

services:
  fcircle:
    image: txm123/fcircle:latest
    container_name: fcircle
    restart: always
    ports:
      - "8521:8080"
    volumes:
      - ./logs:/app/output
    environment:
      - GIN_MODE=release                    # 设定 Gin 运行环境
      - SERVER_PORT=8080                     # 对应 容器启动端口
      - SECRET_KEY=#####                     # /fetch接口手动请求密钥
      - CRON_EXPR=0 0 3 * * *           # 设置定时调用的间隔时间
      - CONFIG_URL=https://source.adoreorg.cn/json/RSS.json  # 配置文件url
      - OUTPUT_FILE=output/feed_result.json   # 朋友圈json文件路径
      - LOG_FILE=output/crawl.log   # 日志文件路径

```

### 2. 启动容器
```shell
docker-compose up -d
```
容器启动后，将会立即执行一次抓取任务，并每隔指定时间自动运行。


## ⚙️ 配置说明

该项目默认不再使用本地配置文件，而是通过环境变量进行控制，以下为主要环境变量说明：

| 环境变量          | 说明             | 示例值                             |
|---------------|----------------|---------------------------------|
| `SERVER_PORT` | HTTP 服务监听端口    | `8080`                          |
| `SECRET_KEY`  | fetch接口请求密钥    | `#####`                         |
| `CRON_EXPR`   | cron表达式        | `0 0 3 * * *`                   |
| `CONFIG_URL`  | 远程 RSS 配置文件地址  | `https://xxx.com/path/RSS.json` |
| `OUTPUT_FILE` | 抓取结果保存路径（容器内）  | `output/feed_result.json`       |
| `LOG_FILE`    | 日志输出路径（容器内）    | `output/crawl.log`              |


## 📦 输出文件说明

默认抓取完成后会生成如下 JSON 文件：
```yaml
output/
├── crawl.log           # 抓取日志
└── feed_result.json    # 抓取到的文章信息
```

在配置完代理之后，可以通过`/feed`来获取`feed_result.json`
例如：https://feed.miraii.cn/feed

可以通过`/fetch?key=xxxxxx`来重新解析RSS，其中的key为docker环境变量中设置的SECRET_KEY，以上接口都根据IP做了限流处理，请不要滥用哦！具体限流速率可以到源代码中查看。
