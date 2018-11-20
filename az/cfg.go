// Package az provides utility types and functions for Azure SDK.
package az

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	azcli "github.com/Azure/go-autorest/autorest/azure/cli"
	"github.com/LuminalHQ/cloudcover/x/cli"
	"github.com/LuminalHQ/cloudcover/x/cloud"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/pkcs12"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// EnvForActiveDirectory returns the environment for the specified Azure AD
// endpoint.
func EnvForActiveDirectory(adEndpoint string) *azure.Environment {
	switch normEndpoint(adEndpoint) {
	case azure.USGovernmentCloud.ActiveDirectoryEndpoint:
		return &azure.USGovernmentCloud
	case azure.ChinaCloud.ActiveDirectoryEndpoint:
		return &azure.ChinaCloud
	case azure.GermanCloud.ActiveDirectoryEndpoint:
		return &azure.GermanCloud
	}
	return &azure.PublicCloud
}

// EnvForLocation returns the environment for the specified location.
func EnvForLocation(loc string) *azure.Environment {
	if len(loc) > 5 {
		switch strings.ToLower(loc[:5]) {
		case "usdod", "usgov":
			return &azure.USGovernmentCloud
		case "china":
			return &azure.ChinaCloud
		case "germa":
			return &azure.GermanCloud
		}
	}
	return &azure.PublicCloud
}

// Env defines environment variables used by the SDK (not to be confused with
// azure.Environment, which defines Azure endpoints).
type Env struct {
	TenantID       string `env:"AZURE_TENANT_ID"`
	ClientID       string `env:"AZURE_CLIENT_ID"`
	SubscriptionID string `env:"AZURE_SUBSCRIPTION"`
	ClientSecret   string `env:"AZURE_CLIENT_SECRET"`
	CertFile       string `env:"AZURE_CERTIFICATE_PATH"`
	CertPass       string `env:"AZURE_CERTIFICATE_PASSWORD"`
	EnvName        string `env:"AZURE_ENVIRONMENT"`
	AuthFile       string `env:"AZURE_AUTH_LOCATION"`

	// AZURE_USERNAME and AZURE_PASSWORD are omitted intentionally.
	// TODO: Handle AZURE_ENVIRONMENT_FILEPATH?
}

// Cfg contains the information necessary to create Azure API clients.
type Cfg struct {
	Src string // Source of config information

	azure.Environment        // Endpoints for the current cloud
	TenantID          string // Active Directory tenant GUID
	SubscriptionID    string // Default subscription GUID

	mu       sync.Mutex
	authz    map[string]autorest.Authorizer
	newAuthz func(resource string) autorest.Authorizer
}

// LoadCfg loads client configuration, automatically selecting environment
// variables, SDK auth file, MSI, or CLI config as the source.
func LoadCfg() (*Cfg, error) {
	var c *Cfg
	var e Env
	err := cli.SetEnvFields(&e)
	if err == nil {
		if e.TenantID != "" {
			c, err = LoadCfgFromEnv(&e)
		} else if e.AuthFile != "" {
			c, err = LoadCfgFromFile(e.AuthFile)
		} else if cloud.Ident().Type == cloud.Azure {
			c, err = LoadCfgFromMSI()
		} else {
			c, err = LoadCfgFromCLI()
		}
		if err == nil && e.SubscriptionID != "" {
			// AZURE_SUBSCRIPTION overrides default subscription ID
			c.SubscriptionID = e.SubscriptionID
		}
	}
	return c, err
}

// NilGUID is an all-zero GUID.
const NilGUID = "00000000-0000-0000-0000-000000000000"

// TestCfg returns a mock Cfg that can be used for unit testing.
func TestCfg(url string) *Cfg {
	if url == "" {
		url = "http://127.0.0.1/"
	} else if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	return &Cfg{
		Src: "test",
		Environment: azure.Environment{
			Name:                      azure.PublicCloud.Name,
			ManagementPortalURL:       url,
			PublishSettingsURL:        url + "publishsettings/index",
			ServiceManagementEndpoint: url,
			ResourceManagerEndpoint:   url,
			ActiveDirectoryEndpoint:   url,
			GalleryEndpoint:           url,
			KeyVaultEndpoint:          url,
			GraphEndpoint:             url,
			ServiceBusEndpoint:        url,
			BatchManagementEndpoint:   url,
			TokenAudience:             url,
		},
		TenantID:       NilGUID,
		SubscriptionID: NilGUID,
		newAuthz: func(string) autorest.Authorizer {
			return autorest.NullAuthorizer{}
		},
	}
}

// LoadCfgFromEnv loads client configuration from environment variables. If e is
// nil, the environment is loaded automatically.
func LoadCfgFromEnv(e *Env) (*Cfg, error) {
	var env Env
	var err error
	if e == nil {
		if err = cli.SetEnvFields(&env); err != nil {
			return nil, err
		}
		e = &env
	}
	c := &Cfg{
		Src:            "Environment",
		TenantID:       e.TenantID,
		SubscriptionID: e.SubscriptionID,
	}
	if e.EnvName == "" || strings.EqualFold(e.EnvName, "AzureCloud") {
		c.Environment = azure.PublicCloud
	} else {
		c.Environment, err = azure.EnvironmentFromName(e.EnvName)
		if err != nil {
			return nil, err
		}
	}
	if e.ClientSecret != "" {
		err = c.useClientSecret(e.ClientID, e.ClientSecret)
	} else if e.CertFile != "" {
		err = c.useClientCert(e.ClientID, e.CertFile, e.CertPass)
	} else {
		err = fmt.Errorf("az: client secret or cert file must be specified")
	}
	if err != nil {
		c = nil
	}
	return c, err
}

// LoadCfgFromFile loads client configuration from an SDK auth file created by
// 'az ad sp create-for-rbac --sdk-auth' command. The file name defaults to
// AZURE_AUTH_LOCATION environment variable if empty.
func LoadCfgFromFile(name string) (*Cfg, error) {
	if name == "" {
		name = os.Getenv("AZURE_AUTH_LOCATION")
	}
	af, err := loadAuthFile(name)
	if err != nil {
		return nil, err
	}
	c := &Cfg{
		Src:            name,
		Environment:    *EnvForActiveDirectory(af.ActiveDirectoryEndpointURL),
		TenantID:       af.TenantID,
		SubscriptionID: af.SubscriptionID,
	}
	af.updateEnv(&c.Environment)
	if err = c.useClientSecret(af.ClientID, af.ClientSecret); err != nil {
		c = nil
	}
	return c, err
}

const msiEndpoint = "http://169.254.169.254/metadata/identity/oauth2/token"

// LoadCfgFromMSI loads client configuration from the Managed Service Identity
// endpoint.
func LoadCfgFromMSI() (*Cfg, error) {
	// Get location and subscription ID from instance metadata
	im, err := GetInstanceMetadata()
	if err != nil {
		return nil, err
	}
	c := &Cfg{
		Src:            "MSI",
		Environment:    *EnvForLocation(im.Compute.Location),
		SubscriptionID: im.Compute.SubscriptionID,
	}

	// Extract tenant ID from an access token
	initRes := c.ResourceManagerEndpoint
	sp, err := adal.NewServicePrincipalTokenFromMSI(msiEndpoint, initRes)
	if err != nil {
		return nil, err
	}
	if err = sp.EnsureFresh(); err != nil {
		return nil, err
	}
	var claims struct {
		jwt.StandardClaims
		TenantID string `json:"tid"`
	}
	_, _, err = new(jwt.Parser).ParseUnverified(sp.OAuthToken(), &claims)
	if err != nil {
		return nil, err
	}
	if c.TenantID = claims.TenantID; claims.TenantID == "" {
		return nil, fmt.Errorf("az: failed to determine tenant ID from token")
	}

	// Configure authorizer
	c.authz = map[string]autorest.Authorizer{
		initRes: autorest.NewBearerAuthorizer(sp),
	}
	c.newAuthz = func(resource string) autorest.Authorizer {
		sp, err := adal.NewServicePrincipalTokenFromMSI(msiEndpoint, resource)
		if err != nil {
			panic(err)
		}
		return autorest.NewBearerAuthorizer(sp)
	}
	return c, nil
}

// LoadCfgFromCLI loads client configuration from Azure CLI.
func LoadCfgFromCLI() (*Cfg, error) {
	// Parse CLI config files
	sub, err := cliSubscription()
	if err != nil {
		return nil, err
	}
	tok, err := cliToken(sub.TenantID, sub.User.Name)
	if err != nil {
		return nil, err
	}
	c := &Cfg{
		Src:            "CLI",
		TenantID:       sub.TenantID,
		SubscriptionID: sub.ID,
	}

	// Find matching environment
	if sub.EnvironmentName == "AzureCloud" {
		c.Environment = azure.PublicCloud
	} else {
		c.Environment, err = azure.EnvironmentFromName(sub.EnvironmentName)
		if err != nil {
			return nil, err
		}
	}

	// Configure authorizer
	oauth, err := adal.NewOAuthConfig(c.ActiveDirectoryEndpoint, c.TenantID)
	if err != nil {
		return nil, err
	}
	switch sub.User.Type {
	case "user":
		c.newAuthz = func(resource string) autorest.Authorizer {
			var at adal.Token
			if resource == tok.Resource {
				at, _ = tok.ToADALToken()
			}
			if at.RefreshToken == "" {
				at = adal.Token{
					RefreshToken: tok.RefreshToken,
					ExpiresIn:    "0",
					ExpiresOn:    "0",
					NotBefore:    "0",
					Resource:     resource,
					Type:         tok.TokenType,
				}
			}
			sp, err := adal.NewServicePrincipalTokenFromManualToken(
				*oauth, tok.ClientID, resource, at)
			if err != nil {
				panic(err)
			}
			return autorest.NewBearerAuthorizer(sp)
		}
	case "servicePrincipal":
		c.newAuthz = func(resource string) autorest.Authorizer {
			sp, err := adal.NewServicePrincipalToken(
				*oauth, tok.ServicePrincipalID, tok.AccessToken, resource)
			if err != nil {
				panic(err)
			}
			return autorest.NewBearerAuthorizer(sp)
		}
	default:
		return nil, fmt.Errorf("az: unsupported subscription user type %q",
			sub.User.Type)
	}
	return c, nil
}

// Authorizer returns an authorizer for the specified resource.
func (c *Cfg) Authorizer(resource string) autorest.Authorizer {
	c.mu.Lock()
	defer c.mu.Unlock()
	authz := c.authz[resource]
	if authz == nil {
		authz = c.newAuthz(resource)
		if c.authz == nil {
			c.authz = make(map[string]autorest.Authorizer)
		}
		c.authz[resource] = authz
	}
	return authz
}

// useClientSecret configures an authorizer using client secret.
func (c *Cfg) useClientSecret(clientID, secret string) error {
	oauth, err := adal.NewOAuthConfig(c.ActiveDirectoryEndpoint, c.TenantID)
	if err != nil {
		return err
	}
	c.newAuthz = func(resource string) autorest.Authorizer {
		sp, err := adal.NewServicePrincipalToken(
			*oauth, clientID, secret, resource)
		if err != nil {
			panic(err)
		}
		return autorest.NewBearerAuthorizer(sp)
	}
	return nil
}

// useClientCert configures an authorizer using client certificate.
func (c *Cfg) useClientCert(clientID, certFile, password string) error {
	b, err := ioutil.ReadFile(certFile)
	if err != nil {
		return err
	}
	key, cert, err := pkcs12.Decode(b, password)
	if err != nil {
		return err
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("az: no RSA private key in %q", certFile)
	}
	oauth, err := adal.NewOAuthConfig(c.ActiveDirectoryEndpoint, c.TenantID)
	if err != nil {
		return err
	}
	c.newAuthz = func(resource string) autorest.Authorizer {
		sp, err := adal.NewServicePrincipalTokenFromCertificate(
			*oauth, clientID, cert, rsaKey, resource)
		if err != nil {
			panic(err)
		}
		return autorest.NewBearerAuthorizer(sp)
	}
	return nil
}

type authFile struct {
	ClientID                       string
	ClientSecret                   string
	SubscriptionID                 string
	TenantID                       string
	ActiveDirectoryEndpointURL     string
	ResourceManagerEndpointURL     string
	ActiveDirectoryGraphResourceID string
	SQLManagementEndpointURL       string
	GalleryEndpointURL             string
	ManagementEndpointURL          string
}

// loadAuthFile loads the contents of an SDK auth file.
func loadAuthFile(name string) (af authFile, err error) {
	b, err := ioutil.ReadFile(name)
	if err == nil {
		b, _, err = transform.Bytes(unicode.BOMOverride(transform.Nop), b)
		if err == nil {
			err = json.Unmarshal(b, &af)
		}
	}
	return
}

// updateEnv overrides default Azure endpoints with those specified in the SDK
// auth file.
func (f *authFile) updateEnv(e *azure.Environment) {
	e.ActiveDirectoryEndpoint = normEndpoint(f.ActiveDirectoryEndpointURL)
	e.ResourceManagerEndpoint = normEndpoint(f.ResourceManagerEndpointURL)
	e.GraphEndpoint = normEndpoint(f.ActiveDirectoryGraphResourceID)
	e.GalleryEndpoint = normEndpoint(f.GalleryEndpointURL)
	e.ServiceManagementEndpoint = normEndpoint(f.ManagementEndpointURL)
}

// cliSubscription loads the default CLI subscription.
func cliSubscription() (*azcli.Subscription, error) {
	profilePath, err := azcli.ProfilePath()
	if err != nil {
		return nil, err
	}
	prof, err := azcli.LoadProfile(profilePath)
	if err != nil {
		return nil, err
	}
	for i := range prof.Subscriptions {
		if sub := &prof.Subscriptions[i]; sub.IsDefault {
			return sub, nil
		}
	}
	// TODO: Use first one if no default?
	return nil, fmt.Errorf("az: subscription not found in %q", profilePath)
}

// spToken adds service principal support to azcli.Token.
type spToken struct {
	azcli.Token
	ServicePrincipalID     string `json:"servicePrincipalId"`
	ServicePrincipalTenant string `json:"servicePrincipalTenant"`
}

// cliToken loads the CLI access token for the given tenant/user IDs.
func cliToken(tenantID, userID string) (*spToken, error) {
	tokensPath, err := azcli.AccessTokensPath()
	if err != nil {
		return nil, err
	}
	// azcli.LoadTokens doesn't handle service principals
	b, err := ioutil.ReadFile(tokensPath)
	if err != nil {
		return nil, err
	}
	var toks []*spToken
	if err = json.Unmarshal(b, &toks); err != nil {
		return nil, err
	}
	for _, t := range toks {
		if (t.IsMRRT && t.UserID == userID &&
			strings.HasSuffix(t.Authority, tenantID)) ||
			(t.ServicePrincipalID == userID &&
				t.ServicePrincipalTenant == tenantID) {
			return t, nil
		}
	}
	return nil, fmt.Errorf("az: access token not found in %q", tokensPath)
}

// normEndpoint normalizes endpoint URL.
func normEndpoint(url string) string {
	if len(url) > 0 {
		if url[len(url)-1] != '/' {
			url += "/"
		}
		return url
	}
	panic("az: invalid endpoint")
}
