from datetime import datetime
from decimal import Decimal
import typing
import os

from uuid import uuid4
import boto3  # type: ignore


BUCKET_NAME = "geojournal-uploads"


class S3Client:
    def __init__(self) -> None:
        if os.environ["STAGE"] == "PROD":
            self.s3 = boto3.client(
                "s3",
                aws_access_key_id=os.environ.get("S3_ACCESS_KEY"),
                aws_secret_access_key=os.environ.get("S3_SECRET_KEY"),
            )
        else:
            self.s3 = boto3.client("s3")

    def create_presigned_post(
        self,
        user_id: str,
        name: str,
        taken_at: datetime,
        latitude: Decimal,
        longitude: Decimal,
    ) -> typing.Tuple[str, typing.Dict[str, str]]:
        result = self.s3.generate_presigned_post(
            BUCKET_NAME,
            "{}/{}".format(user_id, str(uuid4())),
            Fields={
                "x-amz-meta-name": name,
                "x-amz-meta-latitude": str(latitude),
                "x-amz-meta-longitude": str(longitude),
                "x-amz-meta-taken": str(int(taken_at.timestamp())),
            },
            Conditions=[
                {"x-amz-meta-name": name},
                {"x-amz-meta-latitude": str(latitude)},
                {"x-amz-meta-longitude": str(longitude)},
                {"x-amz-meta-taken": str(int(taken_at.timestamp()))},
            ],
            ExpiresIn=3600,
        )

        return result["url"], result["fields"]
