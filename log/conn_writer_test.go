package log

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func testConnTCPServer(sigChan chan string, finChan chan string, revChan chan string) {
	wg := &sync.WaitGroup{}

	ln, err := net.Listen("tcp", ":9999")
	if err != nil {
		fmt.Println("Error listening: ", err.Error())
		os.Exit(1)
	}

	// Listen and accept incoming connections
	tl, _ := ln.(*net.TCPListener)
	for {
		tl.SetDeadline(time.Now().Add(time.Second * 1))
		conn, err := ln.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				select {
				case <-sigChan:
					wg.Wait()
					ln.Close()
					fmt.Println("Listen Done.")
					finChan <- "done"
					return
				default:
				}
				continue
			}
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		wg.Add(1)
		go testConnEcho(conn, wg, revChan)
	}
}

func testConnEcho(conn net.Conn, wg *sync.WaitGroup, revChan chan string) {
	sc := bufio.NewScanner(conn)
	for sc.Scan() {
		s := sc.Text()
		revChan <- s
		os.Stdout.Write([]byte(s))
		os.Stdout.Write([]byte("\r\n"))
		if strings.Contains(s, "Close!") {
			fmt.Println("Server Close Connection!")
			conn.Close()
			wg.Done()
			return
		}
	}
}

func TestConnWriter(t *testing.T) {
	sigChan := make(chan string, 1)
	finChan := make(chan string, 1)
	revChan := make(chan string, 100)
	go testConnTCPServer(sigChan, finChan, revChan)

	time.Sleep(time.Second)
	log := NewLog()
	log.SetWriter(&ConnWriter{Addr: "localhost:9999"})
	log.SetFormatter(NewTextFormatter("%m%n"))

	ss := []string{
		"Hello Trace",
		"Hello Debug",
		"Hello Info - Close!",
		"Hello Warn",
		"Hello Error",
		"Hello Fatal - Close!",
	}

	i := 0
	log.Trace(ss[i])
	i++
	log.Debug(ss[i])
	i++
	log.Info(ss[i])
	i++
	time.Sleep(time.Millisecond * 500)
	log.Info(strings.Repeat("!missing! ", 100))
	log.Warn(ss[i])
	i++
	log.Error(ss[i])
	i++
	log.Fatal(ss[i])

	sigChan <- "done"
	<-finChan

	rs := []string{}
	for {
		if len(revChan) > 0 {
			s := <-revChan
			rs = append(rs, s)
			continue
		}
		break
	}

	if !reflect.DeepEqual(ss, rs) {
		t.Errorf("TestConnWriter() failure\nexcept: %q\nactual: %q", ss, rs)
	}
}
