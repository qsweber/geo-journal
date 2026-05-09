package rpc

import (
	"context"
	"errors"
	"testing"

	"github.com/qsweber/geo-journal/internal/auth"
	"github.com/qsweber/geo-journal/internal/server"
)

type testServer struct{}

func (s testServer) Status() (server.StatusOutput, error) {
	return server.StatusOutput{Text: "ok"}, nil
}

func (s testServer) Images(userID string) (server.ImagesOutput, error) {
	return server.ImagesOutput{}, nil
}

func (s testServer) Presign(userID string, input server.PresignInput) (server.PresignOutput, error) {
	return server.PresignOutput{}, nil
}

func (s testServer) Delete(userID string, input server.DeleteInput) error {
	return nil
}

type deleteTestServer struct {
	deleteErr error
}

func (s deleteTestServer) Status() (server.StatusOutput, error) {
	return server.StatusOutput{Text: "ok"}, nil
}

func (s deleteTestServer) Images(userID string) (server.ImagesOutput, error) {
	return server.ImagesOutput{}, nil
}

func (s deleteTestServer) Presign(userID string, input server.PresignInput) (server.PresignOutput, error) {
	return server.PresignOutput{}, nil
}

func (s deleteTestServer) Delete(userID string, input server.DeleteInput) error {
	return s.deleteErr
}

type failingTokenVerifier struct{}

func (v failingTokenVerifier) VerifyToken(ctx context.Context, tokenString string) (*auth.Claims, error) {
	return nil, errors.New("invalid token")
}

type passingTokenVerifier struct{}

func (v passingTokenVerifier) VerifyToken(ctx context.Context, tokenString string) (*auth.Claims, error) {
	return &auth.Claims{CognitoUser: "user-123"}, nil
}

func TestStatus(t *testing.T) {
	resp := Handler(
		context.Background(),
		Request{Method: "GET", Path: "/api/v0/status", Headers: map[string]string{}},
		testServer{},
		failingTokenVerifier{},
	)

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	if resp.Body != `{"text":"ok"}` {
		t.Fatalf("unexpected body: %s", resp.Body)
	}
}

func TestImagesBadAuth(t *testing.T) {
	resp := Handler(
		context.Background(),
		Request{
			Method:  "GET",
			Path:    "/api/v0/images",
			Headers: map[string]string{"Authorization": "foo"},
		},
		testServer{},
		failingTokenVerifier{},
	)

	if resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestDeleteImageSuccessNoContent(t *testing.T) {
	resp := Handler(
		context.Background(),
		Request{
			Method:  "DELETE",
			Path:    "/api/v0/images/image-123",
			Headers: map[string]string{"Authorization": "Bearer valid-token"},
		},
		deleteTestServer{},
		passingTokenVerifier{},
	)

	if resp.StatusCode != 204 {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}

	if resp.Body != "" {
		t.Fatalf("expected empty body, got %q", resp.Body)
	}

	if _, ok := resp.Headers["Content-Type"]; ok {
		t.Fatalf("expected no content-type header for 204 response")
	}
}

func TestDeleteImageBadAuth(t *testing.T) {
	resp := Handler(
		context.Background(),
		Request{
			Method:  "DELETE",
			Path:    "/api/v0/images/image-123",
			Headers: map[string]string{"Authorization": "foo"},
		},
		deleteTestServer{},
		failingTokenVerifier{},
	)

	if resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestDeleteImageMissingPathID(t *testing.T) {
	resp := Handler(
		context.Background(),
		Request{
			Method:  "DELETE",
			Path:    "/api/v0/images/",
			Headers: map[string]string{"Authorization": "Bearer valid-token"},
		},
		deleteTestServer{},
		passingTokenVerifier{},
	)

	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestDeleteImageInvalidPath(t *testing.T) {
	resp := Handler(
		context.Background(),
		Request{
			Method:  "DELETE",
			Path:    "/api/v0/images/image-123/extra",
			Headers: map[string]string{"Authorization": "Bearer valid-token"},
		},
		deleteTestServer{},
		passingTokenVerifier{},
	)

	if resp.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestDeleteImageNotFound(t *testing.T) {
	resp := Handler(
		context.Background(),
		Request{
			Method:  "DELETE",
			Path:    "/api/v0/images/image-123",
			Headers: map[string]string{"Authorization": "Bearer valid-token"},
		},
		deleteTestServer{deleteErr: server.ErrImageNotFound},
		passingTokenVerifier{},
	)

	if resp.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestDeleteImageBadRequestFromService(t *testing.T) {
	resp := Handler(
		context.Background(),
		Request{
			Method:  "DELETE",
			Path:    "/api/v0/images/image-123",
			Headers: map[string]string{"Authorization": "Bearer valid-token"},
		},
		deleteTestServer{deleteErr: server.ErrImageIDRequired},
		passingTokenVerifier{},
	)

	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
