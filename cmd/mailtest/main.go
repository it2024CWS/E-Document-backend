// Standalone SMTP smoke test.
// Run:  go run ./cmd/mailtest
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"e-document-backend/internal/config"
	"e-document-backend/internal/pkg/mailer"
)

func main() {
	cfg := config.Load()
	if cfg.SMTP.Host == "" {
		fmt.Fprintln(os.Stderr, "SMTP_HOST is empty — nothing to test.")
		os.Exit(1)
	}

	to := "hackter.s@gmail.com"
	if len(os.Args) > 1 {
		to = os.Args[1]
	}

	m := mailer.New(cfg.SMTP)
	html := fmt.Sprintf(`<html><body style="font-family: 'Noto Sans Lao', sans-serif;">
	<h2>E-Document — ທົດສອບການສົ່ງອີເມລ</h2>
	<p>ອີເມລນີ້ຖືກສົ່ງຈາກ mailtest CLI ເພື່ອຢືນຢັນວ່າການຕັ້ງຄ່າ SMTP ໃຊ້ໄດ້.</p>
	<ul>
	  <li>SMTP host: %s:%s</li>
	  <li>From: %s &lt;%s&gt;</li>
	  <li>ເວລາ: %s</li>
	</ul>
	<p style="color:#888;font-size:12px;">ຖ້າທ່ານໄດ້ຮັບອີເມລນີ້ ຫມາຍຄວາມວ່າລະບົບແຈ້ງເຕືອນເອກະສານພ້ອມໃຊ້ງານແລ້ວ.</p>
	</body></html>`,
		cfg.SMTP.Host, cfg.SMTP.Port,
		cfg.SMTP.FromName, cfg.SMTP.FromEmail,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	msg := mailer.Message{
		To:      []string{to},
		Subject: "[E-Document] ທົດສອບການສົ່ງອີເມລ",
		HTML:    html,
	}

	fmt.Printf("Sending test email to %s via %s:%s ...\n", to, cfg.SMTP.Host, cfg.SMTP.Port)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := m.Send(ctx, msg); err != nil {
		fmt.Fprintf(os.Stderr, "FAILED: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("OK — email dispatched.")
}
