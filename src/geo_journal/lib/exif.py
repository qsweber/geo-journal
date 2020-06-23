from typing import Tuple

from PIL import Image, ExifTags


def get_lat_long(path: str) -> Tuple[int, int]:
    img = Image.open(path)

    exif = {
        ExifTags.TAGS[k]: v
        for k, v in img.info["parsed_exif"].items()
        if k in ExifTags.TAGS
    }

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

    return latitude, longitude
