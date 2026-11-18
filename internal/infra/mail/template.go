package mail

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

const (
	DefaultEmailVerifySubject        = "Verification code"
	DefaultExpirationEmailSubject    = "Subscription Expired"
	DefaultMaintenanceEmailSubject   = "系统维护通知 / System Maintenance Notice"
	DefaultTrafficExceedEmailSubject = "Subscription Traffic Exceed"
)

const heading = `<h1 style="margin:0 0 10px;color:#242137;font-size:27px;line-height:1.25;font-weight:700;">`
const subtitle = `<p style="margin:0 0 24px;color:#5f4caa;font-size:14px;line-height:1.5;font-weight:600;">`
const paragraph = `<p style="margin:0 0 14px;">`
const english = `<p lang="en" style="margin:0 0 18px;color:#575368;">`
const detail = `<p style="margin:20px 0;padding:15px 18px;background:#f2edff;border-left:3px solid #6748e5;color:#322466;font-weight:700;">`
const action = `<p style="margin:24px 0 0;"><a href="https://goodvpn.app/#/purchasing" style="display:inline-block;padding:12px 20px;border-radius:9px;background:#6748e5;color:#ffffff;text-decoration:none;font-weight:700;">查看套餐 · View plans</a></p>`

// Templates are bilingual HTML fragments. The sender adds the GoodVPN layout.
const (
	DefaultEmailVerifyTemplate = heading + `{{if eq .Type 1}}注册验证码{{else}}重置密码验证码{{end}}</h1>` +
		subtitle + `{{if eq .Type 1}}Registration Verification Code{{else}}Password Reset Verification Code{{end}}</p>` +
		paragraph + `{{if eq .Type 1}}请输入下方验证码完成注册。{{else}}请输入下方验证码重置密码。{{end}}验证码将在 <strong>{{.Expire}}</strong> 分钟后失效。</p>` +
		english + `Enter the code below to {{if eq .Type 1}}complete your registration{{else}}reset your password{{end}}. It expires in <span>{{.Expire}}</span> minutes.</p>
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="margin:22px 0;">
<tr><td align="center" bgcolor="#F2EDFF" style="padding:20px 12px;border:1px solid #E3DAFF;border-radius:12px;color:#4a2ab9;font-family:Consolas,monospace;font-size:32px;font-weight:700;letter-spacing:0.16em;">{{.Code}}</td></tr>
</table>` +
		`<p style="margin:0;color:#575368;font-size:13px;">如果您没有请求验证码，可以忽略此邮件。请勿与他人分享验证码。<br><span lang="en">If you didn’t request this code, you can ignore this email. Keep the code private.</span></p>`

	DefaultMaintenanceEmailTemplate = heading + `系统维护通知</h1>` +
		subtitle + `Scheduled Maintenance</p>` +
		paragraph + `GoodVPN 将在以下时段进行系统维护。维护期间，服务可能暂时中断。</p>` +
		detail + `{{.MaintenanceDate}}<br>{{.MaintenanceTime}}</p>` +
		english + `GoodVPN may be temporarily unavailable during the maintenance window above. Service will resume once maintenance is complete.</p>`

	DefaultExpirationEmailTemplate = heading + `订阅到期提醒</h1>` +
		subtitle + `Subscription Expiration</p>` +
		paragraph + `以下是您的 GoodVPN 订阅到期时间。如需继续使用，请续订。</p>` +
		detail + `到期时间 · Expiration<br>{{.ExpireDate}}</p>` +
		english + `Your GoodVPN subscription ends on the date above. Renew to continue using the service.</p>` + action

	DefaultTrafficExceedEmailTemplate = heading + `套餐流量已用尽</h1>` +
		subtitle + `Data Allowance Used</p>` +
		paragraph + `您已用完套餐内的流量。如需恢复连接，请购买流量或更换套餐。</p>` +
		english + `You’ve used the data included in your plan. To reconnect, purchase more data or change your plan.</p>` + action
)

// UpgradeDefaultTemplate replaces exact copies of the bundled legacy
// templates, including copies saved in the database. Edits stay in place.
func UpgradeDefaultTemplate(body string) string {
	switch fmt.Sprintf("%x", sha256.Sum256([]byte(strings.TrimSpace(body)))) {
	case "0bac45706096cf7d4ebbe655af6145d5d246c03ce2df42f8f5dc70f11e76c35c":
		return DefaultEmailVerifyTemplate
	case "8dc7cef2305c31bbb1b2f1ea36b6429f3e63438a30a0418d5d009d268f40a416":
		return DefaultMaintenanceEmailTemplate
	case "352bf82199d250a6131dc27515fc5b04b4411724d1e2030f41d8acc9127a802e":
		return DefaultExpirationEmailTemplate
	case "d9a9189df5ba1e667bda502dc516cb07c35bc80845312f21637d353fe60a8d54":
		return DefaultTrafficExceedEmailTemplate
	default:
		return body
	}
}
