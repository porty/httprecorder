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

func readAndReplace(body *io.ReadCloser) ([]byte, error) {
	b, err := ioutil.ReadAll(*body)
	if err != nil {
		return nil, err
	}
	(*body).Close()
	*body = ioutil.NopCloser(bytes.NewReader(b))

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

			requestBody, err := readAndReplace(&r.Body)
			if err != nil {
				// TODO what to do here?
				log.Print("Error in recorder middleware reading request: " + err.Error())
			}

			// hack around lack of gzip support
			r.Header.Del("Accept-Encoding")

			next.ServeHTTP(httprec, r)
			responseReceived := time.Now()

			resp := httprec.Result()
			err = recorder.Record(r, requestBody, resp, requestReceived, responseReceived)
			if err != nil {
				// TODO what to do here?
				log.Print("Error in recorder middleware recording request: " + err.Error())
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
