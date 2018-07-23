package httprecorder

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"time"
)

type Recorder interface {
	Record(req *http.Request, resp *http.Response, requestReceived time.Time, responseReceived time.Time) error
	Length() int
	GetInteractions(start int, length int) []Interaction
	GetInteraction(index int) Interaction
	Clear()
}

func Middleware(recorder Recorder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestReceived := time.Now()
			httprec := httptest.NewRecorder()

			next.ServeHTTP(httprec, r)
			responseReceived := time.Now()

			resp := httprec.Result()
			err := recorder.Record(r, resp, requestReceived, responseReceived)
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
