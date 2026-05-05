# Geo Journal API

This repository contains:

- A local HTTP API entrypoint at cmd/api/main.go
- An AWS Lambda entrypoint at cmd/lambda/main.go
- Shared business logic in internal/server
- Shared request routing in internal/rpc
- Pulumi infrastructure code in infrastructure

## Prerequisites

1. Go 1.24+
2. (Optional, for deployment) Pulumi CLI and AWS credentials

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
{"text":"ok"}
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

## Deploy With Pulumi

1. Initialize/select a stack:

```bash
pulumi stack init dev
```

2. Set AWS region:

```bash
pulumi config set aws:region us-west-2
```

3. Deploy:

```bash
pulumi up
```

4. Tear down when done:

```bash
pulumi destroy --yes
pulumi stack rm --yes
```
