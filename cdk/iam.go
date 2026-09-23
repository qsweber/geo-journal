package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

var assumeRolePolicyDocument = map[string]interface{}{
	"Version": "2012-10-17",
	"Statement": []interface{}{
		map[string]interface{}{
			"Sid":    "",
			"Effect": "Allow",
			"Principal": map[string]interface{}{
				"Service": "lambda.amazonaws.com",
			},
			"Action": "sts:AssumeRole",
		},
	},
}

var logPolicyDocument = map[string]interface{}{
	"Version": "2012-10-17",
	"Statement": []interface{}{
		map[string]interface{}{
			"Effect": "Allow",
			"Action": []interface{}{
				"logs:CreateLogGroup",
				"logs:CreateLogStream",
				"logs:PutLogEvents",
			},
			"Resource": "arn:aws:logs:*:*:*",
		},
	},
}

// createBaseRole builds the Lambda execution role. CloudFormation has no standalone
// "attached role policy" resource for inline policies - they are just entries in the Role's
// own Policies property - so both inline policies are embedded directly on the CfnRole here.
func createBaseRole(scope constructs.Construct, projectStackName string, cfg EnvConfig, s3PolicyDocument map[string]interface{}) awsiam.CfnRole {
	role := awsiam.NewCfnRole(scope, jsii.String("TaskExecRole"), &awsiam.CfnRoleProps{
		RoleName:                 jsii.String(cfg.TaskExecRoleName),
		AssumeRolePolicyDocument: assumeRolePolicyDocument,
		Policies: []interface{}{
			&awsiam.CfnRole_PolicyProperty{
				PolicyName:     jsii.String(projectStackName + "-lambda-log-policy"),
				PolicyDocument: logPolicyDocument,
			},
			&awsiam.CfnRole_PolicyProperty{
				PolicyName:     jsii.String(projectStackName + "-lambda-s3-policy"),
				PolicyDocument: s3PolicyDocument,
			},
		},
	})

	return role
}
