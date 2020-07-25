from typing import NamedTuple

import jwt


class DecodedJwt(NamedTuple):
    id: str
    email: str


KEY_ONE = """
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAlQXrkLhy+a135xnuNFQS
1umMS8BCJSLbrnCdaNKlRGiUDrDuedF0uz6A0hC47tTx6MXCsAupx0cfp44xxRDR
GWSMI/shIPCag5nhASXL+7ZgkUo7hYLeGSJdsS8jlUMYlQIL4V3NqEaP/HUXMboO
f92YjoDK/6E/Z9w7BfANvD3vcAcFNOUDK+oGttnYd1hhRNx/63xYv6ZqqzC55Kyw
zpt5RZaEvHgKlSqZDbk3I5GALEPqevuGs7L+jizkEM5ZwsPUEK4NySAUWqVOBjvT
dAlr9jVA+jPjX2xp0vMGSB1Wj5mNmjYbVCyEaBtrwiKcWMAHZFlMctJJD/I6fIvV
1QIDAQAB
-----END PUBLIC KEY-----
"""

KEY_TWO = """
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAkyeGOuY3DO6XRhBj9f2i
GRwp3IKzpo532X8TD4+PIpeWLus0ssf5kic1pdgcbrK1f4Rz1TdGg3uvvO2Ivptg
OdGKjivggS4e3JN/DlVKP+uFKQwPNmWABA4ihXDY5IACBowVsq9kM2u2AFwxEwHR
z+9kO/4eIMm6duM+ygi7GX378H/C+wjfD6Zk3rNzDKf1YL80FvjffylKKpzL6Puo
YjfIwdN2H9lQKKAf61yHy3FgHEMlKsuDiPd9JlyhqYBBmjraSimKtuCq/+gORwI1
lTDF3Jqgckv4RXW62h1gKslxv3zalJSgOalI6RJe3wiSvMn+oaT2xwajjdKALgaJ
wQIDAQAB
-----END PUBLIC KEY-----
"""

PUBLIC_KEYS = {
    "GUjyT0wt68KHjNBCBjupqs+76IO/77CwLkqnBUna56g=": KEY_ONE,
    "TvPZd3P7mdkrmmRZC7KPdG4kbtql7cpVWIObwivNiOA=": KEY_TWO,
}

CLIENT_ID = "1p5vovpjk10489hipqj7j91ehb"


def decode(jwt_token: str) -> DecodedJwt:
    headers = jwt.get_unverified_header(jwt_token)
    public_key = PUBLIC_KEYS[headers["kid"]]
    decoded = jwt.decode(jwt_token, public_key, algorithms="RS256", audience=CLIENT_ID)

    return DecodedJwt(id=decoded["cognito:username"], email=decoded["email"])
