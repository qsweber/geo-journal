package main

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewGeoJournalStack(scope constructs.Construct, id string, cfg EnvConfig, props *awscdk.StackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, jsii.String(id), props)

	projectStackName := "geo-journal-" + cfg.Env

	s3Resources := createUploadBucketResources(stack, projectStackName, cfg)

	role := createBaseRole(stack, projectStackName, cfg, s3Resources.S3PolicyDocument)

	sesResources := createSESResources(stack, cfg)

	cognitoResources := createCognitoResources(stack, cfg, sesResources.EmailIdentity)

	apigatewayResources := createAPIGatewayResources(stack, projectStackName, cfg, role, cognitoResources)

	createCustomDomainResources(stack, cfg, apigatewayResources.Gateway, cfg.Env)

	awscdk.NewCfnOutput(stack, jsii.String("LambdaName"), &awscdk.CfnOutputProps{
		Value: apigatewayResources.Function.FunctionName(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("UploadBucketOutput"), &awscdk.CfnOutputProps{
		Value: s3Resources.UploadBucket.Ref(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("CognitoUserPoolId"), &awscdk.CfnOutputProps{
		Value: cognitoResources.UserPool.AttrUserPoolId(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("CognitoUserPoolClientId"), &awscdk.CfnOutputProps{
		Value: cognitoResources.UserPoolClient.AttrClientId(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("InvocationURL"), &awscdk.CfnOutputProps{
		Value: jsii.String(fmt.Sprintf(
			"https://%s.execute-api.%s.amazonaws.com/%s",
			*apigatewayResources.Gateway.AttrRestApiId(),
			cfg.Region,
			cfg.Env,
		)),
	})
	awscdk.NewCfnOutput(stack, jsii.String("CustomDomainURL"), &awscdk.CfnOutputProps{
		Value: jsii.String(fmt.Sprintf("https://%s", cfg.ApiDomainName)),
	})

	return stack
}
