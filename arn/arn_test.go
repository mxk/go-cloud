package arn

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestARN(t *testing.T) {
	tests := []*struct {
		arn       ARN
		partition string
		service   string
		region    string
		account   string
		resource  string
		typ       string
		path      string
		name      string
	}{{
		arn: Base,
	}, {
		arn:       "arn:a:b:c:*:x",
		partition: "a",
		service:   "b",
		region:    "c",
		account:   "*",
		resource:  "x",
		name:      "x",
	}, {
		arn:      "arn::::*:x:",
		account:  "*",
		resource: "x:",
		typ:      "x",
	}, {
		arn:      "arn::::*::y",
		account:  "*",
		resource: ":y",
		name:     "y",
	}, {
		arn:      "arn::::*:x:y",
		account:  "*",
		resource: "x:y",
		typ:      "x",
		name:     "y",
	}, {
		arn:      "arn::::*:x/",
		account:  "*",
		resource: "x/",
		typ:      "x",
		path:     "/",
	}, {
		arn:      "arn::::*:/y",
		account:  "*",
		resource: "/y",
		path:     "/",
		name:     "y",
	}, {
		arn:      "arn::::*:x/y",
		account:  "*",
		resource: "x/y",
		typ:      "x",
		path:     "/",
		name:     "y",
	}, {
		arn:      "arn::::*:x/y/z",
		account:  "*",
		resource: "x/y/z",
		typ:      "x",
		path:     "/y/",
		name:     "z",
	}}
	for _, tc := range tests {
		require.True(t, tc.arn.Valid(), "arn=%s", tc.arn)
		assert.Equal(t, tc.partition, tc.arn.Partition(), "arn=%s", tc.arn)
		assert.Equal(t, tc.service, tc.arn.Service(), "arn=%s", tc.arn)
		assert.Equal(t, tc.region, tc.arn.Region(), "arn=%s", tc.arn)
		assert.Equal(t, tc.account, tc.arn.Account(), "arn=%s", tc.arn)
		assert.Equal(t, tc.resource, tc.arn.Resource(), "arn=%s", tc.arn)
		assert.Equal(t, tc.typ, tc.arn.Type(), "arn=%s", tc.arn)
		assert.Equal(t, tc.path, tc.arn.Path(), "arn=%s", tc.arn)
		assert.Equal(t, tc.name, tc.arn.Name(), "arn=%s", tc.arn)

		r := New(tc.partition, tc.service, tc.region, tc.account, tc.resource)
		assert.Equal(t, tc.arn, r)

		assert.Equal(t, r, r.WithPartition(tc.partition))
		assert.Equal(t, r, r.WithService(tc.service))
		assert.Equal(t, r, r.WithRegion(tc.region))
		assert.Equal(t, r, r.WithAccount(tc.account))
		assert.Equal(t, r, r.WithResource(tc.resource))
		if tc.name != "" {
			assert.Equal(t, r, r.WithName(tc.name))
		}
		if tc.path != "" {
			assert.Equal(t, r, r.WithPath(tc.path))
			assert.Equal(t, r, r.WithPathName(tc.path+tc.name))
		}

		assert.Equal(t, r, r.With(New(tc.partition, "", tc.region, "", tc.resource)))
		assert.Equal(t, r, r.With(New("", tc.service, "", tc.account, "")))

		if tc.path != "" {
			r := New("", "", "", tc.account, tc.typ, tc.path, tc.name)
			assert.Equal(t, tc.arn, r)
		}
	}
}

func TestField(t *testing.T) {
	r := Base.WithResource("x:y:z")
	assert.Equal(t, "x", r.Type())
	assert.Equal(t, "x:y:z", r.Field(4))
	assert.Equal(t, "y:z", r.Field(5))
	assert.Equal(t, "z", r.Field(6))
}

func TestInvalid(t *testing.T) {
	assert.False(t, prefix.Valid())
	assert.False(t, ARN("").Valid())
	assert.False(t, ARN(Base[3:]).Valid())
	assert.False(t, ARN(Base[:len(Base)-1]).Valid())
	assert.Panics(t, func() { prefix.Partition() })
	assert.Panics(t, func() { ARN("").Type() })
	assert.PanicsWithValue(t, `arn: invalid arn: `, func() { ARN("").Path() })
	assert.PanicsWithValue(t, `arn: invalid arn: x`, func() { ARN("x").Name() })
	assert.PanicsWithValue(t, `arn: invalid arn or field index 4: arn::::`,
		func() { ARN("arn::::").Resource() })
}

func TestNoPath(t *testing.T) {
	assert.Panics(t, func() { ARN(Base).WithPath("/") })
	assert.Panics(t, func() { ARN(Base).WithPathName("/a") })
}

func TestCleanPath(t *testing.T) {
	tests := []*struct{ in, out string }{
		{in: "", out: ""},
		{in: "a", out: "/a"},
		{in: "/", out: ""},
		{in: "a/", out: "/a"},
		{in: "/a", out: "/a"},
		{in: "/a/", out: "/a"},
		{in: "a/b", out: "/a/b"},
		{in: "//a/b", out: "/a/b"},
		{in: "a/./b//c/", out: "/a/b/c"},
	}
	for _, tc := range tests {
		path := string(cleanPath(tc.in))
		assert.Equal(t, tc.out, path, "in=%q", tc.in)
	}
}

func TestConv(t *testing.T) {
	assert.Equal(t, Base, Value(String(Base)))
	assert.Equal(t, ARN(""), Value(nil))
}
