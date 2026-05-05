# Geo Journal API

Base URL depends on environment. All endpoints are prefixed with `/api/v0`.

All responses have the following headers:

```
Content-Type: application/json
Access-Control-Allow-Origin: *
```

---

## Authentication

Protected endpoints require a Cognito ID token in the `Authorization` header:

```
Authorization: Bearer <cognito-id-token>
```

The token is obtained from Cognito after the user signs in. The API validates the token's signature, issuer, audience, and expiry on every request.

---

## Endpoints

### `GET /api/v0/status`

Health check. No authentication required.

**Response `200`**

```json
{
  "text": "ok"
}
```

---

### `GET /api/v0/images`

Returns all images uploaded by the authenticated user.

**Headers**

| Header          | Required | Description          |
|-----------------|----------|----------------------|
| `Authorization` | Yes      | `Bearer <id-token>`  |

**Response `200`**

```json
{
  "images": [
    {
      "name": "sunset.jpg",
      "latitude": "45.5231",
      "longitude": "-122.6765",
      "taken_at": "1714000000",
      "presigned_url": "https://..."
    }
  ]
}
```

| Field           | Type   | Description                                               |
|-----------------|--------|-----------------------------------------------------------|
| `name`          | string | Original filename                                         |
| `latitude`      | string | Decimal latitude string                                   |
| `longitude`     | string | Decimal longitude string                                  |
| `taken_at`      | string | Unix timestamp (seconds) of when the photo was taken      |
| `presigned_url` | string | Temporary S3 URL to fetch the image (expires in 1 hour)   |

`images` is an empty array `[]` if the user has no uploads.

**Response `401`** — missing or invalid token.

---

### `POST /api/v0/presign`

Generates a presigned S3 POST form that allows the client to upload an image directly to S3.

**Headers**

| Header          | Required | Description                               |
|-----------------|----------|-------------------------------------------|
| `Authorization` | Yes      | `Bearer <id-token>`                       |
| `Content-Type`  | Yes      | `application/x-www-form-urlencoded`       |

**Request body (form-encoded)**

| Field       | Required | Description                               |
|-------------|----------|-------------------------------------------|
| `name`      | Yes      | Filename, e.g. `photo.jpg`                |
| `latitude`  | Yes      | Decimal latitude string, e.g. `45.5231`   |
| `longitude` | Yes      | Decimal longitude string, e.g. `-122.676` |
| `taken_at`  | Yes      | Unix timestamp (seconds) as a string      |

**Response `200`**

```json
{
  "url": "https://geojournal-uploads.s3.us-west-2.amazonaws.com",
  "data": {
    "key": "<user-id>/<uuid>",
    "policy": "<base64-encoded-policy>",
    "x-amz-algorithm": "AWS4-HMAC-SHA256",
    "x-amz-credential": "...",
    "x-amz-date": "...",
    "x-amz-signature": "...",
    "x-amz-meta-name": "photo.jpg",
    "x-amz-meta-latitude": "45.5231",
    "x-amz-meta-longitude": "-122.676",
    "x-amz-meta-taken": "1714000000"
  }
}
```

To upload the file, submit a `multipart/form-data` POST directly to `url`, including all fields from `data` as form fields, with the file attached as the `file` field last:

```js
const form = new FormData();
for (const [key, value] of Object.entries(data)) {
  form.append(key, value);
}
form.append("file", file); // must be last

await fetch(url, { method: "POST", body: form });
```

The presigned form expires in **1 hour**.

**Response `400`** — one or more required fields are missing.

**Response `401`** — missing or invalid token.

---

## Error Responses

All error responses return an empty body with the appropriate HTTP status code:

| Status | Meaning                                       |
|--------|-----------------------------------------------|
| `400`  | Bad request — required fields missing         |
| `401`  | Unauthorized — missing or invalid auth token  |
| `404`  | Not found — unknown path                      |
| `405`  | Method not allowed                            |
| `500`  | Internal server error                         |
