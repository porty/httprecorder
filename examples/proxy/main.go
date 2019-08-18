package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/porty/httprecorder"
)

type Config struct {
	ProxyPort    int    `envconfig:"PROXY_PORT" required:"true"`
	RecorderPort int    `envconfig:"RECORDER_PORT" required:"true"`
	Upstream     string `envconfig:"UPSTREAM" required:"true"`
	MaxRequests  int    `envconfig:"MAX_REQUESTS" default:"100"`
}

func main() {
	_ = godotenv.Load(".env")
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		log.Print("Failed to process config: " + err.Error())
		os.Exit(1)
	}

	upstream, err := url.Parse(config.Upstream)
	if err != nil {
		panic(err)
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = upstream.Scheme
			req.URL.Host = upstream.Host
			req.Host = upstream.Host
		},
	}

	recorder := httprecorder.NewMemoryRecorder(config.MaxRequests)
	proxyAddr := fmt.Sprintf(":%d", config.ProxyPort)
	recorderAddr := fmt.Sprintf(":%d", config.RecorderPort)
	log.Printf("Listening on %s for proxy requests to %s", proxyAddr, config.Upstream)
	log.Printf("Listening on %s for the HTTP Recorder interface", recorderAddr)

	errs := make(chan error)

	go func() {
		errs <- http.ListenAndServe(proxyAddr, httprecorder.Middleware(recorder)(proxy))
	}()
	go func() {
		errs <- http.ListenAndServe(recorderAddr, httprecorder.UIHandler(recorder))
	}()

	err = <-errs
	panic(err)
}
