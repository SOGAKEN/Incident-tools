# ビルドステージ
FROM golang:1.23

# ワーキングディレクトリを設定
WORKDIR /app

# 依存関係をキャッシュするためにgo.modとgo.sumをコピー
COPY go.mod go.sum ./
RUN go mod download

# 残りのソースコードをコピー
COPY ./ ./

# アプリケーションをビルド
RUN go build -o /main

# 実行ステージ
# FROM gcr.io/distroless/base-debian11

# ビルドしたバイナリをコピー
#   COPY --from=builder /app/main /

# ポートを指定
EXPOSE 8080

# コンテナ起動時に実行するコマンド
CMD ["/main"]
