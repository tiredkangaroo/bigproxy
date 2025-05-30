package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	certificate "github.com/tiredkangaroo/bigproxy/proxy/certificates"
	"github.com/tiredkangaroo/bigproxy/proxy/config"
)

var (
	// ErrResponseWriterNoHijack is returned when the ResponseWriter does not implement the Hijacker interface.
	ErrResponseWriterNoHijack = errors.New("response writer cannot be hijacked: does not implement Hijacker interface")
)
var (
	// ResponseHijackingError is the response sent to the client when the proxy cannot hijack the connection.
	ResponseHijackingError []byte = []byte("hijacking error")

	// ResponseRawSuccess is the response sent to the client when the HTTPS proxy successfully hijacks the connection.
	ResponseRawSuccess []byte = []byte("HTTP/1.1 200 OK\r\n" +
		"Connection: close\r\n" +
		"Content-Length: 0\r\n" +
		"\r\n")
)

// look over HTTP/2 handling and streaming responses

// handleHTTP handles a HTTP request to the proxy.
//
// HTTP requests do not create secure tunnels, therefore the original request to the
// proxy is the one meant for the host. The only prep we need to do is strip proxy
// headers. The proxy will then perform the request to the host and send the response
// back to the client.
func (r *Request) handleHTTP(m *Manager) error {
	m.SendRequest(r)

	// perform the request
	resp, raw, err := r.Perform(m)
	if err != nil {
		return fmt.Errorf("perform: %w", err)
	}
	r.resp = resp

	m.SendResponse(r)

	if _, err = r.conn.Write(raw); err != nil {
		return fmt.Errorf("connection write: %w", err)
	}
	return nil
}

// handleHTTPS handles a HTTPS request. This proxy is a MITM proxy, so it will
// use a self-signed certificate for the requested host and send it to the client.
//
// Steps:
// (1) It writes a success response to the client. The client believes it is no longer speaking to the proxy now.
// (2) A TLS handshake (using the self-signed) occurs between the client and the proxy. The client believes it is speaking to the host.
// (3) The proxy reads the actual request from the client that it meant to send the host.
// (4) The proxy performs the request to the real host.
// (5) The proxy sends the response back to the client.
func (r *Request) handleHTTPS(m *Manager, c *certificate.Certificates) error {
	// write a success response to the client (this is meant to be the last thing before the secure tunnel is expected)
	_, err := r.conn.Write(ResponseRawSuccess)
	if err != nil {
		return fmt.Errorf("connection write: %w", err)
	}

	if !config.DefaultConfig.MITM {
		return r.handleNoMITM(m)
	}

	// after the success response, a handshake will occur and the user will
	// send the ACTUAL request

	// NOTE: consider IPV6 square bracket and how that affects the hostname
	tlsconn, err := c.TLSConn(r.conn, r.req.URL.Hostname())
	if err != nil {
		return fmt.Errorf("tls conn: %w", err)
	}

	buf := bufio.NewReader(tlsconn)
	finalReq, err := http.ReadRequest(buf)
	if err != nil {
		return fmt.Errorf("read mitm request: %w", err)
	}
	r.req = finalReq

	m.SendRequest(r)

	resp, raw, err := r.Perform(m)
	if err != nil {
		return fmt.Errorf("perform: %w", err)
	}
	r.resp = resp

	m.SendResponse(r)

	_, err = tlsconn.Write(raw)
	if err != nil {
		return fmt.Errorf("tls connection write: %w", err)
	}

	return nil
}

// handleNoMITM handles an HTTPS connection without man-in-the-middling it. It just establishes a secure
// tunnel.
func (r *Request) handleNoMITM(m *Manager) error {
	// this code will need to be combined because it's the same in Perform and here in tunneling
	if config.DefaultConfig.PerformDelay != 0 {
		time.Sleep(time.Duration(config.DefaultConfig.PerformDelay) * time.Millisecond)
	}
	if config.DefaultConfig.RequireApproval {
		if !m.RecieveApproval(r) {
			return ErrPerformStop
		}
	}

	hconn, err := net.Dial("tcp", r.req.Host)
	if err != nil {
		return fmt.Errorf("dial host: %w", err)
	}

	// me gusta context :)
	ctx, cancel := context.WithCancel(context.Background())

	m.SendTunnel(r)
	go func() {
		defer cancel()
		_, err := io.Copy(hconn, r.conn)
		if err != nil {
			slog.Warn("io.Copy error (conn -> hostconn)", "err", err)
		}
	}()

	go func() {
		defer cancel()
		_, err := io.Copy(r.conn, hconn)
		if err != nil {
			slog.Warn("io.Copy error (hconn -> conn)", "err", err)
		}
	}()

	<-ctx.Done()

	m.SendDone(r)

	return nil
}

// hijack hijacks the ResponseWriter connection to the client.
func hijack(w http.ResponseWriter) (net.Conn, error) {
	h, ok := w.(http.Hijacker)
	if !ok {
		return nil, ErrResponseWriterNoHijack
	}
	c, _, err := h.Hijack()
	return c, err
}
