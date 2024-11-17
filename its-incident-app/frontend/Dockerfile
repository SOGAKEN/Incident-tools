# ベースイメージとして Bun を使用
FROM oven/bun:latest

# 作業ディレクトリを設定
WORKDIR /app

# パッケージファイルとロックファイルをコピー
COPY package.json bun.lockb ./

# 依存関係をインストール
RUN bun install

# アプリケーションコードをコピー
COPY . .

# アプリケーションをビルド
RUN bun run build

# ポートを公開
EXPOSE 8080

# Cloud Run 用の環境変数を設定
ENV PORT=8080

# アプリケーションを起動
CMD ["bun", "run", "start"]