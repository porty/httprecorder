package httprecorder

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"time"
)

type Recorder interface {
	Record(req *http.Request, requestBody []byte, resp *http.Response, requestReceived time.Time, responseReceived time.Time) error
	Length() int
	GetInteractions(start int, length int) []Interaction
	GetInteraction(index int) Interaction
	Clear()
}

func replaceBody(r *http.Request) ([]byte, error) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body.Close()
	r.Body = ioutil.NopCloser(bytes.NewReader(b))

	if len(b) == 0 {
		b = nil
	}
	return b, nil
}

func Middleware(recorder Recorder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestReceived := time.Now()
			httprec := httptest.NewRecorder()

			requestBody, err := replaceBody(r)

			// hack around lack of gzip support
			r.Header.Del("Accept-Encoding")

			next.ServeHTTP(httprec, r)
			responseReceived := time.Now()

			resp := httprec.Result()
			err = recorder.Record(r, requestBody, resp, requestReceived, responseReceived)
			if err != nil {
				// TODO what to do here?
				log.Print("Error in recorder middleware: " + err.Error())
			}

			for k, v := range resp.Header {
				w.Header()[k] = v
			}
			w.WriteHeader(resp.StatusCode)
			// assumes the resp.Body has been replaced by something readable again
			io.Copy(w, resp.Body)
		})
	}
}
