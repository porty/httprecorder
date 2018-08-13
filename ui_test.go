package httprecorder

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/require"
)

func TestUIIndex(t *testing.T) {
	m := getTestRecorder()
	doc := getTestUIRequest(t, "/", m)
	rows := []string{}

	doc.Find("[data-test-name=interactions-table] tbody tr").Each(func(_ int, s *goquery.Selection) {
		row := ""
		s.Find("th").Each(func(i int, s *goquery.Selection) {
			row = row + strings.TrimSpace(s.Text()) + " "
			// log.Printf("index = %d, selector = %#v", i, s.Text())
		})
		s.Find("td").Each(func(i int, s *goquery.Selection) {
			row = row + strings.TrimSpace(s.Text()) + " "
			// log.Printf("index = %d, selector = %#v", i, s.Text())
		})
		rows = append(rows, row)
	})

	require.Equal(t, []string{
		"0 GET / 200 OK - text/html ",
		"1 GET / 302 Found - text/plain ",
		"2 GET / 404 Not Found - - ",
		"3 POST /path/to/url?cat=dog 500 Internal Server Error text/plain application/json ",
	}, rows)
}

type navButtons struct {
	t       *testing.T
	buttons *goquery.Selection
}

func newNavButtons(t *testing.T, doc *goquery.Document) navButtons {
	buttons := doc.Find("nav ul li")
	require.True(t, buttons.Length() >= 3)
	numbered := buttons.Slice(1, buttons.Length()-1)
	numbered.Each(func(i int, s *goquery.Selection) {
		require.Equal(t, strconv.Itoa(i+1), strings.TrimSpace(s.Text()))
		href, _ := s.Find("a").Attr("href")
		require.Equal(t, "?page="+strconv.Itoa(i+1), href)
	})
	return navButtons{
		t:       t,
		buttons: buttons,
	}
}

func TestUIPagination(t *testing.T) {
	interaction := Interaction{
		RequestReceived:  time.Date(2018, time.December, 25, 6, 0, 0, 0, time.UTC),
		ResponseReceived: time.Date(2018, time.December, 25, 6, 0, 1, 1000000, time.UTC),
		Request: Request{
			Method: "GET",
			URL:    parseURL("/"),
		},
		Response: Response{
			StatusCode: 200,
		},
	}
	m := NewMemoryRecorder(100)
	for i := 0; i < 3; i++ {
		m.interactions = append(m.interactions, interaction)
	}
	doc := getTestUIRequest(t, "/", m)

	nav := newNavButtons(t, doc)
	require.True(t, nav.buttons.First().HasClass("disabled"))
	require.True(t, nav.buttons.Last().HasClass("disabled"))
	require.Equal(t, 3, nav.buttons.Length())

	m.interactions = nil
	for i := 0; i < 100; i++ {
		m.interactions = append(m.interactions, interaction)
	}

	doc = getTestUIRequest(t, "/", m)
	nav = newNavButtons(t, doc)
	require.True(t, nav.buttons.First().HasClass("disabled"))
	require.False(t, nav.buttons.Last().HasClass("disabled"))
	require.True(t, nav.buttons.Eq(1).HasClass("active"))
	require.Equal(t, 6, nav.buttons.Length())

	doc = getTestUIRequest(t, "/?page=4", m)
	nav = newNavButtons(t, doc)
	require.True(t, nav.buttons.Last().HasClass("disabled"))
	require.True(t, nav.buttons.Eq(4).HasClass("active"))
}

func TestUISelected(t *testing.T) {
	doc := getTestUIRequest(t, "/?index=0", getTestRecorder())

	timings := doc.Find("[data-test-name=timings]")
	codeBlocks := timings.Find("code")
	require.Equal(t, 3, codeBlocks.Length())

	require.Equal(t, "2018-12-25 06:00:00.000", strings.TrimSpace(codeBlocks.Eq(0).Text()))
	require.Equal(t, "2018-12-25 06:00:01.001", strings.TrimSpace(codeBlocks.Eq(1).Text()))
	require.Equal(t, "1.001", strings.TrimSpace(codeBlocks.Eq(2).Text()))

	request := doc.Find("[data-test-name=request]")
	expected := []string{"Method=GET", "URL=/", "User-Agent=Chrome"}
	var actual []string
	request.Find("tr").Each(func(i int, s *goquery.Selection) {
		key := strings.TrimSpace(s.Find("th").Text())
		value := strings.TrimSpace(s.Find("td").Text())
		actual = append(actual, key+"="+value)
	})
	require.Equal(t, expected, actual)

	response := doc.Find("[data-test-name=response]")
	expected = []string{"Status Code=200", "Content-Type=text/html", "Server=Not chrome"}
	actual = nil
	response.Find("tr").Each(func(i int, s *goquery.Selection) {
		key := strings.TrimSpace(s.Find("th").Text())
		value := strings.TrimSpace(s.Find("td").Text())
		actual = append(actual, key+"="+value)
	})
	// this might break if headers don't come out in the same order
	require.Equal(t, expected, actual)
}

func getTestRecorder() *MemoryRecorder {
	m := NewMemoryRecorder(10)
	m.interactions = []Interaction{
		{
			RequestReceived:  time.Date(2018, time.December, 25, 6, 0, 0, 0, time.UTC),
			ResponseReceived: time.Date(2018, time.December, 25, 6, 0, 1, 1000000, time.UTC),
			Request: Request{
				Body: nil,
				Headers: http.Header{
					"User-Agent": []string{"Chrome"},
				},
				Method: "GET",
				URL:    parseURL("/"),
			},
			Response: Response{
				Body: []byte("<marquee>hello</marquee>"),
				Headers: http.Header{
					"Content-Type": []string{"text/html"},
					"Server":       []string{"Not chrome"},
				},
				StatusCode: 200,
			},
		},
		{
			RequestReceived:  time.Date(2018, time.December, 25, 6, 0, 0, 0, time.UTC),
			ResponseReceived: time.Date(2018, time.December, 25, 6, 0, 1, 1000000, time.UTC),
			Request: Request{
				Body: nil,
				Headers: http.Header{
					"User-Agent": []string{"Chrome"},
				},
				Method: "GET",
				URL:    parseURL("/"),
			},
			Response: Response{
				Body: []byte("redirecting to /other"),
				Headers: http.Header{
					"Location":     []string{"/other"},
					"Content-Type": []string{"text/plain"},
					"Server":       []string{"Not chrome"},
				},
				StatusCode: http.StatusFound,
			},
		},
		{
			RequestReceived:  time.Date(2018, time.December, 25, 6, 0, 0, 0, time.UTC),
			ResponseReceived: time.Date(2018, time.December, 25, 6, 0, 1, 1000000, time.UTC),
			Request: Request{
				Body: nil,
				Headers: http.Header{
					"User-Agent": []string{"Chrome"},
				},
				Method: "GET",
				URL:    parseURL("/"),
			},
			Response: Response{
				Body: []byte("not found"),
				Headers: http.Header{
					"Server": []string{"Not chrome"},
				},
				StatusCode: 404,
			},
		},
		{
			RequestReceived:  time.Date(2018, time.December, 25, 6, 0, 0, 0, time.UTC),
			ResponseReceived: time.Date(2018, time.December, 25, 6, 0, 1, 1000000, time.UTC),
			Request: Request{
				Body: []byte("hello"),
				Headers: http.Header{
					"Content-Type": []string{"text/plain"},
					"User-Agent":   []string{"Chrome"},
				},
				Method: "POST",
				URL:    parseURL("/path/to/url?cat=dog"),
			},
			Response: Response{
				Body: []byte(`{ "response": "sup" }`),
				Headers: http.Header{
					"Content-Type": []string{"application/json"},
					"Server":       []string{"Not chrome"},
				},
				StatusCode: 500,
			},
		},
	}
	return m
}

func getTestUIRequest(t *testing.T, path string, m *MemoryRecorder) *goquery.Document {
	h := UIHandler(m)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", path, nil)

	h.ServeHTTP(rec, req)
	require.Equal(t, 200, rec.Code)

	doc, err := goquery.NewDocumentFromReader(rec.Body)
	require.NoError(t, err)

	return doc
}

func parseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
