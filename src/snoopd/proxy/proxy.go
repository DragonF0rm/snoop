package proxy

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"snoopd/cfg"
	"snoopd/log"
	"strconv"
	"sync"
	"time"
)


func handleTunneling(w http.ResponseWriter, r *http.Request) {
	dstConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second) //TODO put timeout to config
	if err != nil {
		//TODO log error
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		//TODO log error
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	srcConn, _, err := hijacker.Hijack()
	if err != nil {
		//TODO log error
		log.Debug("Error on Hijack:", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	wg := sync.WaitGroup{}
	go func(wg *sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()
		var buf []byte
		dstConn.Read(buf)
		log.Debug("<", string(buf))
		srcConn.Write(buf)
	}(&wg)
	go func(wg *sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()
		var buf []byte
		srcConn.Read(buf)
		log.Debug(">",string(buf))
		dstConn.Write(buf)
	}(&wg)
	wg.Wait()
	srcConn.Close()
	dstConn.Close()
}

func handleHTTP(w http.ResponseWriter, req *http.Request) {
	//TODO save request
	log.Request(req)
	//fmt.Println(req)
	resp, err := http.DefaultTransport.RoundTrip(req)
	//TODO log response
	if err != nil {
		//TODO log error
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			//TODO log error
		}
	}()

	for headerName, headerValues := range resp.Header {
		for _, val := range headerValues {
			w.Header().Add(headerName, val)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		//TODO log error
	}
}

func ListenAndServe() {
	port := cfg.GetInt("snoopd.http_port")
	server := &http.Server{
		Addr: ":" + strconv.Itoa(port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debug("Got some request by HTTP:", r.Method, r.URL.String(), r.Proto)
			handleHTTP(w, r)
		}),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)), // Disable HTTP/2.
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("Unable to start HTTP proxy server on port", err)
	}
}

func LisetnAndServeTLS() {
	port := cfg.GetInt("snoopd.https_port")
	server := &http.Server{
		Addr: ":" + strconv.Itoa(port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debug("Got some request by HTTPS:", r.Method, r.URL.String(), r.Proto)
			if r.Method == http.MethodConnect {
				handleTunneling(w, r)
			} else {
				handleHTTP(w, r)
			}
		}),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)), // Disable HTTP/2.
	}
	certFile := cfg.GetString("snoopd.tls_cert")
	keyFile := cfg.GetString("snoopd.tls_key")
	err := server.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		log.Fatal("Unable to start HTTP proxy server on port", err)
	}
}
