package authmethodadmin

import (
	"context"
	"fmt"

	"github.com/perfect-panel/server/internal/infra/mail"
	dto "github.com/perfect-panel/server/internal/module/identity/contract"
	"github.com/perfect-panel/server/pkg/xerr"
)

// TestEmailSend sends a test email through the configured email sender.
func (s *Service) TestEmailSend(ctx context.Context, req *dto.TestEmailSendRequest) error {
	cfg := s.deps.Config()
	client, err := mail.NewSender(cfg.EmailPlatform, cfg.EmailPlatformConfig, cfg.SiteName)
	if err != nil {
		return xerr.Wrapf(err, xerr.ERROR, "new email sender")
	}
	err = client.SendContext(ctx, []string{req.Email}, "GoodVPN 邮件发送测试 / Email delivery test", "<h1>邮件发送测试</h1><p>您已成功收到通过 GoodVPN 邮件服务发送的测试邮件。</p><p lang=\"en\">Your GoodVPN email delivery test succeeded.</p>")
	if err != nil {
		// The administrator is testing the sender, so the failure itself is
		// the answer.
		return fmt.Errorf("send test email: %w", xerr.NewErrCodeMsg(xerr.SenderTestFailed, fmt.Sprintf("send email err: %v", err.Error())))
	}
	return nil
}
