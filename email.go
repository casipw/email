// Package email allows to send emails with attachments.
// Forked from https://github.com/scorredoira/email
package email

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"mime"
	"net/mail"
	"net/smtp"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Attachment represents an email attachment.
type Attachment struct {
	Filename string
	Data     []byte
	Inline   bool
}

// Message represents a smtp message.
type Message struct {
	From            mail.Address
	To              []string
	Cc              []string
	Bcc             []string
	ReplyTo         string
	Subject         string
	Body            string
	BodyContentType string
	Attachments     map[string]*Attachment
}

func (m *Message) attach(file string, inline bool) error {

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	_, filename := filepath.Split(file)

	m.Attachments[filename] = &Attachment{
		Filename: filename,
		Data:     data,
		Inline:   inline,
	}

	return nil
}

// AttachBuffer attaches a binary attachment.
func (m *Message) AttachBuffer(filename string, buf []byte, inline bool) error {
	m.Attachments[filename] = &Attachment{
		Filename: filename,
		Data:     buf,
		Inline:   inline,
	}
	return nil
}

// Attach attaches a file.
func (m *Message) Attach(file string) error {
	return m.attach(file, false)
}

// Inline includes a file as an inline attachment.
func (m *Message) Inline(file string) error {
	return m.attach(file, true)
}

func newMessage(subject string, body string, bodyContentType string) *Message {
	return &Message{
		Subject:         subject,
		Body:            body,
		BodyContentType: bodyContentType,
		Attachments:     make(map[string]*Attachment),
	}
}

// NewMessage returns a new Message that can compose an email with attachments
func NewMessage(subject string, body string) *Message {
	return newMessage(subject, body, "text/plain")
}

// NewHTMLMessage returns a new Message that can compose an HTML email with attachments
func NewHTMLMessage(subject string, body string) *Message {
	return newMessage(subject, body, "text/html")
}

// ToList returns all the recipients of the email
func (m *Message) ToList() []string {
	ToList := make([]string, len(m.To)+len(m.Cc)+len(m.Bcc))
	ToList = append(ToList, m.To...)
	ToList = append(ToList, m.Cc...)
	ToList = append(ToList, m.Bcc...)
	return ToList
}

func (m *Message) BytesLF() []byte {
	return m.bytes("\n")
}

func (m *Message) BytesCRLF() []byte {
	return m.bytes("\r\n")
}

// bytes assembles the mail.
func (m *Message) bytes(lineSeparator string) []byte {

	buf := bytes.NewBuffer(nil)

	buf.WriteString("From: " + m.From.String() + lineSeparator)

	t := time.Now()
	buf.WriteString("Date: " + t.Format(time.RFC1123Z) + lineSeparator)

	buf.WriteString("To: " + strings.Join(m.To, ",") + lineSeparator)
	if len(m.Cc) > 0 {
		buf.WriteString("Cc: " + strings.Join(m.Cc, ",") + lineSeparator)
	}

	buf.WriteString("Subject: ")
	encode(m.Subject, lineSeparator, buf)
	buf.WriteString(lineSeparator)

	if len(m.ReplyTo) > 0 {
		buf.WriteString("Reply-To: " + m.ReplyTo + lineSeparator)
	}

	buf.WriteString("MIME-Version: 1.0" + lineSeparator)

	boundary := "f46d043c813270fc6b04c2d223da"

	if len(m.Attachments) > 0 {
		buf.WriteString("Content-Type: multipart/mixed; boundary=\"" + boundary + "\"" + lineSeparator)
		buf.WriteString(lineSeparator + "--" + boundary + lineSeparator)
	}

	buf.WriteString("Content-Type: " + m.BodyContentType + "; charset=utf-8" + lineSeparator + lineSeparator)
	buf.WriteString(m.Body)
	buf.WriteString(lineSeparator)

	if len(m.Attachments) > 0 {

		for _, attachment := range m.Attachments {

			buf.WriteString(lineSeparator + lineSeparator + "--" + boundary + lineSeparator)

			if attachment.Inline {

				buf.WriteString("Content-Type: message/rfc822" + lineSeparator)
				buf.WriteString("Content-Disposition: inline; filename=\"")
				encode(attachment.Filename, lineSeparator, buf)
				buf.WriteString("\"" + lineSeparator + lineSeparator)

				buf.Write(attachment.Data)

			} else {

				contenttype := mime.TypeByExtension(filepath.Ext(attachment.Filename))
				if contenttype == "" {
					contenttype = "application/octet-stream"
				}

				buf.WriteString("Content-Type: " + contenttype + lineSeparator)
				buf.WriteString("Content-Transfer-Encoding: base64" + lineSeparator)
				buf.WriteString("Content-Disposition: attachment;" + lineSeparator + " filename=\"")
				encode(attachment.Filename, lineSeparator, buf)
				buf.WriteString("\"" + lineSeparator + lineSeparator)

				encoded := make([]byte, base64.StdEncoding.EncodedLen(len(attachment.Data)))
				base64.StdEncoding.Encode(encoded, attachment.Data)

				for len(encoded) > 60 {
					buf.Write(encoded[:76])
					buf.WriteString(lineSeparator)
					encoded = encoded[76:]
				}

				buf.Write(encoded)
			}

			buf.WriteString(lineSeparator + "--" + boundary)
		}

		buf.WriteString("--")
	}

	return buf.Bytes()
}

// Send sends the message via SMTP. Thus CRLF is used as line separator.
func Send(addr string, auth smtp.Auth, m *Message) error {
	return smtp.SendMail(addr, auth, m.From.Address, m.ToList(), m.BytesCRLF())
}

// Sendmail sends the message via the /usr/sbin/sendmail interface.
// It uses LF as line separator, which is suitable for most MTAs on Linux/Unix.
//
// qmail: "Unlike Sendmail, qmail requires locally-injected messages to use
// Unix newlines (LF only)."
//
// postfix: "The SMTP record delimiter is CRLF. Postfix removes it (as well as
// invalid CR characters at the end of a record) while receiving mail via SMTP,
// and adds it when sending mail via SMTP."
//
// sendmail: The system's line separator seems to be default, as the -bs flag
// changes that behavior.
func Sendmail(from string, m *Message) error {

	args := []string{"-i", "-f", from, "--"}
	args = append(args, m.ToList()...)

	sendmail := exec.Command("/usr/sbin/sendmail", args...)

	stdin, err := sendmail.StdinPipe()
	if err != nil {
		return err
	}

	sendmail.Start()
	stdin.Write(m.BytesLF())
	stdin.Close()

	err = sendmail.Wait()
	if err != nil {
		return err
	}

	return nil
}

// encode uses golang's mime.BEncoding in order to comply with RFC 2047:
// "An 'encoded-word' may not be more than 75 characters long, including
// 'charset', 'encoding', 'encoded-text', and delimiters. If it is desirable to
// encode more text than will fit in an 'encoded-word' of 75 characters,
// multiple 'encoded-word's (separated by CRLF SPACE) may be used.")
//
// mime.BEncoding separates the encoded words with spaces. We replace each of
// these spaces with line separator plus space, as stated by RFC 2047.
//
// If there the raw string is ASCII without special characters, it is split
// each 60 characters, using line separator plus space.
func encode(raw string, lineSeparator string, out *bytes.Buffer) {

	encoded := mime.BEncoding.Encode("UTF-8", raw)

	// mime.BEncoding: "If s is ASCII without special characters, it is returned unchanged."

	if strings.HasPrefix(encoded, "=?UTF-8?b?") {

		i := strings.Index(encoded, " ")

		for i >= 0 {
			out.WriteString(encoded[:i])
			out.WriteString(lineSeparator)
			out.WriteString(" ")
			encoded = encoded[i+1:]
			i = strings.Index(encoded, " ")
		}

		out.WriteString(encoded)

	} else {

		for len(encoded) > 60 {
			out.WriteString(encoded[:60])
			out.WriteString(lineSeparator)
			out.WriteString(" ")
			encoded = encoded[60:]
		}

		out.WriteString(encoded)
	}
}
