package rhsaml

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var roleValueRegexp = regexp.MustCompile(`arn:aws:iam::([0-9]{12}):role/([^,\s]+)`)

type awsAccount struct {
	Name     string
	UID      string
	RoleName string
	RoleARN  string
}

func parseSingleAccountFromAssertion(assertion string) (awsAccount, bool, error) {
	decoded, err := base64.StdEncoding.DecodeString(assertion)
	if err != nil {
		return awsAccount{}, false, fmt.Errorf("decode SAML assertion: %w", err)
	}
	var envelope samlEnvelope
	if err := xml.Unmarshal(decoded, &envelope); err != nil {
		return awsAccount{}, false, fmt.Errorf("parse SAML assertion XML: %w", err)
	}
	if len(envelope.RoleValues) != 1 {
		return awsAccount{}, false, nil
	}
	account, err := parseRoleValue(envelope.RoleValues[0])
	if err != nil {
		return awsAccount{}, false, err
	}
	account.Name = account.UID
	return account, true, nil
}

func parseAccountsFromHTML(html string) ([]awsAccount, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parse AWS SAML account page: %w", err)
	}
	var accounts []awsAccount
	doc.Find("div.saml-account").Each(func(_ int, accountSel *goquery.Selection) {
		rawName := strings.TrimSpace(accountSel.Find(".saml-account-name").First().Text())
		name := parseSAMLAccountName(rawName)
		if name == "" {
			return
		}
		accountSel.Find(".saml-role label").Each(func(_ int, roleSel *goquery.Selection) {
			roleARN, ok := roleSel.Attr("for")
			if !ok {
				return
			}
			roleName := strings.TrimSpace(roleSel.Text())
			account, err := parseRoleValue(roleARN)
			if err != nil {
				return
			}
			account.Name = name
			if roleName != "" {
				account.RoleName = roleName
			}
			accounts = append(accounts, account)
		})
	})
	if len(accounts) == 0 {
		return nil, fmt.Errorf("no AWS accounts found in SAML response")
	}
	return accounts, nil
}

func selectAccount(accounts []awsAccount, keys []string, role string) (awsAccount, error) {
	if len(accounts) == 0 {
		return awsAccount{}, fmt.Errorf("no AWS accounts available")
	}
	if len(accounts) == 1 {
		return accounts[0], nil
	}
	trimmedRole := strings.TrimSpace(role)
	candidates := normalizeLookupKeys(keys)
	for _, key := range candidates {
		keyName := key
		keyRole := trimmedRole
		if keyRole == "" {
			name, parsedRole, ok := strings.Cut(key, "/")
			if ok {
				keyName = strings.TrimSpace(name)
				keyRole = strings.TrimSpace(parsedRole)
			}
		}
		for _, account := range accounts {
			if !matchesLookupKey(account, keyName) {
				continue
			}
			if keyRole != "" && !strings.EqualFold(account.RoleName, keyRole) {
				continue
			}
			return account, nil
		}
	}
	return awsAccount{}, fmt.Errorf("account not found for %v; available: %s", candidates, renderAccounts(accounts))
}

func parseRoleValue(value string) (awsAccount, error) {
	match := roleValueRegexp.FindStringSubmatch(strings.TrimSpace(value))
	if len(match) != 3 {
		return awsAccount{}, fmt.Errorf("invalid role value %q", value)
	}
	return awsAccount{
		UID:      match[1],
		RoleName: match[2],
		RoleARN:  match[0],
	}, nil
}

func parseSAMLAccountName(raw string) string {
	fields := strings.Fields(strings.TrimSpace(raw))
	if len(fields) < 2 {
		return ""
	}
	return strings.TrimSpace(fields[1])
}

func normalizeLookupKeys(keys []string) []string {
	uniq := make(map[string]struct{}, len(keys))
	out := make([]string, 0, len(keys))
	for _, key := range keys {
		trimmed := strings.TrimSpace(key)
		if trimmed == "" {
			continue
		}
		if _, seen := uniq[trimmed]; seen {
			continue
		}
		uniq[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func matchesLookupKey(account awsAccount, key string) bool {
	if key == "" {
		return false
	}
	return strings.EqualFold(account.Name, key) || account.UID == key
}

func renderAccounts(accounts []awsAccount) string {
	items := make([]string, 0, len(accounts))
	for _, account := range accounts {
		items = append(items, fmt.Sprintf("%s (%s) role=%s", account.Name, account.UID, account.RoleName))
	}
	return strings.Join(items, ", ")
}

type samlEnvelope struct {
	RoleValues []string `xml:"Assertion>AttributeStatement>Attribute>AttributeValue"`
}
