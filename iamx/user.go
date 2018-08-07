package iamx

import (
	"github.com/LuminalHQ/cloudcover/x/fast"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// DeleteUsers deletes all users under the specified IAM path.
func DeleteUsers(c iam.IAM, path string) error {
	in := iam.ListUsersInput{PathPrefix: aws.String(path)}
	r := c.ListUsersRequest(&in)
	p := r.Paginate()
	var users []string
	for p.Next() {
		for _, u := range p.CurrentPage().Users {
			users = append(users, aws.StringValue(u.UserName))
		}
	}
	if err := p.Err(); err != nil {
		return err
	}
	return fast.ForEachIO(len(users), func(i int) error {
		return DeleteUser(c, users[i])
	})
}

// DeleteUser deletes the specified user, ensuring that all prerequisites for
// deletion are met.
func DeleteUser(c iam.IAM, name string) error {
	err := fast.Call(
		func() error { return detachUserPolicies(c, name) },
		func() error { return deleteAccessKeys(c, name) },
	)
	if err == nil {
		in := iam.DeleteUserInput{UserName: aws.String(name)}
		_, err = c.DeleteUserRequest(&in).Send()
	}
	return err
}

// deleteAccessKeys deletes all user access keys.
func deleteAccessKeys(c iam.IAM, user string) error {
	in := iam.ListAccessKeysInput{UserName: aws.String(user)}
	r := c.ListAccessKeysRequest(&in)
	p := r.Paginate()
	var ids []string
	for p.Next() {
		for _, key := range p.CurrentPage().AccessKeyMetadata {
			ids = append(ids, aws.StringValue(key.AccessKeyId))
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
func detachUserPolicies(c iam.IAM, user string) error {
	in := iam.ListAttachedUserPoliciesInput{UserName: aws.String(user)}
	r := c.ListAttachedUserPoliciesRequest(&in)
	p := r.Paginate()
	var arns []string
	for p.Next() {
		for _, pol := range p.CurrentPage().AttachedPolicies {
			arns = append(arns, aws.StringValue(pol.PolicyArn))
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
