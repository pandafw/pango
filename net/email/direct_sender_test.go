package email

import (
	"crypto/tls"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pandafw/pango/iox"
	"github.com/pandafw/pango/net/netutil"
)

func testDirectSendEmail(t *testing.T, m *Email) {
	var err error

	s := &DirectSender{}
	s.Timeout = time.Second * 5
	s.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	f := os.Stdout
	w := iox.SyncWriter(f)
	s.ConnDebug = func(conn net.Conn) net.Conn {
		return netutil.DumpConn(
			conn,
			iox.WrapWriter(w, iox.ConsoleColor.Magenta+"< ", iox.ConsoleColor.Reset),
			iox.WrapWriter(w, iox.ConsoleColor.Yellow+"> ", iox.ConsoleColor.Reset),
		)
	}

	sf := os.Getenv("SMTP_FROM")
	if len(sf) < 1 {
		skipTest(t, "SMTP_FROM not set")
		return
	}

	st := os.Getenv("SMTP_TO")
	if len(st) < 1 {
		skipTest(t, "SMTP_TO not set")
		return
	}

	m.SetFrom(sf)
	m.AddTo(st)

	m.Subject = "direct send subject " + time.Now().String() + strings.Repeat(" 一二三四五", 10)
	err = s.DirectSend(m)
	if err != nil {
		t.Error(err)
	}
}

func TestDirectSendTextEmailOnly(t *testing.T) {
	email := &Email{}
	email.Message = ".\nthis is a test email " + time.Now().String() + " from example.com. 一二三四五"
	testDirectSendEmail(t, email)
}

func TestDirectSendTextEmailAttach(t *testing.T) {
	email := &Email{}

	email.Message = ".\nthis is a test email " + time.Now().String() + " from example.com. 一二三四五"
	email.AttachString("string.txt", "abcdefg")
	err := email.EmbedFile("panda.png", "testdata/panda.png")
	if err != nil {
		t.Error(err)
		return
	}

	testDirectSendEmail(t, email)
}

func TestDirectSendHtmlEmailOnly(t *testing.T) {
	email := &Email{}
	email.SetHTMLMsg("<pre><font color=red>.\nthis is a test email " + time.Now().String() + " from example.com. 一二三四五</font></pre>")

	testDirectSendEmail(t, email)
}

func TestDirectSendHtmlEmailAttach(t *testing.T) {
	email := &Email{}

	email.SetHTMLMsg("<pre><IMG src=\"cid:panda.png\"> <font color=red>.\nthis is a test email " + time.Now().String() + " from example.com. 一二三四五</font></pre>")
	email.AttachString("test.txt", "abcdefg")
	err := email.EmbedFile("panda.png", "testdata/panda.png")
	if err != nil {
		t.Error(err)
		return
	}

	testDirectSendEmail(t, email)
}
