// DO NOT EDIT LOCAL SDK - USE v3 okta-sdk-golang FOR API CALLS THAT DO NOT EXIST IN LOCAL SDK
package sdk

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime"

	"github.com/kelseyhightower/envconfig"
	"github.com/okta/terraform-provider-okta/sdk/cache"
	"gopkg.in/yaml.v3"
)

type Client struct {
	// NOTE: do not create and add new resources to this local sdk
	config                     *config
	requestExecutor            *RequestExecutor
	resource                   resource
	Application                *ApplicationResource
	Authenticator              *AuthenticatorResource
	AuthorizationServer        *AuthorizationServerResource
	Domain                     *DomainResource
	EventHook                  *EventHookResource
	Feature                    *FeatureResource
	Group                      *GroupResource
	GroupSchema                *GroupSchemaResource
	IdentityProvider           *IdentityProviderResource
	InlineHook                 *InlineHookResource
	LinkedObject               *LinkedObjectResource
	LogEvent                   *LogEventResource
	NetworkZone                *NetworkZoneResource
	OrgSetting                 *OrgSettingResource
	Policy                     *PolicyResource
	ProfileMapping             *ProfileMappingResource
	Session                    *SessionResource
	SmsTemplate                *SmsTemplateResource
	Subscription               *SubscriptionResource
	ThreatInsightConfiguration *ThreatInsightConfigurationResource
	TrustedOrigin              *TrustedOriginResource
	User                       *UserResource
	UserFactor                 *UserFactorResource
	UserSchema                 *UserSchemaResource
	UserType                   *UserTypeResource
	// NOTE: do not create and add new resources to this local sdk
}

type resource struct {
	client *Client
}

type clientContextKey struct{}

func NewClient(ctx context.Context, conf ...ConfigSetter) (context.Context, *Client, error) {
	config := &config{}

	setConfigDefaults(config)
	config = readConfigFromSystem(*config)
	config = readConfigFromApplication(*config)
	config = readConfigFromEnvironment(*config)

	for _, confSetter := range conf {
		confSetter(config)
	}

	var oktaCache cache.Cache
	if !config.Okta.Client.Cache.Enabled {
		oktaCache = cache.NewNoOpCache()
	} else {
		if config.CacheManager == nil {
			oktaCache = cache.NewGoCache(config.Okta.Client.Cache.DefaultTtl,
				config.Okta.Client.Cache.DefaultTti)
		} else {
			oktaCache = config.CacheManager
		}
	}

	config.CacheManager = oktaCache

	config, err := validateConfig(config)
	if err != nil {
		return nil, nil, err
	}

	c := &Client{}
	c.config = config
	c.requestExecutor = NewRequestExecutor(config.HttpClient, oktaCache, config)

	c.resource.client = c

	c.Application = (*ApplicationResource)(&c.resource)
	c.Authenticator = (*AuthenticatorResource)(&c.resource)
	c.AuthorizationServer = (*AuthorizationServerResource)(&c.resource)
	c.Domain = (*DomainResource)(&c.resource)
	c.EventHook = (*EventHookResource)(&c.resource)
	c.Feature = (*FeatureResource)(&c.resource)
	c.Group = (*GroupResource)(&c.resource)
	c.GroupSchema = (*GroupSchemaResource)(&c.resource)
	c.IdentityProvider = (*IdentityProviderResource)(&c.resource)
	c.InlineHook = (*InlineHookResource)(&c.resource)
	c.LinkedObject = (*LinkedObjectResource)(&c.resource)
	c.LogEvent = (*LogEventResource)(&c.resource)
	c.NetworkZone = (*NetworkZoneResource)(&c.resource)
	c.OrgSetting = (*OrgSettingResource)(&c.resource)
	c.Policy = (*PolicyResource)(&c.resource)
	c.ProfileMapping = (*ProfileMappingResource)(&c.resource)
	c.Session = (*SessionResource)(&c.resource)
	c.SmsTemplate = (*SmsTemplateResource)(&c.resource)
	c.Subscription = (*SubscriptionResource)(&c.resource)
	c.ThreatInsightConfiguration = (*ThreatInsightConfigurationResource)(&c.resource)
	c.TrustedOrigin = (*TrustedOriginResource)(&c.resource)
	c.User = (*UserResource)(&c.resource)
	c.UserFactor = (*UserFactorResource)(&c.resource)
	c.UserSchema = (*UserSchemaResource)(&c.resource)
	c.UserType = (*UserTypeResource)(&c.resource)

	contextReturn := context.WithValue(ctx, clientContextKey{}, c)

	return contextReturn, c, nil
}

func ClientFromContext(ctx context.Context) (*Client, bool) {
	u, ok := ctx.Value(clientContextKey{}).(*Client)
	return u, ok
}

func (c *Client) GetConfig() *config {
	return c.config
}

func (c *Client) SetConfig(conf ...ConfigSetter) (err error) {
	config := c.config
	for _, confSetter := range conf {
		confSetter(config)
	}
	_, err = validateConfig(config)
	if err != nil {
		return
	}
	c.config = config
	return
}

// GetRequestExecutor returns underlying request executor
// Deprecated: please use CloneRequestExecutor() to avoid race conditions
func (c *Client) GetRequestExecutor() *RequestExecutor {
	return c.requestExecutor
}

// CloneRequestExecutor create a clone of the underlying request executor
func (c *Client) CloneRequestExecutor() *RequestExecutor {
	a := *c.requestExecutor
	return &a
}

func setConfigDefaults(c *config) {
	conf := []ConfigSetter{
		WithConnectionTimeout(60),
		WithCache(true),
		WithCacheTtl(300),
		WithCacheTti(300),
		WithUserAgentExtra(""),
		WithTestingDisableHttpsCheck(false),
		WithRequestTimeout(0),
		WithRateLimitMaxBackOff(30),
		WithRateLimitMaxRetries(2),
		WithAuthorizationMode("SSWS"),
	}
	for _, confSetter := range conf {
		confSetter(c)
	}
}

func readConfigFromFile(location string, c config) (*config, error) {
	yamlConfig, err := os.ReadFile(location)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlConfig, &c)
	if err != nil {
		return nil, err
	}
	return &c, err
}

func readConfigFromSystem(c config) *config {
	currUser, err := user.Current()
	if err != nil {
		return &c
	}
	if currUser.HomeDir == "" {
		return &c
	}
	conf, err := readConfigFromFile(currUser.HomeDir+"/.okta/okta.yaml", c)
	if err != nil {
		return &c
	}
	return conf
}

// read config from the project's root directory
func readConfigFromApplication(c config) *config {
	_, b, _, _ := runtime.Caller(0)
	conf, err := readConfigFromFile(filepath.Join(filepath.Dir(path.Join(path.Dir(b))), ".okta.yaml"), c)
	if err != nil {
		return &c
	}
	return conf
}

func readConfigFromEnvironment(c config) *config {
	err := envconfig.Process("okta", &c)
	if err != nil {
		fmt.Println("error parsing")
		return &c
	}
	return &c
}

func boolPtr(b bool) *bool {
	return &b
}

func Int64Ptr(i int64) *int64 {
	return &i
}
