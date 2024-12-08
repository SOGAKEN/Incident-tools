# GCP設定
PROJECT_ID = docker-cloudran-works
REGION = asia-northeast1
REPOSITORY = incident-service

# サービス名
SERVICE = web-front

# Dockerイメージのプレフィックス（Artifact Registryを使用）
IMAGE_PREFIX = $(REGION)-docker.pkg.dev/$(PROJECT_ID)/$(REPOSITORY)

# 環境変数の定義
# ビルド時環境変数
NEXT_PUBLIC_MAIN_NAME = 開発
NEXT_PUBLIC_AUTH_SERVICE_URL = http://localhost:8082

# ランタイム環境変数
ENV_VARS = TZ=Asia/Tokyo,NOTIFY_SERVICE_URL=http://localhost:8083,AUTH_URL=http://localhost:8082,DBPILOT_URL=http://localhost:8081/api/v1

.PHONY: all build push deploy

# すべてのタスクを実行
all: build push deploy

# Dockerイメージをビルド
build:
	@echo "Building $(SERVICE)..."
	docker build --platform=linux/amd64 \
		--build-arg NEXT_PUBLIC_MAIN_NAME=$(NEXT_PUBLIC_MAIN_NAME) \
		--build-arg NEXT_PUBLIC_AUTH_SERVICE_URL=$(NEXT_PUBLIC_AUTH_SERVICE_URL) \
		-t $(IMAGE_PREFIX)/$(SERVICE):latest .

# Dockerイメージをプッシュ
push:
	@echo "Pushing $(SERVICE)..."
	docker push $(IMAGE_PREFIX)/$(SERVICE):latest

# サービスをデプロイ
deploy:
	@echo "Deploying $(SERVICE) to Cloud Run..."
	gcloud run deploy $(SERVICE) \
		--image $(IMAGE_PREFIX)/$(SERVICE):latest \
		--region $(REGION) \
		--platform managed \
		--allow-unauthenticated \
		--set-env-vars $(ENV_VARS)
