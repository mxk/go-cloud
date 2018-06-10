package region

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPartitions(t *testing.T) {
	parts := Partitions()
	assert.Contains(t, parts, "aws")
	assert.Contains(t, parts, "aws-us-gov")
}

func TestPartition(t *testing.T) {
	assert.Equal(t, "aws", Partition("us-east-1"))
	assert.Equal(t, "aws", Partition("aws-global"))
	assert.Equal(t, "aws-us-gov", Partition("aws-us-gov-global"))
}

func TestRelated(t *testing.T) {
	r := Related("us-east-1")
	assert.Contains(t, r, "us-east-1")
	assert.Contains(t, r, "us-west-1")
	assert.NotContains(t, r, "us-gov-west-1")
	assert.Equal(t, r, Related("aws"))
	assert.Nil(t, Related(""))
}

func TestSupports(t *testing.T) {
	assert.True(t, Supports("us-east-1", "ec2"))
	assert.False(t, Supports("us-east-1", "iam"))
	assert.True(t, Supports("aws-global", "iam"))
	assert.True(t, Supports("aws-global", "ec2metadata"))
}
