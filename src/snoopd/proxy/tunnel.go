package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"snoopd/cfg"
	"snoopd/log"
	"sync"
)

var cert tls.Certificate
func init()  {
	certFile := cfg.GetString("snoopd.tls_cert")
	keyFile := cfg.GetString("snoopd.tls_key")
	var err error
	cert, err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatal("Unable to load certificate key pair, err:", err)
	}
}

var roots *x509.CertPool
func init() {
	var err error
	roots, err = x509.SystemCertPool()
	if err != nil {
		log.Fatal("Unable to load system certificate pool for tls client, err:", err)
	}
}

var tlsConfig = &tls.Config{
	Certificates:       []tls.Certificate{cert},
	RootCAs:            roots,
	InsecureSkipVerify: true,
    CipherSuites:       []uint16{
    		tls.TLS_RSA_WITH_RC4_128_SHA,
            tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
            tls.TLS_RSA_WITH_AES_128_CBC_SHA,
            tls.TLS_RSA_WITH_AES_256_CBC_SHA,
            tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
            tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
            tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
            tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
            tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
            tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
            tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
            tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
            tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
            tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
            tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
            tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
            tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
            // TLS 1.3 cipher suites.
            tls.TLS_AES_128_GCM_SHA256,
            tls.TLS_AES_256_GCM_SHA384,
            tls.TLS_CHACHA20_POLY1305_SHA256,
        },
}

type tunnel struct {
	clientConn    net.Conn
	dstConn       net.Conn
	clientTlsConn *tls.Conn
	dstTlsConn    *tls.Conn
}

func newTunnel(clientConn, dstConn net.Conn) tunnel {
	t := tunnel{
		clientConn: clientConn,
		dstConn: 	dstConn,
	}
	t.clientTlsConn = tls.Server(t.clientConn, tlsConfig.Clone())
	t.dstTlsConn = tls.Client(t.dstConn, tlsConfig.Clone())
	err := t.clientTlsConn.Handshake()
	if err != nil {
		log.Error("Client side handshake finished with error:", err)
	}
	return t
}

func (t *tunnel) HandleAndClose() {
	var wg sync.WaitGroup
	wg.Add(2)
	go transfer(t.dstTlsConn, t.clientTlsConn, &wg, ">")
	go transfer(t.clientTlsConn, t.dstTlsConn, &wg, "<")
	wg.Wait()
	err := t.Close()
	if err != nil {
		log.Error("Tunnel handling exited with error:", err)
	}
}

func transfer(destination io.Writer, source io.Reader, wg *sync.WaitGroup, prefix string) {
	_, err := io.Copy(destination, source)
	if err != nil {
		log.Error("Transfer", prefix, "finished with error:", err)
	}
	wg.Done()
}

func (t *tunnel)Close()(err error) {
	defer log.Debug("Tunnel closed")
	err = t.dstTlsConn.Close()
	if err != nil {
		log.Error("Unable to close dstTlsConn, err:", err)
		return
	}

	err = t.clientTlsConn.Close()
	if err != nil {
		log.Error("Unable to close clientTlsListener, err:", err)
		return
	}

	return nil
}

