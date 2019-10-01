package proxy

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"snoopd/cfg"
	"snoopd/log"
	"snoopd/storadge"
	"strconv"
	"time"
)

func handleTunneling(w http.ResponseWriter, r *http.Request) {
	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	t := newTunnel(clientConn, destConn)
	t.HandleAndClose()
}

func handleHTTP(w http.ResponseWriter, req *http.Request) {
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	err = storadge.Store(req, resp)
	if err != nil {
		log.Error("Unable to store round trip result, err:", err)
	}

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

	//certFile := cfg.GetString("snoopd.tls_cert")
	//keyFile := cfg.GetString("snoopd.tls_key")

	log.Info("Listening for HTTPS on port", port)
	err := server.ListenAndServe()//TLS(certFile, keyFile)
	if err != nil {
		log.Fatal("Fatal error on HTTPS listener:", err)
	}
}
