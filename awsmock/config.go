// Package awsmock provides mocks for unit testing AWS SDK v2 code.
package awsmock

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/defaults"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
)

// Config returns a config that calls fn instead of defaults.SendHandler. The
// mock handler should set request's Data and/or Error fields as needed to
// simulate AWS response.
func Config(fn func(*aws.Request)) aws.Config {
	cfg := aws.Config{
		Credentials:      aws.AnonymousCredentials,
		EndpointResolver: endpoints.NewDefaultResolver(),
		Handlers:         defaults.Handlers(),
		Retryer:          aws.DefaultRetryer{},
		Logger:           defaults.Logger(),
	}

	// Remove/disable all data-related handlers
	cfg.Handlers.Send.Remove(defaults.SendHandler)
	cfg.Handlers.Send.Remove(defaults.ValidateReqSigHandler)
	cfg.Handlers.ValidateResponse.Remove(defaults.ValidateResponseHandler)
	disableHandlerList("awsmock.Unmarshal", &cfg.Handlers.Unmarshal)
	disableHandlerList("awsmock.UnmarshalMeta", &cfg.Handlers.UnmarshalMeta)

	// Install mock handler
	cfg.Handlers.Send.PushBackNamed(aws.NamedHandler{
		Name: "awsmock.SendHandler",
		Fn: func(req *aws.Request) {
			req.Retryable = aws.Bool(false)
			fn(req)
		},
	})
	return cfg
}

// disableHandlerList prevents a HandlerList from executing any handlers.
func disableHandlerList(name string, l *aws.HandlerList) {
	l.PushFrontNamed(aws.NamedHandler{Name: name, Fn: func(*aws.Request) {}})
	l.AfterEachFn = func(aws.HandlerListRunItem) bool { return false }
}
