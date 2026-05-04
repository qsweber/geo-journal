package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/qsweber/geo-journal/internal/auth"
	"github.com/qsweber/geo-journal/internal/server"
)

type Request struct {
	Method      string
	Path        string
	Headers     map[string]string
	QueryParams map[string]string
	Form        map[string]string
}

type Response struct {
	StatusCode int
	Body       string
	Headers    map[string]string
}

type TokenVerifier interface {
	VerifyToken(ctx context.Context, tokenString string) (*auth.Claims, error)
}

func authenticateRequest(ctx context.Context, req Request, tokenVerifier TokenVerifier) (*auth.Claims, error) {
	if tokenVerifier == nil {
		return nil, errors.New("authentication is not configured")
	}
	authHeader, ok := req.Headers["Authorization"]
	if !ok {
		authHeader, ok = req.Headers["authorization"]
		if !ok {
			return nil, errors.New("authorization header is missing")
		}
	}
	token, err := auth.ExtractBearerToken(authHeader)
	if err != nil {
		return nil, errors.New("failed to extract authorization token")
	}
	claims, err := tokenVerifier.VerifyToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}
	return claims, nil
}

func corsHeaders() map[string]string {
	return map[string]string{
		"Access-Control-Allow-Origin": "*",
		"Content-Type":                "application/json",
	}
}

func jsonResponse(statusCode int, payload any) Response {
	body, err := json.Marshal(payload)
	if err != nil {
		return errorResponse(500, err)
	}
	return Response{StatusCode: statusCode, Body: string(body), Headers: corsHeaders()}
}

func errorResponse(statusCode int, err error) Response {
	log.Printf("error %d: %v", statusCode, err)
	body, _ := json.Marshal(map[string]string{"error": err.Error()})
	return Response{StatusCode: statusCode, Body: string(body), Headers: corsHeaders()}
}

func Handler(ctx context.Context, req Request, srv server.Server, tokenVerifier TokenVerifier) Response {
	method := strings.ToUpper(req.Method)
	log.Printf("%s %s", method, req.Path)

	switch req.Path {
	case "/api/v0/status":
		if method != "GET" {
			return errorResponse(405, errors.New("method not allowed"))
		}

		result, err := srv.Status()
		if err != nil {
			return errorResponse(500, err)
		}

		return jsonResponse(200, result)
	case "/api/v0/images":
		if method != "GET" {
			return errorResponse(405, errors.New("method not allowed"))
		}

		claims, err := authenticateRequest(ctx, req, tokenVerifier)
		if err != nil {
			return errorResponse(401, err)
		}

		output, err := srv.Images(claims.CognitoUser)
		if err != nil {
			return errorResponse(500, err)
		}

		return jsonResponse(200, output)
	case "/api/v0/presign":
		if method != "POST" {
			return errorResponse(405, errors.New("method not allowed"))
		}

		claims, err := authenticateRequest(ctx, req, tokenVerifier)
		if err != nil {
			return errorResponse(401, err)
		}

		input := server.PresignInput{
			Latitude:  req.Form["latitude"],
			Longitude: req.Form["longitude"],
			TakenAt:   req.Form["taken_at"],
			Name:      req.Form["name"],
		}
		if input.Latitude == "" || input.Longitude == "" || input.TakenAt == "" || input.Name == "" {
			return errorResponse(400, errors.New("latitude, longitude, taken_at and name are required"))
		}

		output, err := srv.Presign(claims.CognitoUser, input)
		if err != nil {
			return errorResponse(500, err)
		}

		return jsonResponse(200, output)

	default:
		return errorResponse(404, errors.New("not found"))
	}
}
