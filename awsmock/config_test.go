package awsmock

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	wantOut := new(sts.GetCallerIdentityOutput)
	cfg := Config(func(q *aws.Request) { q.Data = wantOut })
	out, err := sts.New(cfg).GetCallerIdentityRequest(nil).Send()
	require.True(t, out == wantOut)
	require.NoError(t, err)

	wantErr := errors.New("error")
	cfg = Config(func(q *aws.Request) { q.Error = wantErr })
	out, err = sts.New(cfg).GetCallerIdentityRequest(nil).Send()
	require.Nil(t, out)
	require.True(t, err == wantErr)
}
