import geo_journal.lib.exif as module


def test_exif():
    path = "./tests/lib/fixtures/img_with_tags.jpeg"
    coordinates = module.get_lat_long(path)
    assert coordinates.longitude == -121.778594
    assert coordinates.latitude == 47.435989
