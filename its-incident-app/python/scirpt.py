import os
from datetime import datetime

import pytz
from google.cloud import datastore
from tabulate import tabulate

# Datastore Emulator環境設定
os.environ["DATASTORE_EMULATOR_HOST"] = "localhost:8090"
os.environ["GOOGLE_CLOUD_PROJECT"] = "local-project"

# Datastoreクライアント
client = datastore.Client()

# 日本時間 (Asia/Tokyo) ロケール
JST = pytz.timezone("Asia/Tokyo")

# Datastoreクエリ
query = client.query(kind="ServiceState")
results = list(query.fetch())


# 日本時間に変換するヘルパー関数
def to_japan_time(dt):
    if isinstance(dt, datetime):  # Datastoreの日付はdatetime型
        return dt.astimezone(JST)
    return dt


# 結果の整形とソート
if not results:
    print("No entities found")
else:
    # 各エンティティを変換してリスト化
    formatted_results = []
    for entity in results:
        row = {
            "ID/Name": entity.key.id_or_name,
            **{
                k: to_japan_time(v) if "time" in k or isinstance(v, datetime) else v
                for k, v in entity.items()
            },
        }
        formatted_results.append(row)

    # `created_at`でソート
    formatted_results.sort(key=lambda x: x.get("created_at", datetime.min))

    # 動的にヘッダーを生成
    all_keys = set(key for entity in formatted_results for key in entity.keys())
    headers = sorted(all_keys)  # ヘッダーをソートして一貫性を持たせる

    # 表データを生成
    table_data = [[e.get(h, "") for h in headers] for e in formatted_results]

    # 表形式で表示
    print(tabulate(table_data, headers=headers, tablefmt="grid"))
