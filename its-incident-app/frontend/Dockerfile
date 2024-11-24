# ビルドステージ
FROM oven/bun:latest AS builder
WORKDIR /app

COPY package.json bun.lockb ./
RUN bun install

COPY . .
RUN bun run build

# 実行ステージ
FROM node:20-slim AS runner
WORKDIR /app

# 必要なファイルのコピー
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public
COPY --from=builder /app/middleware.ts ./middleware.ts

# 環境変数の設定
ENV NODE_ENV=production
ENV PORT=8080

# アプリケーションの起動
CMD ["node", "server.js"]
