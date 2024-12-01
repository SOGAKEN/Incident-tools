# ベースイメージとしてBunを使用
FROM oven/bun:1 AS builder
WORKDIR /app

# パッケージファイルをコピー
COPY package.json bun.lockb ./
RUN bun install --frozen-lockfile

# ソースコードをコピー
COPY . .

# Next.jsの環境変数を設定
ENV NEXT_TELEMETRY_DISABLED 1
ENV NODE_ENV production

# ビルドの実行
RUN bun run build

# 実行ステージ
FROM node:20-slim AS runner
WORKDIR /app

# 実行に必要な環境変数を設定
ENV NODE_ENV production
ENV NEXT_TELEMETRY_DISABLED 1
ENV PORT 8080
ENV HOST 0.0.0.0

# 必要なファイルのみをコピー
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public

# RUN npm install -g pm2
# 
# COPY docker/pm2.json ./pm2.json
# COPY docker/entrypoint.sh ./entrypoint.sh

# Cloud Run特有の設定
EXPOSE 8080

# アプリケーションの起動
CMD ["node", "server.js"]
#ENTRYPOINT ["/bin/sh", "./entrypoint.sh"]
