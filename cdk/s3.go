package main

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type S3Resources struct {
	UploadBucket     awss3.CfnBucket
	S3PolicyDocument map[string]interface{}
}

func createUploadBucketResources(scope constructs.Construct, projectStackName string, cfg EnvConfig) *S3Resources {
	bucket := awss3.NewCfnBucket(scope, jsii.String("UploadBucket"), &awss3.CfnBucketProps{
		BucketName: jsii.String(cfg.UploadBucketName),
		PublicAccessBlockConfiguration: &awss3.CfnBucket_PublicAccessBlockConfigurationProperty{
			BlockPublicAcls:       jsii.Bool(true),
			BlockPublicPolicy:     jsii.Bool(true),
			IgnorePublicAcls:      jsii.Bool(true),
			RestrictPublicBuckets: jsii.Bool(true),
		},
		Tags: &[]*awscdk.CfnTag{
			{Key: jsii.String("Project"), Value: jsii.String("geo-journal")},
			{Key: jsii.String("Stack"), Value: jsii.String(cfg.Env)},
		},
	})

	bucketArn := fmt.Sprintf("arn:aws:s3:::%s", cfg.UploadBucketName)
	s3PolicyDocument := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []interface{}{
			map[string]interface{}{
				"Effect": "Allow",
				"Action": []interface{}{
					"s3:PutObject",
					"s3:GetObject",
					"s3:HeadObject",
					"s3:DeleteObject",
					"s3:ListBucket",
				},
				"Resource": []interface{}{
					bucketArn,
					bucketArn + "/*",
				},
			},
		},
	}

	return &S3Resources{
		UploadBucket:     bucket,
		S3PolicyDocument: s3PolicyDocument,
	}
}
