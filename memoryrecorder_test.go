package httprecorder

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMemoryRecorderRecord(t *testing.T) {
	m := NewMemoryRecorder(10)

	req := httptest.NewRequest(http.MethodPost, "http://some.server/rofl/copter?cat=dog", ioutil.NopCloser(strings.NewReader("one=1&two=2")))
	req.Header.Add("User-Agent", "this test")
	resp := http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Server": []string{"also this test"},
		},
		Body: ioutil.NopCloser(strings.NewReader("hello, this is a response!")),
	}
	start := time.Date(2017, time.December, 25, 5, 0, 0, 0, time.UTC)
	end := start.Add(1 * time.Second)

	// TODO does resp need to be a pointer?
	err := m.Record(req, []byte("one=1&two=2"), &resp, start, end)

	require.NoError(t, err)

	// check that we can still read the request body
	require.Equal(t, "one=1&two=2", readString(req.Body))

	// check we can still read the response body
	require.Equal(t, "hello, this is a response!", readString(resp.Body))

	require.Equal(t, 1, m.Length())
	interactions := m.GetInteractions(0, 99)
	require.Equal(t, 1, len(interactions))

	i := interactions[0]
	require.Equal(t, start, i.RequestReceived)
	require.Equal(t, end, i.ResponseReceived)

	require.Equal(t, http.MethodPost, i.Request.Method)
	require.Equal(t, "http://some.server/rofl/copter?cat=dog", i.Request.URL.String())
	require.Equal(t, req.Header, i.Request.Headers)
	require.Equal(t, "one=1&two=2", string(i.Request.Body))

	require.Equal(t, http.StatusOK, i.Response.StatusCode)
	require.Equal(t, resp.Header, i.Response.Headers)
	require.Equal(t, "hello, this is a response!", string(i.Response.Body))

	m.Clear()
	require.Equal(t, 0, m.Length())
}

func TestMemoryRecorderLimit(t *testing.T) {
	m := NewMemoryRecorder(20)

	time := time.Now()
	u, _ := url.Parse("/url")
	req := http.Request{
		Method: "GET",
		URL:    u,
		Body:   ioutil.NopCloser(bytes.NewReader(nil)),
	}
	for i := 0; i < 20; i++ {
		resp := http.Response{
			StatusCode: 200 + i,
			Body:       ioutil.NopCloser(bytes.NewReader(nil)),
		}
		require.NoError(t, m.Record(&req, nil, &resp, time, time))
	}

	require.Equal(t, 20, m.Length())

	resp := http.Response{
		StatusCode: 220,
		Body:       ioutil.NopCloser(bytes.NewReader(nil)),
	}
	require.NoError(t, m.Record(&req, nil, &resp, time, time))

	require.Equal(t, 11, m.Length())
	require.Equal(t, 210, m.GetInteraction(0).Response.StatusCode)
	require.Equal(t, 220, m.GetInteraction(10).Response.StatusCode)
}

func readString(r io.Reader) string {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return string(b)
}
