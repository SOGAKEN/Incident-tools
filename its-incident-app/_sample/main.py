# main.py
import os
import urllib.request
from urllib.error import URLError

import boto3


class RawEmailForwarder:
    def __init__(self):
        self.s3_client = boto3.client("s3")
        self.api_endpoint = os.environ["API_ENDPOINT"]
        self.bucket_name = os.environ["S3_BUCKET_NAME"]

    def forward_email(self, message_id: str) -> None:
        try:
            # S3からメールデータを取得
            response = self.s3_client.get_object(
                Bucket=self.bucket_name, Key=f"{message_id}"
            )

            # POSTリクエストの準備と送信
            data = response["Body"].read()
            req = urllib.request.Request(
                url=self.api_endpoint,
                data=data,
                method="POST",
                headers={"X-Message-ID": message_id},
            )

            with urllib.request.urlopen(req, timeout=300) as response:
                if response.status != 200:
                    print(f"API error status code: {response.status}")

        except Exception as e:
            print(f"Error processing message {message_id}: {str(e)}")
            raise


def lambda_handler(event, context):
    forwarder = RawEmailForwarder()

    for record in event.get("Records", []):
        message_id = record.get("ses", {}).get("mail", {}).get("messageId")
        if message_id:
            forwarder.forward_email(message_id)

    return {"statusCode": 200}