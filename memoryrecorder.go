package httprecorder

import (
	"errors"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Request struct {
	Method  string
	URL     *url.URL
	Headers http.Header
	Body    []byte
}

type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

type Interaction struct {
	RequestReceived  time.Time
	ResponseReceived time.Time
	Request          Request
	Response         Response
}

type MemoryRecorder struct {
	limit int

	m            sync.Mutex
	interactions []Interaction
}

func NewMemoryRecorder(limit int) *MemoryRecorder {
	if limit <= 10 {
		limit = 10
	}
	return &MemoryRecorder{
		limit: limit,
	}
}

func makeRequest(in *http.Request, body []byte) (Request, error) {
	// TODO gzip encoding??
	return Request{
		Method:  in.Method,
		URL:     in.URL,
		Headers: in.Header,
		Body:    body,
	}, nil
}

func makeResponse(in *http.Response) (Response, error) {
	b, err := readAndReplace(&in.Body)
	if err != nil {
		return Response{}, errors.New("failed to read response body: " + err.Error())
	}
	if len(b) == 0 {
		b = nil
	}

	return Response{
		StatusCode: in.StatusCode,
		Headers:    in.Header,
		Body:       b,
	}, nil
}

func (m *MemoryRecorder) Record(req *http.Request, requestBody []byte, resp *http.Response, requestReceived time.Time, responseReceived time.Time) error {
	// TODO this doesn't make a copy of any of the parameters, pointers (and maps) could change via other middleware
	recReq, err := makeRequest(req, requestBody)
	if err != nil {
		return err
	}
	recResp, err := makeResponse(resp)
	if err != nil {
		return err
	}
	interaction := Interaction{
		RequestReceived:  requestReceived,
		ResponseReceived: responseReceived,
		Request:          recReq,
		Response:         recResp,
	}

	m.m.Lock()
	if len(m.interactions) >= m.limit {
		half := m.interactions[len(m.interactions)/2:]
		m.interactions = make([]Interaction, len(half), len(half)+1)
		copy(m.interactions, half)
	}
	m.interactions = append(m.interactions, interaction)
	m.m.Unlock()

	return nil
}

func (m *MemoryRecorder) Length() int {
	m.m.Lock()
	length := len(m.interactions)
	m.m.Unlock()
	return length
}

func (m *MemoryRecorder) GetInteractions(start int, length int) []Interaction {
	m.m.Lock()
	defer m.m.Unlock()

	if start+length > len(m.interactions) {
		length = len(m.interactions) - start
	}

	return m.interactions[start : start+length]
}

func (m *MemoryRecorder) GetInteraction(index int) Interaction {
	m.m.Lock()
	defer m.m.Unlock()

	return m.interactions[index]
}

func (m *MemoryRecorder) Clear() {
	m.m.Lock()
	defer m.m.Unlock()

	m.interactions = nil
}
