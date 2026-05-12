package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func setupUploadBucket(ctx *pulumi.Context, projectStackName string, roleName pulumi.StringInput) (*s3.BucketV2, *iam.RolePolicy, string, error) {
	bucketName := projectStackName + "-uploads"

	uploadBucket, err := s3.NewBucketV2(ctx, bucketName, &s3.BucketV2Args{
		Bucket: pulumi.String(bucketName),
		Tags: pulumi.StringMap{
			"Project": pulumi.String(ctx.Project()),
			"Stack":   pulumi.String(ctx.Stack()),
		},
	})
	if err != nil {
		return nil, nil, "", err
	}

	_, err = s3.NewBucketPublicAccessBlock(ctx, bucketName+"-public-access-block", &s3.BucketPublicAccessBlockArgs{
		Bucket:                uploadBucket.ID(),
		BlockPublicAcls:       pulumi.Bool(true),
		BlockPublicPolicy:     pulumi.Bool(true),
		IgnorePublicAcls:      pulumi.Bool(true),
		RestrictPublicBuckets: pulumi.Bool(true),
	})
	if err != nil {
		return nil, nil, "", err
	}

	s3Policy, err := iam.NewRolePolicy(ctx, projectStackName+"-lambda-s3-policy", &iam.RolePolicyArgs{
		Role: roleName,
		Policy: uploadBucket.Arn.ApplyT(func(arn string) string {
			return fmt.Sprintf(`{
	"Version": "2012-10-17",
	"Statement": [{
		"Effect": "Allow",
		"Action": [
			"s3:PutObject",
			"s3:GetObject",
			"s3:HeadObject",
			"s3:DeleteObject",
			"s3:ListBucket"
		],
		"Resource": [
			"%s",
			"%s/*"
		]
	}]
}`, arn, arn)
		}).(pulumi.StringOutput),
	})
	if err != nil {
		return nil, nil, "", err
	}

	return uploadBucket, s3Policy, bucketName, nil
}
