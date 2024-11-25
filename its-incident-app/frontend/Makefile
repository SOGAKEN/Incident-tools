# GCP設定
PROJECT_ID = docker-cloudran-works
REGION = asia-northeast1
REPOSITORY = incident-service

# サービス名
SERVICE = web-front

# Dockerイメージのプレフィックス（Artifact Registryを使用）
IMAGE_PREFIX = $(REGION)-docker.pkg.dev/$(PROJECT_ID)/$(REPOSITORY)

.PHONY: all build push deploy

# すべてのタスクを実行
all: build push deploy

# Dockerイメージをビルド
build:
	@echo "Building $(SERVICE)..."
	docker build --platform=linux/amd64 -t $(IMAGE_PREFIX)/$(SERVICE):latest .

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
		--allow-unauthenticated
