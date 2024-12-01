#!/bin/sh

set -e
# 環境変数の設定（必要に応じて）
export NEXT_PUBLIC_MAIN_NAME=開発
export NEXT_PUBLIC_STATUS_A=未着手
export NEXT_PUBLIC_STATUS_B=調査中
export NEXT_PUBLIC_AUTH_SERVICE_URL=http://localhost:8082
# PM2でアプリケーションを起動
#
#npm run pm2
HOST=0.0.0.0 pm2 start ./pm2-stage.json --no-daemon
