package main

import (
	"context"
	"log"
	"net/http"

	"github.com/qsweber/geo-journal/internal/auth"
	"github.com/qsweber/geo-journal/internal/rpc"
	"github.com/qsweber/geo-journal/internal/server"
)

type mockVerifier struct{}

func (m mockVerifier) VerifyToken(ctx context.Context, tokenString string) (*auth.Claims, error) {
	return &auth.Claims{}, nil
}

func main() {
	srv := server.New()
	tokenVerifier := mockVerifier{}

	genericHandler := func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()

		headers := make(map[string]string)
		for k, v := range r.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}

		queryParams := make(map[string]string)
		for k, v := range r.URL.Query() {
			if len(v) > 0 {
				queryParams[k] = v[0]
			}
		}

		form := make(map[string]string)
		for k, v := range r.PostForm {
			if len(v) > 0 {
				form[k] = v[0]
			}
		}

		result := rpc.Handler(
			r.Context(),
			rpc.Request{
				Method:      r.Method,
				Path:        r.URL.Path,
				Headers:     headers,
				QueryParams: queryParams,
				Form:        form,
			},
			srv,
			tokenVerifier,
		)
		for key, value := range result.Headers {
			w.Header().Set(key, value)
		}
		w.WriteHeader(result.StatusCode)
		if _, err := w.Write([]byte(result.Body)); err != nil {
			return
		}
	}

	http.HandleFunc("/api/v0/status", genericHandler)
	http.HandleFunc("/api/v0/images", genericHandler)
	http.HandleFunc("/api/v0/presign", genericHandler)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("api server failed: %v", err)
	}
}
