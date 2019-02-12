package httprecorder

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Response-Header", "yo")
		w.WriteHeader(404)
		fmt.Fprint(w, "hello!")
	})
	recorder := NewMemoryRecorder(10)
	handler = Middleware(recorder)(handler)
	server := httptest.NewServer(handler)
	defer server.Close()
	req, err := http.NewRequest(http.MethodGet, server.URL+"/butts", nil)
	require.NoError(t, err)
	req.Header.Set("X-Request-Header", "sup")

	resp, err := http.DefaultClient.Do(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 1, recorder.Length())
	interactions := recorder.GetInteractions(0, 1)
	i := interactions[0]

	require.Equal(t, http.MethodGet, i.Request.Method)
	require.Equal(t, "/butts", i.Request.URL.String())
	require.Equal(t, "sup", i.Request.Headers.Get("X-Request-Header"))
	require.Nil(t, i.Request.Body)

	require.Equal(t, 404, i.Response.StatusCode)
	require.Equal(t, "yo", i.Response.Headers.Get("X-Response-Header"))
	require.Equal(t, "hello!", string(i.Response.Body))
}

func TestMiddlewarePost(t *testing.T) {
	var handler http.Handler
	var receivedBody string
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody = readString(r.Body)
		err := r.Body.Close()
		require.NoError(t, err)

		fmt.Fprint(w, "hello!")
	})
	recorder := NewMemoryRecorder(10)
	handler = Middleware(recorder)(handler)
	server := httptest.NewServer(handler)
	defer server.Close()
	req, err := http.NewRequest(http.MethodPost, server.URL+"/test", ioutil.NopCloser(strings.NewReader("this is request data")))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, "this is request data", receivedBody)

	require.Equal(t, 1, recorder.Length())
	interactions := recorder.GetInteractions(0, 1)
	i := interactions[0]

	require.Equal(t, http.MethodPost, i.Request.Method)
	require.Equal(t, "/test", i.Request.URL.String())
	require.Equal(t, "this is request data", string(i.Request.Body))

	require.Equal(t, 200, i.Response.StatusCode)
	require.Equal(t, "hello!", string(i.Response.Body))
}

func TestMiddlewareWithNilBody(t *testing.T) {
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprint(w, "hello!")
	})
	recorder := NewMemoryRecorder(10)
	handler = Middleware(recorder)(handler)

	req, err := http.NewRequest(http.MethodGet, "https://whatever/butts", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	require.Equal(t, "hello!", w.Body.String())
}
