from typing import NamedTuple

from PIL import Image, ExifTags  # type: ignore


class Coordinates(NamedTuple):
    latitude: float
    longitude: float


def get_lat_long(path: str) -> Coordinates:
    img = Image.open(path)

    print(img.info.keys())

    exif = {ExifTags.TAGS[k]: v for k, v in img.getexif().items() if k in ExifTags.TAGS}

    gpsinfo = {
        ExifTags.GPSTAGS.get(key, key): value for key, value in exif["GPSInfo"].items()
    }

    latitude = (1 if gpsinfo["GPSLatitudeRef"] == "N" else -1) * sum(
        [
            element[0] / (element[1] * pow(60, i))
            for i, element in enumerate(gpsinfo["GPSLatitude"])
        ]
    )

    longitude = (1 if gpsinfo["GPSLongitudeRef"] == "E" else -1) * sum(
        [
            element[0] / (element[1] * pow(60, i))
            for i, element in enumerate(gpsinfo["GPSLongitude"])
        ]
    )

    return Coordinates(latitude=round(latitude, 6), longitude=round(longitude, 6))
