package awsx

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/stretchr/testify/assert"
)

func TestErrCode(t *testing.T) {
	assert.Equal(t, "", ErrCode(nil))
	assert.Equal(t, "TestCode", ErrCode(awserr.New("TestCode", "", nil)))
}

func TestStatusCode(t *testing.T) {
	assert.Equal(t, 0, StatusCode(nil))
	assert.Equal(t, 404, StatusCode(awserr.NewRequestFailure(nil, 404, "")))
}
