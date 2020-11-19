package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

// DirectSender direct send email
type DirectSender struct {
	// LocalName is the hostname sent to the SMTP server with the HELO command.
	// By default, "localhost" is sent.
	Helo string

	// Timeout timeout when connect to the SMTP server
	Timeout time.Duration

	// SSL defines whether an SSL connection is used. It should be false in
	// most cases since the authentication mechanism should use the STARTTLS
	// extension instead.
	SSL bool

	// SkipTLS Skip StartTLS when the STARTTLS extension is used
	SkipTLS bool

	// InsecureSkipVerify controls whether a client verifies the server's
	// certificate chain and host name. If InsecureSkipVerify is true, crypto/tls
	// accepts any certificate presented by the server and any host name in that
	// certificate. In this mode, TLS is susceptible to machine-in-the-middle
	// attacks unless custom verification is used. This should be used only for
	// testing or in combination with VerifyConnection or VerifyPeerCertificate.
	InsecureSkipVerify bool

	sender Sender
}

// DirectSend lookup MX record and direct send to recipter's mail server
func (ds *DirectSender) DirectSend(ms ...*Email) error {
	ds.sender.Helo = ds.Helo
	ds.sender.SSL = ds.SSL
	ds.sender.SkipTLS = ds.SkipTLS
	ds.sender.TLSConfig = &tls.Config{InsecureSkipVerify: ds.InsecureSkipVerify}
	ds.sender.Port = 25

	for i, m := range ms {
		if err := ds.directSend(m); err != nil {
			return fmt.Errorf("Failed to direct send email %d: %v", i+1, err)
		}
	}

	return nil
}

func (ds *DirectSender) directSend(mail *Email) error {
	rm := mail.GetRcptsByDomain()
	for domain, addrs := range rm {
		err := ds.directSends(domain, addrs, mail)
		if err != nil {
			return fmt.Errorf("Failed to direct send email to %s: %v", domain, err)
		}
	}

	return nil
}

func (ds *DirectSender) directSends(domain string, addrs []string, mail *Email) error {
	mxrs, err := net.LookupMX(domain)
	if err != nil {
		return err
	}

	for _, mx := range mxrs {
		if mx.Host == "" {
			continue
		}

		ds.sender.Host = mx.Host
		if ds.sender.Host[len(ds.sender.Host)-1] == '.' {
			ds.sender.Host = ds.sender.Host[:len(ds.sender.Host)-1]
		}
		ds.sender.TLSConfig.ServerName = ds.sender.Host

		err = ds.sender.Dial()
		if err == nil {
			err = ds.sender.Send(mail)
			ds.sender.Close()
			break
		}
	}

	return err
}
