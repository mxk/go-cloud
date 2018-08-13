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

func TestDelTmpUsers(t *testing.T) {
	w := mock.NewAWS(mock.Ctx)
	c := Client{*iam.New(w.Cfg)}
	r := w.Root().UserRouter()

	path := "/test/"
	require.NoError(t, c.DeleteUsers(path))

	r["a"] = &mock.User{User: iam.User{
		Arn:      arn.String(w.Ctx.New("iam", "user/a")),
		Path:     aws.String("/"),
		UserName: aws.String("a"),
	}}
	r["b"] = &mock.User{
		User: iam.User{
			Arn:      arn.String(w.Ctx.New("iam", "user/b")),
			Path:     aws.String(path),
			UserName: aws.String("b"),
		},
		AttachedPolicies: map[arn.ARN]string{
			w.Ctx.New("iam", "policy/TestPolicy"): "TestPolicy",
		},
		AccessKeys: []*iam.AccessKeyMetadata{{
			AccessKeyId: aws.String(mock.AccessKeyID),
			Status:      iam.StatusTypeActive,
			UserName:    aws.String("b"),
		}},
	}

	require.NoError(t, c.DeleteUsers(path))
	assert.Contains(t, r, "a")
	assert.NotContains(t, r, "b")
}
