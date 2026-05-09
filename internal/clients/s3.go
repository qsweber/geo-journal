package clients

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

type ImageMetadata struct {
	Name      string
	TakenAt   time.Time
	Latitude  string
	Longitude string
}

type Image struct {
	ID           string
	Metadata     ImageMetadata
	PresignedURL string
}

type S3Client interface {
	CreatePresignedPost(userID string, imageMetadata ImageMetadata) (string, map[string]string, error)
	GetImages(userID string) ([]Image, error)
	DeleteImage(userID, imageID string) error
}

var ErrImageNotFound = errors.New("image not found")

type s3Client struct {
	bucket string
	region string
	svc    *s3.S3
	creds  *credentials.Credentials
}

func NewS3Client() (S3Client, error) {
	bucket := os.Getenv("GEO_JOURNAL_UPLOAD_BUCKET")
	if bucket == "" {
		return nil, fmt.Errorf("GEO_JOURNAL_UPLOAD_BUCKET environment variable is required")
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}
	if region == "" {
		return nil, fmt.Errorf("AWS_REGION or AWS_DEFAULT_REGION environment variable is required")
	}

	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return nil, fmt.Errorf("create aws session: %w", err)
	}

	return &s3Client{
		bucket: bucket,
		region: region,
		svc:    s3.New(sess),
		creds:  sess.Config.Credentials,
	}, nil
}

func (c *s3Client) CreatePresignedPost(userID string, imageMetadata ImageMetadata) (string, map[string]string, error) {
	credsValue, err := c.creds.Get()
	if err != nil {
		return "", nil, fmt.Errorf("load aws credentials: %w", err)
	}

	key := path.Join(userID, uuid.NewString())
	timestamp := time.Now().UTC()
	expiresAt := timestamp.Add(time.Hour)
	dateStamp := timestamp.Format("20060102")
	credential := fmt.Sprintf("%s/%s/%s/s3/aws4_request", credsValue.AccessKeyID, dateStamp, c.region)
	amzDate := timestamp.Format("20060102T150405Z")

	metadata := map[string]string{
		"x-amz-meta-name":      imageMetadata.Name,
		"x-amz-meta-latitude":  imageMetadata.Latitude,
		"x-amz-meta-longitude": imageMetadata.Longitude,
		"x-amz-meta-taken":     fmt.Sprintf("%d", imageMetadata.TakenAt.Unix()),
	}

	conditions := []interface{}{
		map[string]string{"bucket": c.bucket},
		map[string]string{"key": key},
		map[string]string{"x-amz-algorithm": "AWS4-HMAC-SHA256"},
		map[string]string{"x-amz-credential": credential},
		map[string]string{"x-amz-date": amzDate},
	}

	for field, value := range metadata {
		conditions = append(conditions, map[string]string{field: value})
	}

	if credsValue.SessionToken != "" {
		conditions = append(conditions, map[string]string{"x-amz-security-token": credsValue.SessionToken})
	}

	policy := map[string]interface{}{
		"expiration": expiresAt.Format(time.RFC3339),
		"conditions": conditions,
	}

	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return "", nil, fmt.Errorf("marshal post policy: %w", err)
	}

	policyBase64 := base64.StdEncoding.EncodeToString(policyJSON)
	signature := signPostPolicy(policyBase64, credsValue.SecretAccessKey, dateStamp, c.region)

	fields := map[string]string{
		"key":              key,
		"policy":           policyBase64,
		"x-amz-algorithm":  "AWS4-HMAC-SHA256",
		"x-amz-credential": credential,
		"x-amz-date":       amzDate,
		"x-amz-signature":  signature,
	}

	for field, value := range metadata {
		fields[field] = value
	}

	if credsValue.SessionToken != "" {
		fields["x-amz-security-token"] = credsValue.SessionToken
	}

	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com", c.bucket, c.region)
	return url, fields, nil
}

func signPostPolicy(policyBase64, secretKey, dateStamp, region string) string {
	dateKey := hmacSHA256([]byte("AWS4"+secretKey), dateStamp)
	regionKey := hmacSHA256(dateKey, region)
	serviceKey := hmacSHA256(regionKey, "s3")
	signingKey := hmacSHA256(serviceKey, "aws4_request")
	signature := hmacSHA256(signingKey, policyBase64)
	return hex.EncodeToString(signature)
}

func hmacSHA256(key []byte, message string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return h.Sum(nil)
}

func (c *s3Client) GetImages(userID string) ([]Image, error) {
	result, err := c.svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucket),
		Prefix: aws.String(userID),
	})
	if err != nil {
		return nil, fmt.Errorf("list objects: %w", err)
	}

	if len(result.Contents) == 0 {
		return []Image{}, nil
	}

	images := make([]Image, 0, len(result.Contents))
	for _, object := range result.Contents {
		if object.Key == nil {
			continue
		}
		image, err := c.getImage(*object.Key)
		if err != nil {
			return nil, err
		}
		images = append(images, image)
	}

	return images, nil
}

func (c *s3Client) getImage(key string) (Image, error) {
	head, err := c.svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return Image{}, fmt.Errorf("head object: %w", err)
	}

	req, _ := c.svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	presignedURL, err := req.Presign(time.Hour)
	if err != nil {
		return Image{}, fmt.Errorf("presign get object: %w", err)
	}

	// Go's net/http canonicalizes response header keys to Title-Case, so the
	// AWS SDK returns metadata keys like "Latitude" instead of "latitude".
	// Normalize all keys to lowercase for consistent lookup.
	rawMetadata := make(map[string]*string, len(head.Metadata))
	for k, v := range head.Metadata {
		rawMetadata[strings.ToLower(k)] = v
	}

	takenAt := time.Unix(0, 0).UTC()
	if rawMetadata["taken"] != nil {
		takenUnix, parseErr := strconv.ParseInt(*rawMetadata["taken"], 10, 64)
		if parseErr == nil {
			takenAt = time.Unix(takenUnix, 0).UTC()
		}
	}

	metadata := ImageMetadata{
		Name:      getMetadataValue(rawMetadata, "name"),
		TakenAt:   takenAt,
		Latitude:  getMetadataValue(rawMetadata, "latitude"),
		Longitude: getMetadataValue(rawMetadata, "longitude"),
	}

	return Image{ID: path.Base(key), Metadata: metadata, PresignedURL: presignedURL}, nil
}

func getMetadataValue(metadata map[string]*string, key string) string {
	value, ok := metadata[key]
	if !ok || value == nil {
		return ""
	}
	return *value
}

func (c *s3Client) DeleteImage(userID, imageID string) error {
	key := path.Join(userID, imageID)

	_, err := c.svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var requestErr awserr.RequestFailure
		if errors.As(err, &requestErr) && requestErr.StatusCode() == 404 {
			return ErrImageNotFound
		}
		return fmt.Errorf("head object: %w", err)
	}

	_, err = c.svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete object: %w", err)
	}
	return nil
}
