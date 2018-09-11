package awsx

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/stretchr/testify/assert"
)

func TestIsCode(t *testing.T) {
	assert.False(t, IsCode(nil, "TestCode"))
	assert.True(t, IsCode(awserr.New("TestCode", "", nil), "TestCode"))
}

func TestIsStatus(t *testing.T) {
	assert.False(t, IsStatus(nil, 404))
	assert.True(t, IsStatus(awserr.NewRequestFailure(nil, 404, ""), 404))
}
