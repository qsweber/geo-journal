# Geo Journal API

This repository contains:

- A local HTTP API entrypoint at cmd/api/main.go
- An AWS Lambda entrypoint at cmd/lambda/main.go
- Shared business logic in internal/server
- Shared request routing in internal/rpc
- AWS CDK (Go) infrastructure code in cdk

## Prerequisites

1. Go 1.24+
2. (Optional, for deployment) Node.js 20+, the AWS CDK CLI, and AWS credentials

## Run The API Locally

Start the local API server from the repository root:

```bash
go run ./cmd/api/main.go
```

The server listens on port 8080.

### Endpoints

1. Service status:

```bash
curl -i http://localhost:8080/api/v0/status
```

2. List images without auth (expected 401):

```bash
curl -i http://localhost:8080/api/v0/images
```

3. List images with JWT auth (expected 200 with a valid token):

```bash
curl -i -H "Authorization: <jwt-token>" http://localhost:8080/api/v0/images
```

4. Create an upload presign form:

```bash
curl -i -X POST \
	-H "Authorization: <jwt-token>" \
	-H "Content-Type: application/x-www-form-urlencoded" \
	--data "latitude=45.12&longitude=-122.64&taken_at=1714000000&name=my-photo.jpg" \
	http://localhost:8080/api/v0/presign
```

Expected body for successful status request:

```json
{"text":"ok","timestamp":"2026-01-01T00:00:00Z"}
```

## Local Auth Behavior

For local development, cmd/api/main.go uses a mock token verifier.

- Protected routes still require an Authorization header in Bearer format.
- The token value is not validated against Cognito in local mode.

## S3 Behavior

- Upload presign responses include form-style `url` and `data` fields to preserve old client behavior.
- Images are listed from the per-user S3 prefix and returned with string `latitude`, `longitude`, and `taken_at` fields.
- Default upload bucket is `geojournal-uploads` and can be overridden with `GEO_JOURNAL_UPLOAD_BUCKET`.

## Lambda Auth Behavior

The Lambda entrypoint uses Cognito verification when these environment variables are set:

- COGNITO_REGION
- COGNITO_USER_POOL_ID
- COGNITO_CLIENT_ID

If configuration is missing, auth verification is disabled in Lambda initialization.

## Build Lambda Artifact

Use the included make target:

```bash
make build-lambda
```

This produces:

- bootstrap
- handler.zip

## Deploy With CDK

The `cdk/` directory is a standalone Go module containing a CDK app that defines two stacks,
`geo-journal-dev` and `geo-journal-prod`, each targeting `us-west-2`. It also owns the Cognito
user pool/client and SES domain identity used for authentication - the paired
[geo-journal-web](https://github.com/qsweber/geo-journal-web) frontend just consumes the
resulting pool ID/client ID as config.

1. Install the pinned CDK CLI (only needed once):

```bash
cd cdk && npm install
```

2. One-time per AWS account/region, bootstrap the CDK toolkit:

```bash
npx cdk bootstrap aws://<account-id>/us-west-2
```

3. Build the Lambda artifact (from the repo root) so the CDK app has something to package:

```bash
make build-lambda
```

4. Review and deploy a stack:

```bash
cd cdk
npx cdk diff geo-journal-dev        # or geo-journal-prod
npx cdk deploy geo-journal-dev
```

5. Tear down when done:

```bash
npx cdk destroy geo-journal-dev
```

Note: the Cognito user pool/client and SES domain identity/Route53 records are deployed with a
`RETAIN` removal policy, so `cdk destroy` will leave them in place.
