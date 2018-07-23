package httprecorder

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	var handler http.Handler
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Response-Header", "yo")
		w.WriteHeader(404)
		fmt.Fprint(w, "hello!")
	})
	recorder := NewMemoryRecorder()
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
