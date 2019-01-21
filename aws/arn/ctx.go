package arn

// Ctx maintains location information for constructing context-aware ARNs.
type Ctx struct{ Partition, Region, Account string }

// In returns a new context for the specified region. No validation is performed
// to ensure that the new region is valid for the current partition.
func (c Ctx) In(region string) Ctx {
	c.Region = region
	return c
}

// New constructs a context-aware ARN for the specified service/resource.
func (c Ctx) New(service string, resource ...string) ARN {
	typ := "any"
	switch service {
	case "apigateway":
		c.Account = ""
	case "artifact", "s3":
		c.Region = ""
		c.Account = ""
	case "cloudfront", "iam", "sts", "waf", "waf-regional":
		c.Region = ""
	case "cloudwatch", "monitoring":
		service = "cloudwatch"
		if typ = resType(resource); typ == "dashboard" {
			c.Region = ""
		}
	case "ec2":
		switch typ = resType(resource); typ {
		case "image", "snapshot":
			c.Account = ""
		}
	case "elasticbeanstalk":
		if typ = resType(resource); typ == "solutionstack" {
			c.Account = ""
		}
	case "health":
		if typ = resType(resource); typ == "event" {
			c.Account = ""
		}
	case "route53":
		switch typ = resType(resource); typ {
		case "hostedzone", "change":
			c.Account = ""
			fallthrough
		case "domain":
			c.Region = ""
		}
	case "":
		panic("arn: service not specified")
	}
	if typ == "" {
		panic("arn: " + service + " requires resource")
	}
	return New(c.Partition, service, c.Region, c.Account, resource...)
}

// resType returns the resource prefix up to the first '/' or ':' character.
func resType(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	r := parts[0]
	for i := range r {
		switch r[i] {
		case '/', ':':
			return r[:i]
		}
	}
	return r
}
