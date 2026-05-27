package rhsaml

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type samlForm struct {
	AWSURL        string
	SAMLAssertion string
}

func parseSAMLForm(idpURL, html string) (samlForm, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return samlForm{}, fmt.Errorf("parse SAML IDP page: %w", err)
	}
	form := doc.Find("form").First()
	if form.Length() == 0 {
		return samlForm{}, fmt.Errorf("SAML form not found in IDP response")
	}
	action, ok := form.Attr("action")
	if !ok || strings.TrimSpace(action) == "" {
		return samlForm{}, fmt.Errorf("SAML form action missing")
	}
	assertion, ok := form.Find("input[type='hidden']").First().Attr("value")
	if !ok || strings.TrimSpace(assertion) == "" {
		return samlForm{}, fmt.Errorf("SAML assertion missing")
	}
	awsURL, err := resolveActionURL(idpURL, action)
	if err != nil {
		return samlForm{}, err
	}
	return samlForm{
		AWSURL:        awsURL,
		SAMLAssertion: strings.TrimSpace(assertion),
	}, nil
}

func resolveActionURL(base, action string) (string, error) {
	actionURL, err := url.Parse(strings.TrimSpace(action))
	if err != nil {
		return "", fmt.Errorf("parse SAML form action URL: %w", err)
	}
	if actionURL.IsAbs() {
		return actionURL.String(), nil
	}
	baseURL, err := url.Parse(strings.TrimSpace(base))
	if err != nil {
		return "", fmt.Errorf("parse SAML IDP URL: %w", err)
	}
	return baseURL.ResolveReference(actionURL).String(), nil
}
