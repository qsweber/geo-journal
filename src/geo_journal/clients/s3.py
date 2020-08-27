from datetime import datetime
from decimal import Decimal
import typing
from uuid import uuid4

import boto3  # type: ignore


BUCKET_NAME = "geojournal-uploads"


class Image(typing.NamedTuple):
    name: str
    taken_at: datetime
    latitude: Decimal
    longitude: Decimal


class S3Client:
    def __init__(self) -> None:
        self.s3 = boto3.client("s3")

    def create_presigned_post(
        self,
        user_id: str,
        image: Image,
    ) -> typing.Tuple[str, typing.Dict[str, str]]:
        metadata = {
            "x-amz-meta-name": image.name,
            "x-amz-meta-latitude": str(image.latitude),
            "x-amz-meta-longitude": str(image.longitude),
            "x-amz-meta-taken": str(int(image.taken_at.timestamp())),
        }
        result = self.s3.generate_presigned_post(
            BUCKET_NAME,
            "{}/{}".format(user_id, str(uuid4())),
            Fields=metadata,
            Conditions=[{key: value} for key, value in metadata.items()],
            ExpiresIn=3600,
        )

        return result["url"], result["fields"]

    def get_images(
        self,
        user_id: str,
    ) -> typing.List[Image]:
        s3_objects = self.s3.list_objects_v2(Bucket=BUCKET_NAME, Prefix=user_id)
        if "Contents" not in s3_objects:
            return []

        return [
            self._image_from_raw_metadata(
                self.s3.head_object(Bucket=BUCKET_NAME, Key=s3_object["Key"])[
                    "Metadata"
                ]
            )
            for s3_object in s3_objects["Contents"]
        ]

    def _image_from_raw_metadata(self, metadata: typing.Dict[str, str]) -> Image:
        return Image(
            name=metadata["name"],
            taken_at=datetime.fromtimestamp(int(metadata["taken"])),
            latitude=Decimal(metadata["latitude"]),
            longitude=Decimal(metadata["longitude"]),
        )
