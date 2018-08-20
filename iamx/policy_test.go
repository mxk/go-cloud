package iamx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManagedPolicyARN(t *testing.T) {
	tests := []struct{ part, name, want string }{
		// Invalid
		{"", "", ""},
		{"", "/", ""},
		{"", "policy", ""},
		{"", "policy/job-function/", ""},
		{"aws", "", ""},

		// Partition/name
		{"", "AdministratorAccess",
			"arn:aws:iam::aws:policy/AdministratorAccess"},
		{"", "policy/ViewOnlyAccess",
			"arn:aws:iam::aws:policy/job-function/ViewOnlyAccess"},
		{"", "service-role/AmazonEC2RoleforSSM",
			"arn:aws:iam::aws:policy/service-role/AmazonEC2RoleforSSM"},
		{"aws-us-gov", "/AdministratorAccess",
			"arn:aws-us-gov:iam::aws:policy/AdministratorAccess"},

		// ARN
		{"", "arn:aws:iam::aws:policy/AdministratorAccess",
			"arn:aws:iam::aws:policy/AdministratorAccess"},
		{"", "arn::iam::aws:policy/AdministratorAccess",
			"arn:aws:iam::aws:policy/AdministratorAccess"},
		{"aws-us-gov", "arn:aws:iam::aws:policy/AdministratorAccess",
			"arn:aws-us-gov:iam::aws:policy/AdministratorAccess"},
	}
	for _, tc := range tests {
		have := ManagedPolicyARN(tc.part, tc.name)
		assert.Equal(t, tc.want, string(have), "part=%s name=%s", tc.part, tc.name)
	}
}

const policyTpl = `{
	"Version": "` + PolicyVersion2012 + `",
	"Statement": [{
		"Effect": "%s",
		"Principal": {"AWS": "%s"},
		"Action": "sts:AssumeRole"
	}]
}`

func TestAssumeRolePolicy(t *testing.T) {
	tests := []*struct {
		effect Effect
		id     string
		doc    string
		p      Policy
	}{{
		effect: Deny,
		id:     "*",
		doc:    fmt.Sprintf(policyTpl, "Deny", "*"),
		p: Policy{
			Version: PolicyVersion2012,
			Statement: []*Statement{{
				Effect:    Deny,
				Principal: NewAWSPrincipal("*"),
				Action:    PolicyMultiVal{"sts:AssumeRole"},
			}},
		},
	}, {
		effect: Allow,
		id:     "000000000000",
		doc:    fmt.Sprintf(policyTpl, "Allow", "000000000000"),
		p: Policy{
			Version: PolicyVersion2012,
			Statement: []*Statement{{
				Effect:    Allow,
				Principal: NewAWSPrincipal("000000000000"),
				Action:    PolicyMultiVal{"sts:AssumeRole"},
			}},
		},
	}, {
		effect: Allow,
		id:     "test",
		doc:    fmt.Sprintf(policyTpl, "Allow", "test"),
		p: Policy{
			Version: PolicyVersion2012,
			Statement: []*Statement{{
				Effect:    Allow,
				Principal: NewAWSPrincipal("test"),
				Action:    PolicyMultiVal{"sts:AssumeRole"},
			}},
		},
	}}
	for _, tc := range tests {
		p := AssumeRolePolicy(tc.effect, tc.id)
		assert.Equal(t, &tc.p, p, "id=%q", tc.id)
		p.Version = ""

		var have, want bytes.Buffer
		require.NoError(t, json.Indent(&have, []byte(*p.Doc()), "", "  "))
		require.NoError(t, json.Indent(&want, []byte(tc.doc), "", "  "))
		assert.Equal(t, want.String(), have.String())

		p, err := ParsePolicy(&tc.doc)
		require.NoError(t, err)
		assert.Equal(t, &tc.p, p)
	}
}

func TestParsePolicy(t *testing.T) {
	tpl := `{"Version":"%s"}`
	doc := fmt.Sprintf(tpl, PolicyVersion2012)
	_, err := ParsePolicy(&doc)
	assert.NoError(t, err)

	doc = fmt.Sprintf(tpl, "2018-02-08")
	_, err = ParsePolicy(&doc)
	assert.Error(t, err)

	_, err = ParsePolicy(nil)
	assert.Error(t, err)

	doc = `{
		"Version": "2012-10-17",
		"Id": "cd3ad3d9-2776-4ef1-a904-4c229d1642ee",
		"Statement": [{
			"Sid": "1",
			"Effect": "Allow",
			"Principal": "*",
			"NotPrincipal": {"AWS":"*"},
			"Action": "*",
			"NotAction": ["a","b"],
			"Resource": "*",
			"NotResource": ["c","d"],
			"Condition": {
				"Bool": {"aws:SecureTransport":"true"},
				"NumericLessThanEquals": {"s3:max-keys":"10"},
				"StringEquals": {"s3:x-amz-server-side-encryption":"AES256"}
			}
		},{
			"Effect": "Deny",
			"Action": ["<e>","f&"]
		}]
	}`
	p, err := ParsePolicy(&doc)
	require.NoError(t, err)
	var want, have bytes.Buffer
	require.NoError(t, json.Indent(&want, []byte(doc), "", "  "))
	require.NoError(t, json.Indent(&have, []byte(*p.Doc()), "", "  "))
	assert.Equal(t, want.String(), have.String())

	doc = url.QueryEscape(doc)
	q, err := ParsePolicy(&doc)
	require.NoError(t, err)
	assert.Equal(t, p, q)
}

func TestPolicyPrincipal(t *testing.T) {
	tests := []*struct {
		p   Principal
		doc string
	}{{
		p:   Principal{Any: true},
		doc: `"*"`,
	}, {
		p:   Principal{PrincipalMap: PrincipalMap{AWS: PolicyMultiVal{"*"}}},
		doc: `{"AWS":"*"}`,
	}, {
		p:   Principal{PrincipalMap: PrincipalMap{Federated: PolicyMultiVal{"*"}}},
		doc: `{"Federated":"*"}`,
	}, {
		p:   Principal{PrincipalMap: PrincipalMap{Service: PolicyMultiVal{"*"}}},
		doc: `{"Service":"*"}`,
	}, {
		p: Principal{PrincipalMap: PrincipalMap{AWS: PolicyMultiVal{"a", "b"},
			Service: PolicyMultiVal{"c"}}},
		doc: `{"AWS":["a","b"],"Service":"c"}`,
	}, {
		p:   *NewAWSPrincipal(),
		doc: `{}`,
	}, {
		p:   *NewAWSPrincipal("a"),
		doc: `{"AWS":"a"}`,
	}, {
		p:   *NewAWSPrincipal("a", "b"),
		doc: `{"AWS":["a","b"]}`,
	}}
	for _, tc := range tests {
		doc, err := json.Marshal(&tc.p)
		require.NoError(t, err)
		assert.Equal(t, tc.doc, string(doc))
		var p Principal
		require.NoError(t, json.Unmarshal(doc, &p))
		assert.Equal(t, &tc.p, &p)
	}

	var p Principal
	require.Error(t, json.Unmarshal([]byte(`""`), &p))
	require.Error(t, json.Unmarshal([]byte(`"x"`), &p))

	p = *NewAWSPrincipal("")
	p.Any = true
	_, err := json.Marshal(&p)
	require.Error(t, err)
}

func TestPolicyMultiVal(t *testing.T) {
	tests := []*struct {
		in   []byte
		want PolicyMultiVal
	}{{
		[]byte(`[]`),
		PolicyMultiVal{},
	}, {
		[]byte(`""`),
		PolicyMultiVal{""},
	}, {
		[]byte(`"a"`),
		PolicyMultiVal{"a"},
	}, {
		[]byte(`["a","b"]`),
		PolicyMultiVal{"a", "b"},
	}}
	for _, tc := range tests {
		var v PolicyMultiVal
		require.NoError(t, json.Unmarshal(tc.in, &v))
		assert.Equal(t, tc.want, v, "in=%#q", tc.in)
		out, err := json.Marshal(v)
		require.NoError(t, err)
		assert.Equal(t, tc.in, out)
	}
}
