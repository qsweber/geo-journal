package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
)

// EnvConfig holds the per-stack (dev/prod) settings, plus the pre-existing physical resource
// names/values that need to be preserved so that `cdk import` can adopt the already-deployed
// AWS resources without replacing them.
type EnvConfig struct {
	Env     string
	Account string
	Region  string
	ZoneID  string

	// The frontend's domain name (owned by geo-journal-web). Cognito and SES resources were
	// originally created there and are being relocated here; their names/identities still
	// reference this domain for fidelity with the resources being imported.
	FrontendDomainName string

	// Physical names of already-deployed resources (with their pre-existing random suffix).
	// These must match exactly for `cdk import` to adopt the existing resources.
	TaskExecRoleName string
	UploadBucketName string
	LambdaName       string

	// Already-existing Cognito/SES resources being relocated from geo-journal-web's Pulumi
	// state into this stack.
	CognitoUserPoolID       string
	CognitoUserPoolClientID string
	SESVerificationToken    string

	// Custom domain for the API Gateway (e.g. "dev-geojournal-api.quinnweber.com").
	ApiDomainName string
}

func main() {
	app := awscdk.NewApp(nil)

	account := "120356305272"

	NewGeoJournalStack(app, "geo-journal-dev", EnvConfig{
		Env:                     "dev",
		Account:                 account,
		Region:                  "us-west-2",
		ZoneID:                  "Z07290422T6PHN4TKVNW0",
		FrontendDomainName:      "dev-geojournal.quinnweber.com",
		TaskExecRoleName:        "geo-journal-dev-task-exec-role-49dcca9",
		UploadBucketName:        "geo-journal-dev-uploads",
		LambdaName:              "geo-journal-dev-function-190659b",
		CognitoUserPoolID:       "us-west-2_cZd2FyC4g",
		CognitoUserPoolClientID: "1lg8ab88foqkg9upe0v3t7krs6",
		SESVerificationToken:    "P+vE8JXBsoTf9G5Swd9upRPr5KdHZZRnIT5Nyz4CHmQ=",
		ApiDomainName:           "dev-geojournal-api.quinnweber.com",
	}, &awscdk.StackProps{
		Env: &awscdk.Environment{
			Account: jsii.String(account),
			Region:  jsii.String("us-west-2"),
		},
	})

	NewGeoJournalStack(app, "geo-journal-prod", EnvConfig{
		Env:                     "prod",
		Account:                 account,
		Region:                  "us-west-2",
		ZoneID:                  "Z07290422T6PHN4TKVNW0",
		FrontendDomainName:      "geojournal.quinnweber.com",
		TaskExecRoleName:        "geo-journal-prod-task-exec-role-6fde2e6",
		UploadBucketName:        "geo-journal-prod-uploads",
		LambdaName:              "geo-journal-prod-function-984dc69",
		CognitoUserPoolID:       "us-west-2_E17RlloBP",
		CognitoUserPoolClientID: "4chj22ktu0op51bs8obgrn46p0",
		SESVerificationToken:    "uKKn/6jG9xnXTSbgVuYiPV8qlbJwQCt091Ko+kH43U0=",
		ApiDomainName:           "geojournal-api.quinnweber.com",
	}, &awscdk.StackProps{
		Env: &awscdk.Environment{
			Account: jsii.String(account),
			Region:  jsii.String("us-west-2"),
		},
	})

	app.Synth(nil)
}
