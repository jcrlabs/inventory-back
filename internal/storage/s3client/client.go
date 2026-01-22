package s3client

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	Bucket         string
	PublicBaseURL  string
	S3             *s3.Client
	Presign        *s3.PresignClient
	ForcePathStyle bool
	Endpoint       string
}

type Options struct {
	Endpoint       string
	Region         string
	Bucket         string
	AccessKey      string
	SecretKey      string
	ForcePathStyle bool
	PublicBaseURL  string
}

func New(ctx context.Context, opt Options) (*Client, error) {
	ep := strings.TrimRight(opt.Endpoint, "/")

	customResolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, _ ...interface{}) (aws.Endpoint, error) {
			if service == s3.ServiceID {
				return aws.Endpoint{
					URL:               ep,
					SigningRegion:     opt.Region,
					HostnameImmutable: true,
				}, nil
			}
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		},
	)

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(opt.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(opt.AccessKey, opt.SecretKey, "")),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, err
	}

	s3c := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = opt.ForcePathStyle
	})

	return &Client{
		Bucket:         opt.Bucket,
		PublicBaseURL:  strings.TrimRight(opt.PublicBaseURL, "/"),
		S3:             s3c,
		Presign:        s3.NewPresignClient(s3c),
		ForcePathStyle: opt.ForcePathStyle,
		Endpoint:       ep,
	}, nil
}

func (c *Client) PresignPutObject(ctx context.Context, objectKey, contentType string, expires time.Duration) (uploadURL string, objectURL string, err error) {
	in := &s3.PutObjectInput{
		Bucket:      aws.String(c.Bucket),
		Key:         aws.String(objectKey),
		ContentType: aws.String(contentType),
	}

	ps, err := c.Presign.PresignPutObject(ctx, in, func(po *s3.PresignOptions) {
		po.Expires = expires
	})
	if err != nil {
		return "", "", err
	}

	uploadURL = ps.URL
	objectURL = c.objectURL(objectKey)

	return uploadURL, objectURL, nil
}

func (c *Client) objectURL(objectKey string) string {
	if c.PublicBaseURL != "" {
		return c.PublicBaseURL + "/" + strings.TrimLeft(objectKey, "/")
	}

	// Fallback: construye URL desde endpoint
	// path-style: {endpoint}/{bucket}/{key}
	// virtual-host: {bucket}.{endpoint-host}/{key} (no siempre aplica con MinIO)
	u, err := url.Parse(c.Endpoint)
	if err != nil {
		return ""
	}

	if c.ForcePathStyle {
		u.Path = "/" + c.Bucket + "/" + strings.TrimLeft(objectKey, "/")
		return u.String()
	}

	// virtual-host style (si endpoint lo soporta)
	u.Host = c.Bucket + "." + u.Host
	u.Path = "/" + strings.TrimLeft(objectKey, "/")
	return u.String()
}
