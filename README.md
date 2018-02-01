# email [![GoDoc](https://godoc.org/github.com/casipw/email?status.svg)](http://godoc.org/github.com/casipw/email) [![Report card](https://goreportcard.com/badge/github.com/casipw/email)](https://goreportcard.com/report/github.com/casipw/email)

An easy way to send emails with attachments in Go.

Forked from `github.com/scorredoira/email`, with these modifications:

* RFC-2047-compliant splitting of attachment filenames and subject
* Can send mail via `/usr/sbin/sendmail`

# Install

```bash
go get github.com/casipw/email
```

# Usage

```go
package main

import (
    "github.com/casipw/email"
    "net/mail"
    "net/smtp"
)

func main() {
    m := email.NewMessage("Subject", "Body")
    m.From = mail.Address{Name: "Sender", Address: "from@example.com"}
    m.To = []string{"to@example.com"}
    if err := m.Attach("file.pdf"); err != nil {
        panic(err)
    }

    // pass it to sendmail
    email.Sendmail("sender@example.com", m)

    // or send it via SMTP
    auth := smtp.PlainAuth("", "from@example.com", "mypassword", "smtp.example.com")
    if err := email.Send("smtp.example.com:587", auth, m); err != nil {
        panic(err)
    }
}

```

# Html

```go
// use the html constructor
m := email.NewHTMLMessage("Hi", "this is the body")
```

# Inline

```go
// use Inline to display the attachment inline.
if err := m.Inline("main.go"); err != nil {
    log.Fatal(err)
}
```
