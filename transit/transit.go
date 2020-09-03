// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package transit

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"regexp"
)

// Transit is a package for TCP forwarding
//
// It's different from general TCP proxy forwarding tools, such as
// Nginx reverse proxy. Transit's behaviors are more like a TCP listener.
// Unlike Nginx which can only perform one-direction TCP forwarding, Transit
// can perform two-directions TCP forwarding.
//
// Transit opens *a single* TCP port listens both for downstream and upstream.
//
// When a TCP connection comes from the downstream, transit will forward it to
// upstream. At the same time, it can also forward to a 3rd-party host.
//
// DownStream -> Transit -> UpStream
//					|-> 3rd-party host
//
// When a TCP connection comes from the upstream, transit will forward it to
// downstream. However, unlike above, Transit will not forward to 3rd-party host.
//
// DownStream <- Transit <- Upstream

// Transit main struct
type Transit struct {
	// IPArray
	// 0: Downstream, 1: Upstream
	IPArray [2]net.IP

	// ThirdPartyAddr is addr of 3rd-party
	ThirdPartyAddr string

	// IP is the interface IP
	IP net.IP

	// Port is the interface listen Port
	Port int

	// ln is the opened listener
	ln net.Listener
}

// Open opens the port for listening
func (t *Transit) Open() error {
	var err error
	// TODO: considering remove "tcp4"?
	addr := fmt.Sprintf("%s:%d", t.IP, t.Port)
	t.ln, err = net.Listen("tcp4", addr)
	if err != nil {
		return err
	}

	return nil
}

// Close closes the listening
func (t *Transit) Close() {
	t.ln.Close()
}

// Transit do the transiting
func (t *Transit) Transit() error {
	for {
		conn, err := t.ln.Accept()
		if err != nil {
			break
		}

		fmt.Printf("accepted %s\n", conn.RemoteAddr())

		remoteAddr := conn.RemoteAddr().String()
		match := regexp.MustCompile(`:[0-9][0-9]*$`)

		remoteIP := net.ParseIP(match.ReplaceAllString(remoteAddr, ""))

		srcID := -1
		for i, ip := range t.IPArray {
			if remoteIP.Equal(ip) {
				srcID = i
			}
		}

		if srcID == -1 {
			fmt.Printf("src IP %s is not recognized\n", conn.RemoteAddr())

			conn.Close()
			continue
		}

		var dstConn0, dstConn1 net.Conn

		addr := fmt.Sprintf("%s:%d", t.IPArray[(srcID+1)%2].String(), t.Port)
		dstConn0, err = net.Dial("tcp", addr)
		if err != nil {
			fmt.Printf("%v", err)
			dstConn0 = nil
		}

		if srcID == 0 {
			dstConn1, err = net.Dial("tcp", t.ThirdPartyAddr)
			if err != nil {
				fmt.Printf("%v", err)
				dstConn1 = nil
			}
		}

		if dstConn0 == nil && dstConn1 == nil {
			fmt.Printf("none of ports is ready for forwarding\n")

			conn.Close()
			continue
		}

		go func() {
			var writers []io.Writer
			if dstConn0 != nil {
				writers = append(writers, dstConn0)
			}

			if dstConn1 != nil {
				writers = append(writers, dstConn1)
			}

			pattern := `(serverip=)'\d+\.\d+\.\d+\.\d+'`
			replace := []byte(fmt.Sprintf("$1'%s'", t.IP))

			if _, err := copySed(io.MultiWriter(writers...), conn,
				0x0A, pattern, replace); err != nil {
				// panic(err)
			}

			if dstConn0 != nil {
				dstConn0.Close()
			}

			if dstConn1 != nil {
				dstConn1.Close()
			}

			conn.Close()
		}()

		go func() {
			if dstConn0 == nil {
				return
			}

			if _, err := io.Copy(conn, dstConn0); err != nil {
				// panic(err)
			}

			dstConn0.Close()
			conn.Close()
		}()
	}

	return nil
}

// copySed forked from io.Copy(), but will do sed-like substitution
func copySed(w io.Writer, r io.Reader, delim byte, pattern string, replace []byte) (written int64, err error) {
	src := bufio.NewReader(r)
	dst := w

	match := regexp.MustCompile(pattern)

	for {
		buf, er := src.ReadBytes(delim)
		toWrite := match.ReplaceAll(buf, replace)

		nr := len(toWrite)

		if nr > 0 {
			nw, ew := dst.Write(toWrite)

			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}

	return written, err
}
