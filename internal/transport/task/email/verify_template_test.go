package email

import (
	"strings"
	"testing"
)

func TestVerificationTemplateSafetyFallback(t *testing.T) {
	data := map[string]any{"Type": uint8(1), "Expire": float64(5), "Code": "481516"}
	for name, template := range map[string]string{
		"empty": "", "signature only": "<p>NewRouters</p>", "malformed": "{{.Code", "wrong field": "<p>{{.Token}}</p>",
		"comment only": "<!-- {{.Code}} -->", "attribute only": "<a href='/{{.Code}}'>link</a>",
		"script only": "<script>{{.Code}}</script>", "hidden element": "<p hidden>{{.Code}}</p>",
	} {
		t.Run(name, func(t *testing.T) {
			body, fallback, err := renderVerificationBody(template, data)
			if err != nil || !fallback || !strings.Contains(body, "481516") || !strings.Contains(body, "Registration Verification Code") {
				t.Fatalf("fallback=%v err=%v", fallback, err)
			}
		})
	}
}

func TestVerificationPreservesValidCustomTemplate(t *testing.T) {
	for _, purpose := range []uint8{1, 2} {
		const custom = "<h1>Custom brand</h1><p>{{.Code}}</p><p>Expires {{.Expire}} min</p>"
		body, fallback, err := renderVerificationBody(custom, map[string]any{"Type": purpose, "Expire": float64(5), "Code": "481516"})
		if err != nil || fallback || body != "<h1>Custom brand</h1><p>481516</p><p>Expires 5 min</p>" {
			t.Fatalf("valid template changed, fallback=%v err=%v", fallback, err)
		}
	}
}

func TestVerificationRejectsMalformedPayload(t *testing.T) {
	for name, value := range map[string]any{"missing": nil, "empty": "", "whitespace": "  ", "numeric": float64(481516)} {
		t.Run(name, func(t *testing.T) {
			body, _, err := renderVerificationBody("<p>{{.Code}}</p>", map[string]any{"Type": uint8(1), "Expire": float64(5), "Code": value})
			if err == nil || body != "" {
				t.Fatal("invalid code was accepted")
			}
		})
	}
	for name, data := range map[string]map[string]any{
		"bad purpose":    {"Type": uint8(9), "Expire": float64(5), "Code": "481516"},
		"missing expiry": {"Type": uint8(1), "Code": "481516"},
		"expired":        {"Type": uint8(1), "Expire": float64(0), "Code": "481516"},
	} {
		t.Run(name, func(t *testing.T) {
			body, _, err := renderVerificationBody("{{.Code}}", data)
			if err == nil || body != "" {
				t.Fatal("invalid payload accepted")
			}
			if strings.Contains(err.Error(), "481516") {
				t.Fatal("error leaked code")
			}
		})
	}
}

func TestVerificationResetFallbackExplainsPurpose(t *testing.T) {
	body, fallback, err := renderVerificationBody("<p>NewRouters</p>", map[string]any{"Type": uint8(2), "Expire": float64(5), "Code": "481516"})
	if err != nil || !fallback || !strings.Contains(body, "Password Reset Verification Code") || !strings.Contains(body, "481516") {
		t.Fatalf("reset fallback failed: fallback=%v err=%v", fallback, err)
	}
}
