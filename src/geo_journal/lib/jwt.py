from typing import NamedTuple

import jwt


class DecodedJwt(NamedTuple):
    id: str
    email: str


def decode(jwtToken: str) -> DecodedJwt:
    parsed = jwt.decode(jwtToken, algorithms="RS256", verify=False)

    return DecodedJwt(id=parsed["cognito:username"], email=parsed["email"])
