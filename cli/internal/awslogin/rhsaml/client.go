package rhsaml

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	awsconfig "github.com/openshift-online/finops-tools/cli/internal/aws"
)

const (
	defaultSAMLURL              = "https://auth.redhat.com/auth/realms/EmployeeIDP/protocol/saml/clients/itaws"
	defaultRegion               = "us-east-1"
	defaultSessionTimeoutSecond = int32(3600)
	defaultHTTPTimeout          = 30 * time.Second
)

type Client struct {
	SAMLURL               string
	Region                string
	SessionTimeoutSeconds int32
	SPNEGOGetter          spnegoHTTPGetter
	HTTPClient            *http.Client
}

type ObtainRequest struct {
	LookupKeys []string
	Role       string
}

func (c Client) Obtain(ctx context.Context, req ObtainRequest) (awsconfig.ProfileSession, error) {
	if err := hasKerberosTicket(ctx); err != nil {
		return awsconfig.ProfileSession{}, err
	}
	samlURL := strings.TrimSpace(c.SAMLURL)
	if samlURL == "" {
		samlURL = defaultSAMLURL
	}
	region := strings.TrimSpace(c.Region)
	if region == "" {
		region = defaultRegion
	}
	sessionTimeout := c.SessionTimeoutSeconds
	if sessionTimeout <= 0 {
		sessionTimeout = defaultSessionTimeoutSecond
	}

	getter := c.SPNEGOGetter
	if getter == nil {
		getter = defaultSPNEGOHTTPGetter{timeout: defaultHTTPTimeout}
	}
	idpHTML, err := getter.Get(ctx, samlURL)
	if err != nil {
		return awsconfig.ProfileSession{}, fmt.Errorf("fetch SAML login page: %w", err)
	}
	form, err := parseSAMLForm(samlURL, idpHTML)
	if err != nil {
		return awsconfig.ProfileSession{}, err
	}

	accounts, err := c.resolveAccounts(ctx, form)
	if err != nil {
		return awsconfig.ProfileSession{}, err
	}
	account, err := selectAccount(accounts, req.LookupKeys, req.Role)
	if err != nil {
		return awsconfig.ProfileSession{}, err
	}
	return assumeRoleWithSAML(ctx, region, account, form.SAMLAssertion, sessionTimeout)
}

func (c Client) resolveAccounts(ctx context.Context, form samlForm) ([]awsAccount, error) {
	if account, ok, err := parseSingleAccountFromAssertion(form.SAMLAssertion); err != nil {
		return nil, err
	} else if ok {
		return []awsAccount{account}, nil
	}
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultHTTPTimeout}
	}
	formBody := url.Values{"SAMLResponse": []string{form.SAMLAssertion}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, form.AWSURL, strings.NewReader(formBody.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create SAML AWS POST request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("submit SAML assertion to AWS: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read AWS SAML response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("submit SAML assertion to AWS returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return parseAccountsFromHTML(string(body))
}
