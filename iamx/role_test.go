package iamx

import (
	"testing"

	"github.com/LuminalHQ/cloudcover/oktapus/mock"
	"github.com/LuminalHQ/cloudcover/x/arn"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelTmpRoles(t *testing.T) {
	w := mock.NewAWS("")
	c := Client{*iam.New(w.Cfg)}

	path := "/temp/"
	require.NoError(t, c.DeleteRoles(path))

	arnCtx := w.Ctx()
	r := w.AccountRouter().Get("").RoleRouter()
	r["a"] = &mock.Role{Role: iam.Role{
		Arn:      arn.String(arnCtx.New("iam", "role/a")),
		Path:     aws.String("/"),
		RoleName: aws.String("a"),
	}}
	r["b"] = &mock.Role{
		Role: iam.Role{
			Arn:      arn.String(arnCtx.New("iam", "role/b")),
			Path:     aws.String(path),
			RoleName: aws.String("b"),
		},
		AttachedPolicies: map[arn.ARN]string{
			arnCtx.New("iam", "policy/AttachedPolicy1"): "AttachedPolicy1",
			arnCtx.New("iam", "policy/AttachedPolicy2"): "AttachedPolicy2",
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
