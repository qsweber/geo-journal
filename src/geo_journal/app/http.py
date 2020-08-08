import logging
import json
from decimal import Decimal

import typing

from flask import Flask, jsonify, request, Response, Request
from raven import Client  # type: ignore
from raven.contrib.flask import Sentry  # type: ignore
from raven.transport.requests import RequestsHTTPTransport  # type: ignore

from geo_journal.lib.jwt import decode, DecodedJwt

app = Flask(__name__)
sentry = Sentry(app, client=Client(transport=RequestsHTTPTransport,),)
logger = logging.getLogger(__name__)

INPUT = typing.TypeVar("INPUT")
OUTPUT = typing.TypeVar("OUTPUT")

JSONType = typing.Dict[str, typing.Union[str, int, Decimal]]


def typedRoute(
    input_converter: typing.Callable[[Request], INPUT],
    output_converter: typing.Callable[[OUTPUT], JSONType],
) -> typing.Callable[[typing.Callable[[INPUT], OUTPUT]], typing.Callable[[], Response]]:
    def wrapper(
        func: typing.Callable[[INPUT], OUTPUT]
    ) -> typing.Callable[[], Response]:
        def what_gets_called() -> Response:
            logger.info("recieved request {}".format(json.dumps(request.json)))
            inpt = input_converter(request)
            output = func(inpt)
            output_json = output_converter(output)

            response = jsonify(output_json)
            response.headers.add("Access-Control-Allow-Origin", "*")

            return typing.cast(Response, response)

        return what_gets_called

    return wrapper


@app.route("/api/v0/status", methods=["GET"])
def status() -> Response:
    logger.info("recieved request {}".format(json.dumps(request.json)))

    response = jsonify({"text": "ok"})
    response.headers.add("Access-Control-Allow-Origin", "*")

    return typing.cast(Response, response)


class ImagesRequest(typing.NamedTuple):
    user_id: str


class Image(typing.NamedTuple):
    latitude: Decimal
    longitude: Decimal
    name: str


class ImagesResult(typing.NamedTuple):
    images: typing.List[Image]


def images_request_converter(r: Request, jwt: DecodedJwt) -> ImagesRequest:
    return ImagesRequest(user_id=jwt.id)


def images_result_converter(images_result: ImagesResult) -> typing.Any:
    return {
        "images": [
            {
                "latitude": image.latitude,
                "longitude": image.longitude,
                "name": image.name,
            }
            for image in images_result.images
        ]
    }


@app.route("/api/v0/images", methods=["GET"])
@typedRouteAuth(images_request_converter, images_result_converter)
def user(images_request: ImagesRequest) -> ImagesResult:
    return ImagesResult(images=[])


class Foo(typing.NamedTuple):
    foo: str


class Bar(typing.NamedTuple):
    bar: str


def foo_converter(r: Request) -> Foo:
    return Foo(foo=r.args["foo"])


def bar_converter(b: Bar) -> JSONType:
    return {"bar": b.bar}


@app.route("/api/v0/status_typed", methods=["GET"])
@typedRoute(foo_converter, bar_converter)
def status_typed(foo: Foo) -> Bar:
    return Bar(bar=foo.foo)
