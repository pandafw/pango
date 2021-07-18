package email

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/pandafw/pango/iox"
	"github.com/pandafw/pango/str"
)

// Sender email sender
type Sender struct {
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

	// TSLConfig represents the TLS configuration used for the TLS (when the
	// STARTTLS extension is used) or SSL connection.
	TLSConfig *tls.Config

	// ConnDebug  a conn wrap func
	ConnDebug func(conn net.Conn) net.Conn

	// DataDebug  a data writer wrap func
	DataDebug func(w io.Writer) io.Writer

	client *smtp.Client
}

// IsDialed return true if the sender is dialed
func (s *Sender) IsDialed() bool {
	return s.client != nil
}

// Close close the SMTP client
func (s *Sender) Close() error {
	if s.client == nil {
		return nil
	}

	err := s.client.Quit()
	s.client = nil
	return err
}

func (s *Sender) setConn(i interface{}, c net.Conn) {
	v := reflect.ValueOf(i).Elem()
	f := v.FieldByName("conn")
	pc := (*net.Conn)(unsafe.Pointer(f.UnsafeAddr()))
	*pc = c
}

func (s *Sender) getConn(i interface{}) net.Conn {
	v := reflect.ValueOf(i).Elem()
	f := v.FieldByName("conn")
	pc := (*net.Conn)(unsafe.Pointer(f.UnsafeAddr()))
	return *pc
}

func (s *Sender) tlsConfig(host string) *tls.Config {
	if s.TLSConfig == nil {
		s.TLSConfig = &tls.Config{ServerName: host}
	}
	return s.TLSConfig
}

func (s *Sender) dial(host string, port int) error {
	addr := host + ":" + strconv.Itoa(port)
	conn, err := net.DialTimeout("tcp", addr, s.Timeout)
	if err != nil {
		return fmt.Errorf("Failed to dial %s - %v", addr, err)
	}

	if s.SSL {
		conn = tls.Client(conn, s.tlsConfig(host))
	}

	corg := conn
	if s.ConnDebug != nil {
		conn = s.ConnDebug(conn)
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	if s.Helo != "" {
		if err := c.Hello(s.Helo); err != nil {
			c.Close()
			return err
		}
	}

	if !s.SSL && !s.SkipTLS {
		if s.ConnDebug != nil {
			s.setConn(c, corg)
			c.Text = textproto.NewConn(corg)
		}
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(s.tlsConfig(host)); err != nil {
				c.Close()
				return err
			}
		}
		if s.ConnDebug != nil {
			ctls := s.getConn(c)
			conn = s.ConnDebug(ctls)
			c.Text = textproto.NewConn(conn)
		}
	}

	s.client = c
	return nil
}

func (s *Sender) send(mail *Email) error {
	c := s.client
	if err := c.Mail(mail.GetSender()); err != nil {
		return err
	}

	for _, addr := range mail.GetTos() {
		if err := c.Rcpt(addr.Address); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	err = s.writeMail(w, mail)
	if err != nil {
		w.Close()
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *Sender) writeMail(w io.Writer, m *Email) error {
	header := header{}

	header["MIME-Version"] = "1.0"
	if m.MsgID != "" {
		header["Message-ID"] = m.MsgID
	}
	header["Date"] = formatDate(m.GetDate())
	header["From"] = m.GetFrom().String()
	header["To"] = encodeAddress(m.GetTos()...)
	if len(m.GetCcs()) > 0 {
		header["Cc"] = encodeAddress(m.GetCcs()...)
	}
	if len(m.GetBccs()) > 0 {
		header["Bcc"] = encodeAddress(m.GetBccs()...)
	}
	if len(m.GetReplys()) > 0 {
		header["Reply-To"] = encodeAddress(m.GetReplys()...)
	}
	header["Subject"] = encodeWord(m.Subject)

	var boundary string
	if m.HTML || len(m.Attachments) > 0 {
		boundary = str.RandLetterNumbers(28)
		header["Content-Type"] = "multipart/mixed; boundary=" + boundary
	} else {
		header["Content-Type"] = "text/plain; charset=UTF-8"
	}

	enc := "7bit"
	if !str.IsASCII(m.Message) {
		enc = "Base64"
	}
	header["Content-Transfer-Encoding"] = enc

	if s.DataDebug != nil {
		w = s.DataDebug(w)
	}

	err := writeHeader(w, header)
	if err != nil {
		return err
	}

	return writeBody(w, enc, m, boundary)
}

// header SMTP message header
type header map[string]string

func formatDate(t time.Time) string {
	return t.Format(time.RFC1123Z)
}

// Encode a RFC 822 "word" token into mail-safe form as per RFC 2047.
func encodeWord(s string) string {
	return mime.QEncoding.Encode("UTF-8", s)
}

func encodeBody(s string) string {
	sb := strings.Builder{}
	b := base64.NewEncoder(base64.StdEncoding, iox.NewMimeChunkWriter(&sb))
	b.Write([]byte(s))
	b.Close()
	return sb.String()
}

func encodeAddress(as ...*mail.Address) string {
	sb := strings.Builder{}
	for i, a := range as {
		sb.WriteString(a.String())
		if i < len(as)-1 {
			sb.WriteString(", ")
		}
	}
	return sb.String()
}

// http://www.faqs.org/rfcs/rfc2822.html
func writeHeader(w io.Writer, h header) error {
	var err error
	for k, v := range h {
		err = writeString(w, k)
		err = writeString(w, ": ")
		if len(k)+len(v) > 74 {
			// folding header
			v = strings.ReplaceAll(v, " ", "\r\n ")
		}
		err = writeString(w, v)
		err = writeString(w, "\r\n")
	}
	return err
}

func closeAttach(r io.Reader) {
	if c, ok := r.(io.Closer); ok {
		c.Close()
	}
}

func copyAttach(w io.Writer, r io.Reader) error {
	defer closeAttach(r)

	b := base64.NewEncoder(base64.StdEncoding, iox.NewMimeChunkWriter(w))
	_, err := io.Copy(b, r)
	if err != nil {
		return err
	}

	return b.Close()
}

func writeMsg(w io.Writer, encoding string, msg string) error {
	if encoding == "Base64" {
		b := base64.NewEncoder(base64.StdEncoding, iox.NewMimeChunkWriter(w))
		err := writeString(b, msg)
		if err != nil {
			return err
		}
		return b.Close()
	}
	return writeString(w, msg)
}

func writeString(w io.Writer, s string) error {
	_, err := w.Write([]byte(s))
	return err
}

func writeBody(w io.Writer, enc string, m *Email, boundary string) error {
	if boundary == "" {
		return writeMsg(w, enc, m.Message)
	}

	var err error

	header := header{}
	if m.HTML {
		header["Content-Type"] = "text/html; charset=UTF-8"
	} else {
		header["Content-Type"] = "text/plain; charset=UTF-8"
	}
	header["Content-Disposition"] = "inline"
	header["Content-Transfer-Encoding"] = enc

	// Write the message part
	err = writeString(w, "--"+boundary+"\n")
	err = writeHeader(w, header)
	err = writeString(w, "\r\n")
	if err != nil {
		return err
	}

	err = writeMsg(w, enc, m.Message)
	err = writeString(w, "\r\n")
	if err != nil {
		return err
	}

	// Append attachments
	err = writeAttachments(w, m.Attachments, boundary)
	if err != nil {
		return err
	}

	err = writeString(w, "--"+boundary+"--\r\n\r\n")
	return err
}

func writeAttachments(w io.Writer, as []*Attachment, boundary string) error {
	var err error
	for _, a := range as {
		mt := mime.TypeByExtension(filepath.Ext(a.Name))
		if mt == "" {
			mt = "application/octet-stream"
		}

		header := header{}
		header["Content-Type"] = mt + "; name=\"" + encodeWord(a.Name) + "\""
		if a.Cid != "" {
			header["Content-Disposition"] = "inline; filename=\"" + encodeWord(a.Name) + "\""
		} else {
			header["Content-Disposition"] = "attachment; filename=\"" + encodeWord(a.Name) + "\""
		}
		if a.Cid != "" {
			header["Content-ID"] = a.Cid
		}
		header["Content-Transfer-Encoding"] = "Base64"

		err = writeString(w, "--"+boundary+"\r\n")
		err = writeHeader(w, header)
		err = writeString(w, "\r\n")

		err = copyAttach(w, a.Data)
		if err != nil {
			return err
		}

		err = writeString(w, "\r\n")
		if err != nil {
			return err
		}
	}
	return nil
}
