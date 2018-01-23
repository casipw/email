package email

import (
	"net/mail"
	"testing"
	"time"
)

func TestEncoding(t *testing.T) {

	message := NewMessage("Subject", "Body")
	message.From = mail.Address{Name: "From", Address: "from@example.com"}
	message.To = []string{"to@example.com"}

	message.AttachBuffer(
		"äöü1234567890123456789012345678901234567890123456789012345678901234567890.txt",
		[]byte("Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam"),
		false,
	)

	output := string(message.BytesLF())

	expected :=
		`From: "From" <from@example.com>
Date: ` + time.Now().Format(time.RFC1123Z) + `
To: to@example.com
Subject: Subject
MIME-Version: 1.0
Content-Type: multipart/mixed; boundary="f46d043c813270fc6b04c2d223da"

--f46d043c813270fc6b04c2d223da
Content-Type: text/plain; charset=utf-8

Body


--f46d043c813270fc6b04c2d223da
Content-Type: text/plain; charset=utf-8
Content-Transfer-Encoding: base64
Content-Disposition: attachment;
 filename="=?UTF-8?b?w6TDtsO8MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5?=
 =?UTF-8?b?MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MC50eHQ=?="

TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQsIGNvbnNldGV0dXIgc2FkaXBzY2luZyBlbGl0ciwg
c2VkIGRpYW0gbm9udW15IGVpcm1vZCB0ZW1wb3IgaW52aWR1bnQgdXQgbGFib3JlIGV0IGRvbG9y
ZSBtYWduYSBhbGlxdXlhbQ==
--f46d043c813270fc6b04c2d223da--`

	if output != expected {
		t.Fatal("TestEncoding failed")
	}
}
