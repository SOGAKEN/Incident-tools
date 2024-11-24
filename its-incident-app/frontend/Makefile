# GCP設定
PROJECT_ID = docker-cloudran-works
REGION = asia-northeast1
REPOSITORY = incident-service
SERVICE = web-front

# サービスアカウント名
SERVICE_ACCOUNT = $(SERVICE)
SERVICE_ACCOUNT_EMAIL = $(SERVICE_ACCOUNT)@$(PROJECT_ID).iam.gserviceaccount.com

# Dockerイメージのプレフィックス（Artifact Registryを使用）
IMAGE_PREFIX = $(REGION)-docker.pkg.dev/$(PROJECT_ID)/$(REPOSITORY)

# 環境変数の定義
PUBLIC_ENV_VARS = \
	NEXT_PUBLIC_AUTH_SERVICE_URL=https://auth-service-url.com \
	NEXT_PUBLIC_API_URL=https://api-url.com

# シークレット名の定義
SECRETS = \
	DATABASE_URL \
	AUTH_SERVICE_TOKEN \
	GOOGLE_CLIENT_SECRET

.PHONY: all build push deploy create-secrets set-secrets create-service-account set-permissions

# サービスアカウントを作成
create-service-account:
	@echo "Creating service account..."
	gcloud iam service-accounts create $(SERVICE_ACCOUNT) \
		--project $(PROJECT_ID) \
		--display-name="$(SERVICE) service account" || true
	@echo "Granting Cloud Run invoker role..."
	gcloud projects add-iam-policy-binding $(PROJECT_ID) \
		--member="serviceAccount:$(SERVICE_ACCOUNT_EMAIL)" \
		--role="roles/run.invoker" || true
	@echo "Granting Secret Manager accessor role..."
	gcloud projects add-iam-policy-binding $(PROJECT_ID) \
		--member="serviceAccount:$(SERVICE_ACCOUNT_EMAIL)" \
		--role="roles/secretmanager.secretAccessor" || true

# シークレットを作成
create-secrets:
	@echo "Creating secrets..."
	@for secret in $(SECRETS); do \
		echo "Creating secret: $$secret"; \
		gcloud secrets create $$secret \
			--project $(PROJECT_ID) \
			--replication-policy="automatic" || true; \
	done

# シークレットの値を設定（対話的）
set-secrets:
	@echo "Setting secret values..."
	@for secret in $(SECRETS); do \
		echo "Enter value for $$secret:"; \
		read value; \
		echo -n "$$value" | gcloud secrets versions add $$secret \
			--project $(PROJECT_ID) \
			--data-file=-; \
	done

# サービスアカウントにシークレットアクセス権限を付与
set-permissions: create-service-account
	@echo "Setting permissions..."
	@for secret in $(SECRETS); do \
		gcloud secrets add-iam-policy-binding $$secret \
			--project $(PROJECT_ID) \
			--member="serviceAccount:$(SERVICE_ACCOUNT_EMAIL)" \
			--role="roles/secretmanager.secretAccessor"; \
	done

# シークレットの一覧を表示
list-secrets:
	@echo "Listing secrets..."
	gcloud secrets list --project $(PROJECT_ID)

# Dockerイメージをビルド
build:
	@echo "Building $(SERVICE)..."
	docker build --platform linux/amd64 -t $(IMAGE_PREFIX)/$(SERVICE):latest .

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
		--service-account $(SERVICE_ACCOUNT_EMAIL) \
		--allow-unauthenticated \
		--project $(PROJECT_ID) \
		$(foreach var,$(PUBLIC_ENV_VARS),--set-env-vars $(var)) \
		$(foreach secret,$(SECRETS),--set-secrets $(secret)=$(secret):latest)

# 開発環境用の.env.localファイルを生成
create-env-local:
	@echo "Creating .env.local file..."
	@echo "# Public environment variables" > .env.local
	@for var in $(PUBLIC_ENV_VARS); do \
		echo $$var >> .env.local; \
	done
	@echo "\n# Secret environment variables" >> .env.local
	@for secret in $(SECRETS); do \
		echo "Enter development value for $$secret:"; \
		read value; \
		echo "$$secret=$$value" >> .env.local; \
	done

# すべての設定を一括で行う
setup-all: create-service-account create-secrets set-secrets set-permissions create-env-local

# すべてのタスクを実行
all: build push deploy

# 使い方を表示
help:
	@echo "Available commands:"
	@echo "  make create-service-account  - Create s
