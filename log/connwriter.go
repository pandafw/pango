package log

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

// ConnWriter implements Writer.
// it writes messages in keep-live tcp connection.
type ConnWriter struct {
	Net     string
	Addr    string
	Timeout time.Duration
	Logfmt  Formatter // log formatter
	Logfil  Filter    // log filter

	conn io.WriteCloser
	bb   bytes.Buffer
}

// SetFormat set a log formatter
func (cw *ConnWriter) SetFormat(format string) {
	cw.Logfmt = NewTextFormatter(format)
}

// SetTimeout set timeout
func (cw *ConnWriter) SetTimeout(timeout string) error {
	tmo, err := time.ParseDuration(timeout)
	if err != nil {
		return fmt.Errorf("ConnkWriter - Invalid timeout: %v", err)
	}
	cw.Timeout = tmo
	return nil
}

// Write write logger message to connection.
func (cw *ConnWriter) Write(le *Event) {
	if cw.Logfil != nil && cw.Logfil.Reject(le) {
		return
	}

	le.Logger.Lock()
	defer le.Logger.Unlock()

	if cw.Logfmt == nil {
		cw.Logfmt = le.Logger.GetFormatter()
	}

	cw.dial()
	if cw.conn == nil {
		return
	}

	// format msg
	cw.bb.Reset()
	cw.Logfmt.Write(&cw.bb, le)

	// write log
	_, err := cw.conn.Write(cw.bb.Bytes())
	if err != nil {
		// This is probably due to a timeout, so reconnect and try again.
		cw.Close()
		cw.dial()
		if cw.conn == nil {
			return
		}
		_, err := cw.conn.Write(cw.bb.Bytes())
		if err != nil {
			fmt.Fprintf(os.Stderr, "ConnWriter(%q) - Write(%s): %v\n", cw.Addr, cw.bb.Bytes(), err)
			cw.Close()
		}
	}
}

// Flush implementing method. empty.
func (cw *ConnWriter) Flush() {
}

// Close close the file description, close file writer.
func (cw *ConnWriter) Close() {
	if cw.conn != nil {
		err := cw.conn.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ConnWriter(%q) - Close(): %v\n", cw.Addr, err)
		}
		cw.conn = nil
	}
}

func (cw *ConnWriter) dial() {
	if cw.conn != nil {
		return
	}

	if cw.Net == "" {
		cw.Net = "tcp"
	}

	conn, err := net.DialTimeout(cw.Net, cw.Addr, cw.Timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ConnWriter(%q) - Dial(%q): %v\n", cw.Addr, cw.Net, err)
		return
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
	}

	cw.conn = conn
}
