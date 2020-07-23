from typing import NamedTuple

from geo_journal.clients.cognito import CognitoClient


class Clients(NamedTuple):
    cognito: CognitoClient


class ServiceContext(NamedTuple):
    clients: Clients


service_context = ServiceContext(clients=Clients(cognito=CognitoClient(),))
