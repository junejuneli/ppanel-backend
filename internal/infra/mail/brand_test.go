package mail

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
	"testing"
)

func TestBrandedSenderPreservesCampaignAndEmbedsLogo(t *testing.T) {
	sender, err := NewSender("smtp", `{"from":"noreply@example.test","reply_to":"support@example.test"}`, "Old brand")
	if err != nil {
		t.Fatal(err)
	}
	const campaign = `<!doctype html><html lang="en"><head><style>.offer{color:purple}</style></head><BODY class="campaign"><p class="offer">Member offer</p><a href="https://example.test/unsubscribe">Unsubscribe</a></BODY></html>`
	client := sender.(*SMTPClient)
	var wire bytes.Buffer
	if _, err := client.message([]string{"reader@example.test"}, "Offer", campaign).WriteTo(&wire); err != nil {
		t.Fatal(err)
	}
	message, err := mail.ReadMessage(&wire)
	if err != nil {
		t.Fatal(err)
	}
	from, err := mail.ParseAddress(message.Header.Get("From"))
	if err != nil || from.Name != BrandName || from.Address != "noreply@example.test" {
		t.Fatalf("From = %v, err = %v", from, err)
	}
	if message.Header.Get("Reply-To") != "support@example.test" || message.Header.Get("Subject") != "Offer" {
		t.Fatal("the sender changed the reply address or subject")
	}
	contentType, params, err := mime.ParseMediaType(message.Header.Get("Content-Type"))
	if err != nil || contentType != "multipart/related" {
		t.Fatalf("Content-Type = %q, err = %v", contentType, err)
	}
	parts := multipart.NewReader(message.Body, params["boundary"])
	htmlPart, err := parts.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(htmlPart)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`<style>.offer{color:purple}</style>`,
		`<BODY class="campaign"><table`,
		`cid:goodvpn-logo.png`,
		`<p class="offer">Member offer</p>`,
		`https://example.test/unsubscribe`,
		`</table></BODY>`,
	} {
		if !strings.Contains(string(body), want) {
			t.Errorf("body lacks %q", want)
		}
	}
	logoPart, err := parts.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	if logoPart.Header.Get("Content-ID") != "<goodvpn-logo.png>" ||
		!strings.HasPrefix(logoPart.Header.Get("Content-Type"), "image/png") {
		t.Fatalf("unexpected logo MIME headers: %v", logoPart.Header)
	}
}
