from datetime import datetime
from decimal import Decimal

# import requests
from uuid import uuid4

import boto3  # type: ignore


BUCKET_NAME = "geojournal-uploads"


class S3Client:
    def __init__(self) -> None:
        self.s3 = boto3.client("s3")

    def create_presigned_post(
        self,
        user_id: str,
        name: str,
        taken_at: datetime,
        latitude: Decimal,
        longitude: Decimal,
    ) -> str:
        url: str = self.s3.generate_presigned_post(
            BUCKET_NAME,
            "{}/{}".format(user_id, str(uuid4())),
            Fields={
                "x-amz-meta-name": name,
                "x-amz-meta-latitude": str(latitude),
                "x-amz-meta-longitude": str(longitude),
                "x-amz-meta-taken": str(taken_at.timestamp()),
            },
            Conditions=[
                {"x-amz-meta-name": name},
                {
                    "x-amz-meta-latitude": str(latitude),
                    "x-amz-meta-longitude": str(longitude),
                    "x-amz-meta-taken": str(taken_at.timestamp()),
                },
            ],
            ExpiresIn=3600,
        )

        return url


# result = s3.generate_presigned_post(
#     BUCKET_NAME,
#     "{}/{}".format(user_id, "foo6"),
#     Fields={"x-amz-meta-foo": "bar", "x-amz-meta-latitude": "123"},
#     Conditions=[{"x-amz-meta-foo": "bar"}, {"x-amz-meta-latitude": "123"}],
#     ExpiresIn=3600,
# )

# with open("tox.ini", "rb") as f:
#     foo = requests.post(
#         result["url"],
#         # data={ k: v for k, v in result['fields'].items() if k != 'x-amz-meta-foo' },
#         data=result["fields"],
#         files={"file": ("tox.ini", f)},
#         # headers={'x-amz-meta-foo': 'foo'}
#     )

# s3.head_object(Bucket=BUCKET_NAME, Key="quinn/foo6")["Metadata"]

