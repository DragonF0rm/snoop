package proxy
import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"snoopd/cfg"
	"snoopd/log"
	"strconv"
	"time"
)

type MirrorConn struct {
	conn net.Conn
	closed bool
}

func NewMirrorConn(conn net.Conn) (*MirrorConn, error) {
	return &MirrorConn{
		conn:   conn,
		closed: false,
	}, nil
}

func (mc *MirrorConn)Read(p []byte) (n int, err error) {
	n, err = mc.conn.Read(p)
	fmt.Println(string(p[:n]))
	return n, err
}

func (mc *MirrorConn)Write(p []byte) (n int, err error) {
	n, err = mc.conn.Write(p)
	fmt.Println(string(p[:n]))
	return n, err
}

func (mc *MirrorConn)Close()(err error) {
	err = mc.conn.Close()
	if !mc.closed {
		//This method may being called multiple times for each MirrorConn
		//but connSlot must be released only once
		mc.closed = true
	}
	return
}

func handleTunneling(w http.ResponseWriter, r *http.Request) {
	dest_conn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	dest_mc, _ := NewMirrorConn(dest_conn)

	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	client_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	client_mc, _ := NewMirrorConn(client_conn)

	go transfer(dest_mc, client_mc)
	go transfer(client_conn, dest_conn)
}
func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}
func handleHTTP(w http.ResponseWriter, req *http.Request) {
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func ListenAndServe() {
	port := cfg.GetInt("snoopd.http_port")
	server := &http.Server{
		Addr: ":" + strconv.Itoa(port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debug("Got some HTTP request")
			handleHTTP(w, r)
		}),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	log.Info("Listening for HTTP on port", port)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("Fatal error on HTTP listener:", err)
	}
}

func ListenAndServeTLS() {
	port := cfg.GetInt("snoopd.https_port")
	server := &http.Server{
		Addr: ":" + strconv.Itoa(port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debug("Got some HTTPS request")
			if r.Method == http.MethodConnect {
				handleTunneling(w, r)
			} else {
				handleHTTP(w, r)
			}
		}),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	certFile := cfg.GetString("snoopd.tls_cert")
	keyFile := cfg.GetString("snoopd.tls_key")

	log.Info("Listening for HTTPS on port", port)
	err := server.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		log.Fatal("Fatal error on HTTPS listener:", err)
	}
}
