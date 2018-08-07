package iamx

import (
	"testing"

	"github.com/LuminalHQ/cloudcover/oktapus/mock"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelTmpRoles(t *testing.T) {
	s := mock.NewSession()
	c := Client{*iam.New(s.Config)}

	path := "/temp/"
	require.NoError(t, c.DeleteRoles(path))

	r := s.OrgsRouter().Account("").RoleRouter()
	r["a"] = &mock.Role{Role: iam.Role{
		Arn:      aws.String(mock.RoleARN("", "a")),
		Path:     aws.String("/"),
		RoleName: aws.String("a"),
	}}
	r["b"] = &mock.Role{
		Role: iam.Role{
			Arn:      aws.String(mock.RoleARN("", "b")),
			Path:     aws.String(path),
			RoleName: aws.String("b"),
		},
		AttachedPolicies: map[string]string{
			mock.PolicyARN("", "AttachedPolicy1"): "AttachedPolicy1",
			mock.PolicyARN("", "AttachedPolicy2"): "AttachedPolicy2",
		},
		InlinePolicies: map[string]string{
			"InlinePolicy1": "{}",
			"InlinePolicy2": "{}",
		},
	}

	require.NoError(t, c.DeleteRoles(path))
	assert.Contains(t, r, "a")
	assert.NotContains(t, r, "b")
}
