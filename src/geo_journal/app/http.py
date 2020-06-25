import logging
import json

import typing

from flask import Flask, jsonify, request, Response
from raven import Client  # type: ignore
from raven.contrib.flask import Sentry  # type: ignore
from raven.transport.requests import RequestsHTTPTransport  # type: ignore

from geo_journal.lib.exif import get_lat_long


app = Flask(__name__)
sentry = Sentry(app, client=Client(transport=RequestsHTTPTransport,),)
logger = logging.getLogger(__name__)


@app.route("/api/v0/status", methods=["GET"])
def status() -> Response:
    logger.info("recieved request with args {}".format(json.dumps(request.args)))

    response = jsonify({"text": "ok"})
    response.headers.add("Access-Control-Allow-Origin", "*")

    return typing.cast(Response, response)


@app.route("/api/v0/upload", methods=["POST"])
def upload() -> Response:
    if "file" not in request.files:
        raise Exception("file not uploaded")
    file = request.files["file"]
    if not file or file.filename == "":
        raise Exception("file not uploaded")

    path = "/tmp/foo"
    file.save(path)

    coordinates = get_lat_long(path)

    if coordinates:
        response = jsonify(
            {
                "text": {
                    "latitude": coordinates.latitude,
                    "longitude": coordinates.longitude,
                }
            }
        )
    else:
        response = jsonify({"text": "bad"})

    response.headers.add("Access-Control-Allow-Origin", "*")

    return typing.cast(Response, response)
