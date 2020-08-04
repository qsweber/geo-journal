from typing import NamedTuple

from geo_journal.clients.s3 import S3Client


class Clients(NamedTuple):
    s3: S3Client


class ServiceContext(NamedTuple):
    clients: Clients


service_context = ServiceContext(clients=Clients(s3=S3Client(),))
