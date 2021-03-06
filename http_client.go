package httpretry

import (
	"net"
	"net/http"
	"time"
)

// ClientWithTimeout is an http client optimized for high throughput.  It times
// out more agressively than the default http client in net/http as well as
// setting deadlines on the TCP connection.
//
// Taken from s3gof3r:
// https://github.com/rlmcpherson/s3gof3r/blob/1e759738ff170bd0381a848337db677dbdd6aa62/http_client.go
//
func ClientWithTimeout(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial:  DialWithTimeout(timeout),
		ResponseHeaderTimeout: timeout,
		MaxIdleConnsPerHost:   10,
	}
	return &http.Client{Transport: transport}
}

// DialWithTimeout creates a Dial function that returns a connection with an
// inactivity timeout for the given duration.  This is designed for long running
// HTTP requests.
func DialWithTimeout(timeout time.Duration) func(netw, addr string) (net.Conn, error) {
	return func(netw, addr string) (net.Conn, error) {
		c, err := net.DialTimeout(netw, addr, timeout)
		if err != nil {
			return nil, err
		}
		if tc, ok := c.(*net.TCPConn); ok {
			tc.SetKeepAlive(true)
			tc.SetKeepAlivePeriod(timeout)
		}
		return &deadlineConn{timeout, c}, nil
	}
}

type deadlineConn struct {
	Timeout time.Duration
	net.Conn
}

func (c *deadlineConn) Read(b []byte) (int, error) {
	if err := c.Conn.SetDeadline(time.Now().Add(c.Timeout)); err != nil {
		return 0, err
	}
	return c.Conn.Read(b)
}

func (c *deadlineConn) Write(b []byte) (int, error) {
	if err := c.Conn.SetDeadline(time.Now().Add(c.Timeout)); err != nil {
		return 0, err
	}

	return c.Conn.Write(b)
}
