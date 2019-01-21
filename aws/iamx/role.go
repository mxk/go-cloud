package iamx

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/mxk/go-fast"
)

// DeleteRoles deletes all roles under the specified IAM path.
func (c Client) DeleteRoles(path string) error {
	in := iam.ListRolesInput{PathPrefix: aws.String(path)}
	r := c.ListRolesRequest(&in)
	p := r.Paginate()
	var roles []string
	for p.Next() {
		out := p.CurrentPage().Roles
		for i := range out {
			roles = append(roles, aws.StringValue(out[i].RoleName))
		}
	}
	if err := p.Err(); err != nil {
		return err
	}
	return fast.ForEachIO(len(roles), func(i int) error {
		return c.DeleteRole(roles[i])
	})
}

// DeleteRole deletes the specified role, ensuring that all prerequisites for
// deletion are met.
func (c Client) DeleteRole(role string) error {
	err := fast.Call(
		func() error { return c.detachRolePolicies(role) },
		func() error { return c.deleteRolePolicies(role) },
	)
	if err == nil {
		in := iam.DeleteRoleInput{RoleName: aws.String(role)}
		_, err = c.DeleteRoleRequest(&in).Send()
	}
	return err
}

// detachRolePolicies detaches all role policies.
func (c Client) detachRolePolicies(role string) error {
	in := iam.ListAttachedRolePoliciesInput{RoleName: aws.String(role)}
	r := c.ListAttachedRolePoliciesRequest(&in)
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
		in := iam.DetachRolePolicyInput{
			PolicyArn: aws.String(arns[i]),
			RoleName:  aws.String(role),
		}
		_, err := c.DetachRolePolicyRequest(&in).Send()
		return err
	})
}

// deleteRolePolicies deletes all inline role policies.
func (c Client) deleteRolePolicies(role string) error {
	in := iam.ListRolePoliciesInput{RoleName: aws.String(role)}
	r := c.ListRolePoliciesRequest(&in)
	p := r.Paginate()
	var names []string
	for p.Next() {
		names = append(names, p.CurrentPage().PolicyNames...)
	}
	if err := p.Err(); err != nil {
		return err
	}
	return fast.ForEachIO(len(names), func(i int) error {
		in := iam.DeleteRolePolicyInput{
			PolicyName: aws.String(names[i]),
			RoleName:   aws.String(role),
		}
		_, err := c.DeleteRolePolicyRequest(&in).Send()
		return err
	})
}
