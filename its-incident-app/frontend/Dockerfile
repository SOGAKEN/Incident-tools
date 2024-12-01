# ビルドステージ
FROM oven/bun:1 AS builder
WORKDIR /app

# パッケージファイルをコピー
COPY package.json bun.lockb ./
RUN bun install --frozen-lockfile

ENV TZ=Asia/Tokyo
ENV NOTIFY_SERVICE_URL=http://localhost:8083
ENV AUTH_URL=http://localhost:8082
ENV DBPILOT_URL=http://localhost:8081/api/v1

# ソースコードをコピー
COPY . .
# ビルド時に渡す環境変数を定義
ARG NEXT_PUBLIC_MAIN_NAME
ARG NEXT_PUBLIC_STATUS_A
ARG NEXT_PUBLIC_STATUS_B
ARG NEXT_PUBLIC_AUTH_SERVICE_URL

# 環境変数をビルドプロセスに埋め込む
ENV NEXT_PUBLIC_MAIN_NAME=$NEXT_PUBLIC_MAIN_NAME
ENV NEXT_PUBLIC_STATUS_A=$NEXT_PUBLIC_STATUS_A
ENV NEXT_PUBLIC_STATUS_B=$NEXT_PUBLIC_STATUS_B
ENV NEXT_PUBLIC_AUTH_SERVICE_URL=$NEXT_PUBLIC_AUTH_SERVICE_URL

# ビルド時に必要な環境変数は、Dockerfileに含めず、CLIで渡します
ENV NEXT_TELEMETRY_DISABLED=1
ENV NODE_ENV=production

# Next.jsアプリケーションをビルド
RUN bun run build || exit 1

# 実行ステージ
FROM node:20-slim AS runner
WORKDIR /app

# 実行時の環境変数はCloud Runに設定
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV TZ=Asia/Tokyo

# 必要なファイルのみをコピー
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public

# ポートを公開
EXPOSE 8080

# アプリケーションを起動
CMD ["node", "server.js"]
