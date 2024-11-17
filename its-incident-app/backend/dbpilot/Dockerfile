# ビルドステージ
FROM golang:1.23 AS builder

WORKDIR /app

# 依存関係をコピー＆インストール
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# バイナリをビルド
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main

# 実行ステージ
FROM gcr.io/distroless/static-debian11

# 作業ディレクトリを作成
WORKDIR /app

# バイナリをコピー
COPY --from=builder /app/main /app/main

# Cloud Run用の環境変数
ENV PORT=8080

# アプリケーションが使用するポートを公開
EXPOSE ${PORT}

# 非rootユーザーで実行
USER nonroot:nonroot

# アプリケーションの起動（パスを修正）
CMD ["/app/main"]
