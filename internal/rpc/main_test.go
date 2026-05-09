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

type failingTokenVerifier struct{}

func (v failingTokenVerifier) VerifyToken(ctx context.Context, tokenString string) (*auth.Claims, error) {
	return nil, errors.New("invalid token")
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
