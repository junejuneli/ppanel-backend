package email

import (
	"fmt"
	"strings"

	"github.com/perfect-panel/server/internal/infra/mail"
	"golang.org/x/net/html"
)

// renderVerificationBody preserves working custom templates, but falls back to
// the bundled template rather than sending a signature-only/empty notification.
// Invalid payloads are not sent. Never include their values in errors or logs.
func renderVerificationBody(configured string, data map[string]any) (body string, fallback bool, err error) {
	code, ok := data["Code"].(string)
	if !ok || strings.TrimSpace(code) == "" {
		return "", false, fmt.Errorf("verification payload lacks code")
	}
	purpose, ok := data["Type"].(uint8)
	if !ok || (purpose != 1 && purpose != 2) {
		return "", false, fmt.Errorf("verification payload has invalid purpose")
	}
	expiry, ok := data["Expire"].(float64)
	if !ok || expiry <= 0 {
		return "", false, fmt.Errorf("verification payload has invalid expiry")
	}
	body, err = renderEmailTemplate("verification", mail.UpgradeDefaultTemplate(configured), data)
	if err == nil && verificationBodyContainsCode(body, code) {
		return body, false, nil
	}
	body, err = renderEmailTemplate("verification", mail.DefaultEmailVerifyTemplate, data)
	if err != nil || !verificationBodyContainsCode(body, code) {
		return "", true, fmt.Errorf("default verification template failed validation")
	}
	return body, true, nil
}

// Look in HTML text, not comments, URLs, metadata or scripts. This checks content
// presence, not pixel visibility under every possible email-client CSS engine.
func verificationBodyContainsCode(body, code string) bool {
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return false
	}
	var text strings.Builder
	var visit func(*html.Node)
	visit = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "head", "script", "style", "template":
				return
			}
			for _, a := range n.Attr {
				if a.Key == "hidden" {
					return
				}
			}
		}
		if n.Type == html.TextNode {
			text.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}
	visit(doc)
	return strings.Contains(text.String(), code)
}
