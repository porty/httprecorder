package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/porty/httprecorder"
)

func makeProxy(s string) http.Handler {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}

	return httputil.NewSingleHostReverseProxy(u)
}

func startServer(message string, addr string, errs chan<- error) {
	log.Printf("Starting server at http://%s/", addr)
	go func() {
		errs <- http.ListenAndServe(addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, ".json") {
				m := map[string]string{
					"message": message,
					"url":     r.URL.String(),
				}
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(m); err != nil {
					log.Print("Error encoding JSON to client: " + err.Error())
				}
			} else if strings.HasSuffix(r.URL.Path, ".html") {
				w.Header().Set("Content-Type", "text/html")
				fmt.Fprint(w, "<html><head><title>The Title</title></head><body>"+message+" - URL = "+r.URL.String()+"</body></html>")
			} else {
				fmt.Fprint(w, message+" - URL = "+r.URL.String())
			}
		}))
	}()
}

func main() {
	proxyRouter := http.NewServeMux()

	proxyRouter.Handle("/one/", makeProxy("http://localhost:9001/"))
	proxyRouter.Handle("/two/", makeProxy("http://localhost:9002/"))
	proxyRouter.Handle("/three/", makeProxy("http://localhost:9003/"))

	errs := make(chan error)
	startServer("this is server one", "localhost:9001", errs)
	startServer("this is server two", "localhost:9002", errs)
	startServer("this is server three", "localhost:9003", errs)

	log.Print("Connect to http://localhost:9000/ for proxied connections")
	log.Print("Connect to http://localhost:9004/ for recorder UI")
	r := httprecorder.NewMemoryRecorder()
	recordingProxyRouter := httprecorder.Middleware(r)(proxyRouter)
	go func() {
		errs <- http.ListenAndServe("localhost:9004", httprecorder.UIHandler(r))
	}()
	go func() {
		errs <- http.ListenAndServe("localhost:9000", recordingProxyRouter)
	}()

	err := <-errs
	panic(err)
}
