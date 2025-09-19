#!/bin/sh

# 设置应用程序的名称
APP_NAME="fcircle"

# 创建必要的目录
echo "📁 创建必要的目录..."
mkdir -p /app/output
chmod 755 /app/output

# 检查进程是否存在
pid=$(pgrep -f $APP_NAME)

if [ -n "$pid" ]; then
    echo "Process $APP_NAME is running, killing it..."
    kill -9 $pid
    sleep 2
else
    echo "Process $APP_NAME is not running."
fi

# 启动新的应用程序
echo "Starting $APP_NAME..."
exec /app/$APP_NAME
