# env

```bash

SERVER_PORT=        # サーバーが待ち受けるポート番号
GIN_MODE=release       # Ginのモード (debug/release)
LOG_LEVEL=info        # ログレベル (debug/info/warn/error)
ENVIRONMENT=production # 実行環境 (development/production)

DBPILOT_URL=""        # DBPilotサービスのベースURL
SERVICE_TOKEN=""      # サービス認証用トークン

SHUTDOWN_TIMEOUT=10s   # シャットダウン時の待機時間
HTTP_READ_TIMEOUT=15s  # リクエスト読み込みタイムアウト
HTTP_WRITE_TIMEOUT=90s # レスポンス書き込みタイムアウト
HTTP_IDLE_TIMEOUT=60s  # アイドル接続のタイムアウト

AI_ENDPOINT=""         # AI APIのエンドポイントURL
AI_TOKEN=""           # AI API認証用トークン
AI_SHORT_TIMEOUT=30s  # 短時間処理用タイムアウト
AI_LONG_TIMEOUT=90s   # 長時間処理用タイムアウト
AI_MAX_RETRIES=3      # 最大リトライ回数
AI_MIN_RETRY_DELAY=2s # 最小リトライ待機時間
AI_MAX_RETRY_DELAY=5s # 最大リトライ待機時間

# GCP設定
K_SERVICE=auto-pilot  # Cloud Run サービス名
DATASTORE_EMULATOR_HOST=
GOOGLE_CLOUD_PROJECT=

```
