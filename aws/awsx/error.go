package awsx

import "github.com/aws/aws-sdk-go-v2/aws/awserr"

// ErrCode returns the error code of an awserr.Error.
func ErrCode(err error) string {
	if e, ok := err.(awserr.Error); ok {
		return e.Code()
	}
	return ""
}

// StatusCode returns the HTTP status code of an awserr.RequestFailure.
func StatusCode(err error) int {
	if e, ok := err.(awserr.RequestFailure); ok {
		return e.StatusCode()
	}
	return 0
}
