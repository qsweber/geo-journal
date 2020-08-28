from datetime import datetime
from decimal import Decimal
import typing
from uuid import uuid4

import boto3  # type: ignore


BUCKET_NAME = "geojournal-uploads"


class ImageMetadata(typing.NamedTuple):
    name: str
    taken_at: datetime
    latitude: Decimal
    longitude: Decimal


class Image(typing.NamedTuple):
    metadata: ImageMetadata
    presigned_url: str


class S3Client:
    def __init__(self) -> None:
        self.s3 = boto3.client("s3")

    def _generate_presigned_post(
        self, *args: typing.Any, **kwargs: typing.Any
    ) -> typing.Any:
        return self.s3.generate_presigned_post(*args, **kwargs)

    def create_presigned_post(
        self,
        user_id: str,
        imageMetadata: ImageMetadata,
    ) -> typing.Tuple[str, typing.Dict[str, str]]:
        metadata = {
            "x-amz-meta-name": imageMetadata.name,
            "x-amz-meta-latitude": str(imageMetadata.latitude),
            "x-amz-meta-longitude": str(imageMetadata.longitude),
            "x-amz-meta-taken": str(int(imageMetadata.taken_at.timestamp())),
        }
        result = self._generate_presigned_post(
            BUCKET_NAME,
            "{}/{}".format(user_id, str(uuid4())),
            Fields=metadata,
            Conditions=[{key: value} for key, value in metadata.items()],
            ExpiresIn=3600,
        )

        return result["url"], result["fields"]

    def _list_objects_v2(self, *args: typing.Any, **kwargs: typing.Any) -> typing.Any:
        return self.s3.list_objects_v2(*args, **kwargs)

    def _head_object(self, *args: typing.Any, **kwargs: typing.Any) -> typing.Any:
        return self.s3.head_object(*args, **kwargs)

    def _generate_presigned_url(self, *args: typing.Any, **kwargs: typing.Any) -> str:
        return typing.cast(str, self.s3.generate_presigned_url(*args, **kwargs))

    def get_images(
        self,
        user_id: str,
    ) -> typing.List[Image]:
        s3_objects = self._list_objects_v2(Bucket=BUCKET_NAME, Prefix=user_id)

        if "Contents" not in s3_objects:
            return []

        return [
            self.get_image(s3_object["Key"]) for s3_object in s3_objects["Contents"]
        ]

    def get_image(self, key: str) -> Image:
        metadata = self._head_object(Bucket=BUCKET_NAME, Key=key)["Metadata"]
        return Image(
            metadata=ImageMetadata(
                name=metadata["name"],
                taken_at=datetime.fromtimestamp(int(metadata["taken"])),
                latitude=Decimal(metadata["latitude"]),
                longitude=Decimal(metadata["longitude"]),
            ),
            presigned_url=self._generate_presigned_url(
                "get_object",
                Params={"Bucket": BUCKET_NAME, "Key": key},
                ExpiresIn=3600,
            ),
        )
