package arn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCtx(t *testing.T) {
	tests := []*struct {
		svc string
		res string
		out ARN
	}{{
		svc: "apigateway",
		res: "/restapis/a123456789012bc3de45678901f23a45/*",
		out: "arn:aws:apigateway:us-east-1::/restapis/a123456789012bc3de45678901f23a45/*",
	}, {
		svc: "cloudfront",
		res: "*",
		out: "arn:aws:cloudfront::123456789012:*",
	}, {
		svc: "cloudwatch",
		res: "alarm:*",
		out: "arn:aws:cloudwatch:us-east-1:123456789012:alarm:*",
	}, {
		svc: "ec2",
		res: "dedicated-host/h-12345678",
		out: "arn:aws:ec2:us-east-1:123456789012:dedicated-host/h-12345678",
	}, {
		svc: "ec2",
		res: "image/ami-1a2b3c4d",
		out: "arn:aws:ec2:us-east-1::image/ami-1a2b3c4d",
	}, {
		svc: "elasticbeanstalk",
		res: "solutionstack/32bit Amazon Linux running Tomcat 7",
		out: "arn:aws:elasticbeanstalk:us-east-1::solutionstack/32bit Amazon Linux running Tomcat 7",
	}, {
		svc: "health",
		res: "event/AWS_EC2_EXAMPLE_ID",
		out: "arn:aws:health:us-east-1::event/AWS_EC2_EXAMPLE_ID",
	}, {
		svc: "monitoring",
		res: "dashboard/MyDashboardName",
		out: "arn:aws:cloudwatch::123456789012:dashboard/MyDashboardName",
	}, {
		svc: "route53",
		res: "hostedzone/Z148QEXAMPLE8V",
		out: "arn:aws:route53:::hostedzone/Z148QEXAMPLE8V",
	}, {
		svc: "route53",
		res: "change/C2RDJ5EXAMPLE2",
		out: "arn:aws:route53:::change/C2RDJ5EXAMPLE2",
	}, {
		svc: "route53",
		res: "domain:example.com",
		out: "arn:aws:route53::123456789012:domain:example.com",
	}, {
		svc: "s3",
		res: "my_corporate_bucket",
		out: "arn:aws:s3:::my_corporate_bucket",
	}}
	arn := Ctx{"aws", "us-east-1", "123456789012"}
	for _, tc := range tests {
		assert.Equal(t, ARN(tc.out), arn.New(tc.svc, tc.res))
	}
	assert.Panics(t, func() { arn.New("", "xyz") })
	assert.Panics(t, func() { arn.New("cloudwatch") })
	assert.Panics(t, func() { arn.New("route53", "") })
	assert.Equal(t, Ctx{"aws", "us-west-2", "123456789012"}, arn.In("us-west-2"))
}
