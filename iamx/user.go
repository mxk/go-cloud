package iamx

import (
	"github.com/LuminalHQ/cloudcover/x/fast"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// Client is an extended IAM client with additional methods for managing users
// and roles.
type Client struct{ iam.IAM }

// New returns a new extended IAM client.
func New(cfg *aws.Config) Client {
	return Client{*iam.New(*cfg)}
}

// DeleteUsers deletes all users under the specified IAM path.
func (c Client) DeleteUsers(path string) error {
	in := iam.ListUsersInput{PathPrefix: aws.String(path)}
	r := c.ListUsersRequest(&in)
	p := r.Paginate()
	var users []string
	for p.Next() {
		out := p.CurrentPage().Users
		for i := range out {
			users = append(users, aws.StringValue(out[i].UserName))
		}
	}
	if err := p.Err(); err != nil {
		return err
	}
	return fast.ForEachIO(len(users), func(i int) error {
		return c.DeleteUser(users[i])
	})
}

// DeleteUser deletes the specified user, ensuring that all prerequisites for
// deletion are met.
func (c Client) DeleteUser(name string) error {
	err := fast.Call(
		func() error { return c.detachUserPolicies(name) },
		func() error { return c.deleteAccessKeys(name) },
	)
	if err == nil {
		in := iam.DeleteUserInput{UserName: aws.String(name)}
		_, err = c.DeleteUserRequest(&in).Send()
	}
	return err
}

// deleteAccessKeys deletes all user access keys.
func (c Client) deleteAccessKeys(user string) error {
	in := iam.ListAccessKeysInput{UserName: aws.String(user)}
	r := c.ListAccessKeysRequest(&in)
	p := r.Paginate()
	var ids []string
	for p.Next() {
		out := p.CurrentPage().AccessKeyMetadata
		for i := range out {
			ids = append(ids, aws.StringValue(out[i].AccessKeyId))
		}
	}
	if err := p.Err(); err != nil {
		return err
	}
	return fast.ForEachIO(len(ids), func(i int) error {
		in := iam.DeleteAccessKeyInput{
			AccessKeyId: aws.String(ids[i]),
			UserName:    aws.String(user),
		}
		_, err := c.DeleteAccessKeyRequest(&in).Send()
		return err
	})
}

// detachUserPolicies detaches all user policies.
func (c Client) detachUserPolicies(user string) error {
	in := iam.ListAttachedUserPoliciesInput{UserName: aws.String(user)}
	r := c.ListAttachedUserPoliciesRequest(&in)
	p := r.Paginate()
	var arns []string
	for p.Next() {
		out := p.CurrentPage().AttachedPolicies
		for i := range out {
			arns = append(arns, aws.StringValue(out[i].PolicyArn))
		}
	}
	if err := p.Err(); err != nil {
		return err
	}
	return fast.ForEachIO(len(arns), func(i int) error {
		in := iam.DetachUserPolicyInput{
			PolicyArn: aws.String(arns[i]),
			UserName:  aws.String(user),
		}
		_, err := c.DetachUserPolicyRequest(&in).Send()
		return err
	})
}
