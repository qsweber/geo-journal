package main

import (
	"context"
	"encoding/base64"
	"log"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/qsweber/geo-journal/internal/auth"
	"github.com/qsweber/geo-journal/internal/rpc"
	"github.com/qsweber/geo-journal/internal/server"
)

var cognitoVerifier *auth.CognitoVerifier

func init() {
	// Initialize Cognito verifier at startup
	config, err := auth.GetCognitoConfig()
	if err != nil {
		log.Printf("Warning: Failed to load Cognito config: %v", err)
		log.Printf("Authentication will be disabled. Set COGNITO_REGION, COGNITO_USER_POOL_ID, and COGNITO_CLIENT_ID to enable.")
		cognitoVerifier = nil
	} else {
		cognitoVerifier = auth.NewCognitoVerifier(config)
		log.Printf("Cognito authentication enabled for region: %s", config.Region)
	}
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Handle CORS preflight requests
	if request.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization",
			},
		}, nil
	}

	req := rpc.Request{
		Method:      request.HTTPMethod,
		Path:        request.Path,
		Headers:     request.Headers,
		QueryParams: request.QueryStringParameters,
		Form:        parseFormValues(request),
	}

	srv := server.New()

	resp := rpc.Handler(ctx, req, srv, cognitoVerifier)

	return events.APIGatewayProxyResponse{
		StatusCode: resp.StatusCode,
		Body:       resp.Body,
		Headers:    resp.Headers,
	}, nil

}

func parseFormValues(request events.APIGatewayProxyRequest) map[string]string {
	form := map[string]string{}

	contentType := request.Headers["Content-Type"]
	if contentType == "" {
		contentType = request.Headers["content-type"]
	}
	if !strings.Contains(strings.ToLower(contentType), "application/x-www-form-urlencoded") {
		return form
	}

	body := request.Body
	if request.IsBase64Encoded {
		decoded, err := base64.StdEncoding.DecodeString(request.Body)
		if err != nil {
			return form
		}
		body = string(decoded)
	}

	parsed, err := url.ParseQuery(body)
	if err != nil {
		return form
	}

	for key, value := range parsed {
		if len(value) > 0 {
			form[key] = value[0]
		}
	}

	return form
}

func main() {
	lambda.Start(handler)
}
