package iamx

import (
	"testing"

	"github.com/LuminalHQ/cloudcover/oktapus/mock"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelTmpUsers(t *testing.T) {
	s := mock.NewSession()
	c := *iam.New(s.Config)

	path := "/test/"
	require.NoError(t, DeleteUsers(c, path))

	r := s.OrgsRouter().Account("").UserRouter()
	r["a"] = &mock.User{User: iam.User{
		Arn:      aws.String(mock.UserARN("", "a")),
		Path:     aws.String("/"),
		UserName: aws.String("a"),
	}}
	r["b"] = &mock.User{
		User: iam.User{
			Arn:      aws.String(mock.UserARN("", "b")),
			Path:     aws.String(path),
			UserName: aws.String("b"),
		},
		AttachedPolicies: map[string]string{
			mock.PolicyARN("", "TestPolicy"): "TestPolicy",
		},
		AccessKeys: []*iam.AccessKeyMetadata{{
			AccessKeyId: aws.String(mock.AccessKeyID),
			Status:      iam.StatusTypeActive,
			UserName:    aws.String("b"),
		}},
	}

	require.NoError(t, DeleteUsers(c, path))
	assert.Contains(t, r, "a")
	assert.NotContains(t, r, "b")
}
