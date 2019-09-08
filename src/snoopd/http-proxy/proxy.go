package http_proxy

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
		//log.Debug(wg)
		io.Copy(dstConn, srcConn)
	}(&wg)
	go func(wg *sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()
		io.Copy(srcConn, dstConn)
	}(&wg)
	wg.Wait()
	srcConn.Close()
	dstConn.Close()
}

func handleHTTP(w http.ResponseWriter, req *http.Request) {
	//TODO save request
	log.Request(req)
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
	port := cfg.GetInt("snoopd.http_proxy.port")
	server := &http.Server{
		Addr: ":" + strconv.Itoa(port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debug("Got some request:", r.Method, r.RequestURI, r.Proto)
			if r.Method == http.MethodConnect {
				handleTunneling(w, r)
			} else {
				handleHTTP(w, r)
			}
		}),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)), // Disable HTTP/2.
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("Unable to start HTTP proxy server on port")
	}
}
