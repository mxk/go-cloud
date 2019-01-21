package iamx

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/mxk/cloudcover/oktapus/mock"
	"github.com/mxk/cloudcover/x/arn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelTmpRoles(t *testing.T) {
	w := mock.NewAWS(mock.Ctx)
	c := Client{*iam.New(w.Cfg)}
	r := w.Root().RoleRouter()

	path := "/temp/"
	require.NoError(t, c.DeleteRoles(path))

	r["a"] = &mock.Role{Role: iam.Role{
		Arn:      arn.String(w.Ctx.New("iam", "role/a")),
		Path:     aws.String("/"),
		RoleName: aws.String("a"),
	}}
	r["b"] = &mock.Role{
		Role: iam.Role{
			Arn:      arn.String(w.Ctx.New("iam", "role/b")),
			Path:     aws.String(path),
			RoleName: aws.String("b"),
		},
		AttachedPolicies: map[arn.ARN]string{
			w.Ctx.New("iam", "policy/AttachedPolicy1"): "AttachedPolicy1",
			w.Ctx.New("iam", "policy/AttachedPolicy2"): "AttachedPolicy2",
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
