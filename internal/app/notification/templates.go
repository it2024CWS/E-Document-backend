package notification

import (
	"bytes"
	"html/template"
)

// eventKind identifies which of the five notification templates to use.
type eventKind int

const (
	eventIncomingReceived eventKind = iota
	eventIncomingApproved
	eventIncomingRejected
	eventOwnerApproved
	eventOwnerRejected
)

type eventCopy struct {
	SubjectLao    string // short title used in subject line
	SentenceLao   string // main verb phrase — completes the sentence "ເອກະສານຂອງທ່ານ ... ."
	ActorLabelLao string // label for the actor row (ຜູ້ຮັບ / ຜູ້ອະນຸມັດ / ຜູ້ປະຕິເສດ)
}

var copyByEvent = map[eventKind]eventCopy{
	eventIncomingReceived: {
		SubjectLao:    "ຮັບແລ້ວ",
		SentenceLao:   "ຖືກຮັບແລ້ວໂດຍພະແນກປາຍທາງ",
		ActorLabelLao: "ຜູ້ຮັບ",
	},
	eventIncomingApproved: {
		SubjectLao:    "ອະນຸມັດແລ້ວ",
		SentenceLao:   "ຖືກອະນຸມັດຢູ່ພະແນກປາຍທາງ",
		ActorLabelLao: "ຜູ້ອະນຸມັດ",
	},
	eventIncomingRejected: {
		SubjectLao:    "ຖືກປະຕິເສດ",
		SentenceLao:   "ຖືກປະຕິເສດຢູ່ພະແນກປາຍທາງ",
		ActorLabelLao: "ຜູ້ປະຕິເສດ",
	},
	eventOwnerApproved: {
		SubjectLao:    "ຫົວໜ້າແຜກອະນຸມັດການສົ່ງ",
		SentenceLao:   "ຖືກຫົວໜ້າພະແນກອະນຸມັດ ແລະ ເລີ່ມສົ່ງອອກ",
		ActorLabelLao: "ຜູ້ອະນຸມັດ",
	},
	eventOwnerRejected: {
		SubjectLao:    "ຫົວໜ້າແຜກປະຕິເສດການສົ່ງ",
		SentenceLao:   "ຖືກຫົວໜ້າພະແນກປະຕິເສດ ບໍ່ໄດ້ສົ່ງອອກ",
		ActorLabelLao: "ຜູ້ປະຕິເສດ",
	},
}

// templateData is the model passed to the HTML template.
type templateData struct {
	OwnerName     string
	DocNo         string
	DocName       string
	SentenceLao   string
	DeptName      string
	ActorLabelLao string
	ActorName     string
	Remark        string
	Timestamp     string
}

const bodyTemplateHTML = `<!DOCTYPE html>
<html lang="lo">
<head><meta charset="UTF-8"></head>
<body style="font-family: 'Phetsarath OT', 'Noto Sans Lao', sans-serif; color:#222; line-height:1.6;">
  <p>ຮຽນ ທ່ານ {{.OwnerName}},</p>
  <p>ເອກະສານຂອງທ່ານ (ເລກທີ <b>{{.DocNo}}</b>{{if .DocName}}{{end}}) {{.SentenceLao}}.</p>
  <table cellpadding="4" cellspacing="0" style="border-collapse:collapse;">
    <tr><td><b>ພະແນກ</b></td><td>{{.DeptName}}</td></tr>
    {{if and .ActorLabelLao .ActorName}}<tr><td><b>{{.ActorLabelLao}}</b></td><td>{{.ActorName}}</td></tr>{{end}}
    {{if .Remark}}<tr><td><b>ໝາຍເຫດ</b></td><td>{{.Remark}}</td></tr>{{end}}
    <tr><td><b>ວັນທີ</b></td><td>{{.Timestamp}}</td></tr>
  </table>
  <p style="color:#888;font-size:12px;margin-top:24px;">ອີເມລນີ້ຖືກສົ່ງອັດຕະໂນມັດຈາກລະບົບ E-Document. ກະລຸນາຢ່າຕອບກັບ.</p>
</body>
</html>`

var bodyTemplate = template.Must(template.New("notification").Parse(bodyTemplateHTML))

// render returns (subject, htmlBody) for the given event and data.
func render(kind eventKind, data templateData) (string, string, error) {
	c := copyByEvent[kind]
	data.SentenceLao = c.SentenceLao
	data.ActorLabelLao = c.ActorLabelLao

	var buf bytes.Buffer
	if err := bodyTemplate.Execute(&buf, data); err != nil {
		return "", "", err
	}

	subject := "[E-Office] ເອກະສານ " + data.DocNo + "  " + c.SubjectLao
	return subject, buf.String(), nil
}
