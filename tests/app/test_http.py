import json

import pytest

import geo_journal.app.http as module


@pytest.fixture
def client():
    client = module.app.test_client()

    yield client


def test_status(mocker, client):
    result = client.get("/api/v0/status")

    assert result.status_code == 200
    assert json.loads(result.data) == {"text": "ok"}
