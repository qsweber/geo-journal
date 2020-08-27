from datetime import datetime
from decimal import Decimal
from functools import wraps
import logging
import json
import typing

from jsonschema import validate  # type: ignore

from flask import Flask, jsonify, request, Response, g
from raven import Client  # type: ignore
from raven.contrib.flask import Sentry  # type: ignore
from raven.transport.requests import RequestsHTTPTransport  # type: ignore

from geo_journal.clients.s3 import Image
from geo_journal.lib.jwt import decode
from geo_journal.app.service_context import service_context


app = Flask(__name__)
sentry = Sentry(
    app,
    client=Client(
        transport=RequestsHTTPTransport,
    ),
)
logger = logging.getLogger(__name__)


def schema(
    inputSchema: typing.Dict[str, typing.Any],
    outputSchema: typing.Dict[str, typing.Any],
) -> typing.Callable[[typing.Callable[..., Response]], typing.Callable[..., Response]]:
    def wrapper(
        func: typing.Callable[..., Response],
    ) -> typing.Callable[..., Response]:
        @wraps(func)
        def what_gets_called(*args: typing.Any, **kwargs: typing.Any) -> Response:
            if request.method == "GET":
                validate(instance=request.args, schema=inputSchema)
            elif request.method == "POST":
                validate(instance=request.form, schema=inputSchema)
            else:
                raise Exception("unexpected method {}".format(request.method))

            result = func(*args, **kwargs)

            validate(instance=json.loads(result.data), schema=outputSchema)

            return result

        return what_gets_called

    return wrapper


def authenticate(
    func: typing.Callable[..., Response]
) -> typing.Callable[..., Response]:
    @wraps(func)
    def what_gets_called(*args: typing.Any, **kwargs: typing.Any) -> Response:
        try:
            jwt = decode(request.headers["Authorization"])
        except Exception:
            return Response(status=401)

        g.jwt = jwt
        return func(*args, **kwargs)

    return what_gets_called


@app.route("/api/v0/status", methods=["GET"])
def status() -> Response:
    logger.info("recieved request {}".format(json.dumps(request.json)))

    response = jsonify({"text": "ok"})
    response.headers.add("Access-Control-Allow-Origin", "*")

    return typing.cast(Response, response)


@app.route("/api/v0/images", methods=["GET"])
@authenticate
@schema(
    {},
    {
        "type": "object",
        "properties": {
            "images": {
                "type": "array",
                "items": {
                    "type": "object",
                    "properties": {
                        "latitude": {"type": "string"},
                        "longitude": {"type": "string"},
                        "taken_at": {"type": "string"},
                        "name": {"type": "string"},
                    },
                },
            }
        },
    },
)
def images() -> Response:
    logger.info("recieve request for user {}".format(g.jwt.id))
    response = jsonify({"images": service_context.clients.s3.get_images(g.jwt.id)})
    response.headers.add("Access-Control-Allow-Origin", "*")

    return typing.cast(Response, response)


@app.route("/api/v0/presign", methods=["POST"])
@authenticate
@schema(
    {
        "type": "object",
        "properties": {
            "latitude": {"type": "string"},
            "longitude": {"type": "string"},
            "taken_at": {"type": "string"},
            "name": {"type": "string"},
        },
        "required": ["latitude", "longitude", "taken_at", "name"],
    },
    {
        "type": "object",
        "properties": {
            "data": {},
            "url": {"type": "string"},
        },
        "required": ["data", "url"],
    },
)
def presign() -> Response:
    url, data = service_context.clients.s3.create_presigned_post(
        g.jwt.id,
        Image(
            name=request.form["name"],
            taken_at=datetime.fromtimestamp(int(request.form["taken_at"])),
            latitude=Decimal(request.form["latitude"]),
            longitude=Decimal(request.form["longitude"]),
        ),
    )
    logger.info("recieve request for user {}".format(g.jwt.id))
    response = jsonify({"url": url, "data": data})
    response.headers.add("Access-Control-Allow-Origin", "*")

    return typing.cast(Response, response)
