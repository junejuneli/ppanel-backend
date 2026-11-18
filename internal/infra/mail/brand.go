package mail

import (
	_ "embed"
	"strings"

	"golang.org/x/net/html"
)

const BrandName = "GoodVPN"
const BrandLogoURL = "https://goodvpn.app/pwa-192x192.png"

const brandLogoCID = "goodvpn-logo.png"

//go:embed goodvpn-logo.png
var brandLogo []byte

const brandHeader = `<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="background:#f4f1fb;font-family:Arial,sans-serif;">
<tr><td align="center" style="padding:28px 16px;">
<!--[if mso]><table role="presentation" width="560" cellpadding="0" cellspacing="0" border="0"><tr><td><![endif]-->
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="max-width:560px;">
<tr><td bgcolor="#FFFFFF" style="padding:19px 25px;border-top:4px solid #6748e5;border-radius:14px 14px 0 0;border-bottom:1px solid #eee9f6;">
<a href="https://goodvpn.app/" style="color:#242137;font-size:20px;font-weight:700;text-decoration:none;"><img src="cid:` + brandLogoCID + `" width="132" height="28" alt="GoodVPN" style="display:block;border:0;color:#242137;font-size:18px;font-weight:700;"></a>
</td></tr><tr><td bgcolor="#FFFFFF" style="padding:28px 25px;color:#242137;font-size:15px;line-height:1.65;overflow-wrap:anywhere;">`

const brandFooter = `</td></tr>
<tr><td bgcolor="#FBF9FF" style="padding:16px 25px;border-top:1px solid #eee9f6;border-radius:0 0 14px 14px;color:#575368;font-size:12px;line-height:1.6;">
<strong style="color:#242137;">GoodVPN</strong>&nbsp;&nbsp;·&nbsp;&nbsp;<a href="https://goodvpn.app/" style="color:#5433c5;text-decoration:underline;">goodvpn.app</a>
</td></tr></table>
<!--[if mso]></td></tr></table><![endif]-->
</td></tr></table>`

// brandHTML wraps fragments and inserts the layout inside existing HTML
// documents so saved campaign styles and body attributes survive.
func brandHTML(body string) string {
	tokens := html.NewTokenizer(strings.NewReader(body))
	offset, bodyStart, bodyEnd, headEnd := 0, -1, -1, -1
	for {
		kind := tokens.Next()
		if kind == html.ErrorToken {
			break
		}
		start := offset
		offset += len(tokens.Raw())
		if kind != html.StartTagToken && kind != html.EndTagToken {
			continue
		}
		name, _ := tokens.TagName()
		switch {
		case string(name) == "body" && kind == html.StartTagToken && bodyStart == -1:
			bodyStart = offset
		case string(name) == "body" && kind == html.EndTagToken && bodyStart != -1 && bodyEnd == -1:
			bodyEnd = start
		case string(name) == "head" && kind == html.EndTagToken:
			headEnd = offset
		case string(name) == "html" && kind == html.EndTagToken && bodyEnd == -1:
			bodyEnd = start
		}
	}
	if bodyEnd == -1 {
		bodyEnd = len(body)
	}
	if bodyStart != -1 && bodyStart <= bodyEnd {
		return body[:bodyStart] + brandHeader + body[bodyStart:bodyEnd] + brandFooter + body[bodyEnd:]
	}
	const openBody = `<body style="margin:0;background:#f4f1fb;">`
	if headEnd != -1 && headEnd <= bodyEnd {
		return body[:headEnd] + openBody + brandHeader + body[headEnd:bodyEnd] + brandFooter + "</body>" + body[bodyEnd:]
	}
	return `<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>GoodVPN</title></head>` +
		openBody + brandHeader + body + brandFooter + "</body></html>"
}
