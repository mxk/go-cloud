package s3x

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mxk/go-cloud/aws/awsx"
)

// ErrCodeBucketNotEmpty indicates a failure to delete a non-empty bucket.
const ErrCodeBucketNotEmpty = "BucketNotEmpty"

// Client is an extended S3 client with additional methods for managing buckets.
type Client struct{ s3.S3 }

// New returns a new extended S3 client.
func New(cfg *aws.Config) *Client { return &Client{*s3.New(*cfg)} }

// DeleteBucket empties and deletes the specified bucket.
func (c *Client) DeleteBucket(name string) error {
	ch := make(chan error)
	go func() { ch <- c.EmptyBucket(name) }()
	in := s3.DeleteBucketInput{Bucket: aws.String(name)}
	_, err := c.DeleteBucketRequest(&in).Send()
	switch awsx.ErrCode(err) {
	case ErrCodeBucketNotEmpty:
		if err = <-ch; awsx.ErrCode(err) == s3.ErrCodeNoSuchBucket {
			return nil
		}
	case s3.ErrCodeNoSuchBucket:
		err = nil
		fallthrough
	default:
		<-ch
		return err
	}
	retries := 3
	for err == nil {
		_, err = c.DeleteBucketRequest(&in).Send()
		switch awsx.ErrCode(err) {
		case ErrCodeBucketNotEmpty:
			if retries--; retries >= 0 {
				err = c.EmptyBucket(name)
				continue
			}
		case s3.ErrCodeNoSuchBucket:
			err = nil
		}
		break
	}
	return err
}

// EmptyBucket deletes all objects from the specified bucket.
func (c *Client) EmptyBucket(name string) error {
	makeObjs := func(del *s3.Delete, n int) []s3.ObjectIdentifier {
		if cap(del.Objects) < n {
			del.Objects = make([]s3.ObjectIdentifier, n)
		} else {
			del.Objects = del.Objects[:n]
		}
		return del.Objects
	}

	// Delete all objects
	listIn := s3.ListObjectsV2Input{
		Bucket:  aws.String(name),
		MaxKeys: aws.Int64(1000),
	}
	delIn := s3.DeleteObjectsInput{
		Bucket: listIn.Bucket,
		Delete: &s3.Delete{Quiet: aws.Bool(true)},
	}
	for {
		out, err := c.ListObjectsV2Request(&listIn).Send()
		if err != nil {
			return err
		}
		if len(out.Contents) > 0 {
			objs := makeObjs(delIn.Delete, len(out.Contents))
			for i := range out.Contents {
				objs[i].Key = out.Contents[i].Key
			}
			if _, err := c.DeleteObjectsRequest(&delIn).Send(); err != nil {
				return err
			}
		}
		if !aws.BoolValue(out.IsTruncated) {
			break
		}
	}

	// Delete all object versions
	verIn := s3.ListObjectVersionsInput{
		Bucket:  listIn.Bucket,
		MaxKeys: listIn.MaxKeys,
	}
	for {
		out, err := c.ListObjectVersionsRequest(&verIn).Send()
		if err != nil {
			return err
		}
		if len(out.Versions) > 0 {
			objs := makeObjs(delIn.Delete, len(out.Versions))
			for i := range out.Versions {
				dst, src := &objs[i], &out.Versions[i]
				dst.Key, dst.VersionId = src.Key, src.VersionId
			}
			if _, err := c.DeleteObjectsRequest(&delIn).Send(); err != nil {
				return err
			}
		}
		if !aws.BoolValue(out.IsTruncated) {
			return nil
		}
	}
}
