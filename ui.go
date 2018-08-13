package httprecorder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"

	"github.com/porty/httprecorder/embedded"
)

func UIHandler(recorder Recorder) http.Handler {
	mux := chi.NewMux()
	mux.HandleFunc("/", index(recorder))
	mux.Get("/data", data(recorder))
	return mux
}

type indexData struct {
	Index             int
	Page              int
	NumPages          int
	Interaction       *Interaction
	Interactions      []Interaction
	ShowFormattedJSON bool
	AllPages          []int
	PrevPage          int
	NextPage          int
}

func statusBadgeClass(statusCode int) string {
	switch {
	case statusCode < 300:
		return "badge-success"
	case statusCode < 400:
		return "badge-secondary"
	case statusCode < 500:
		return "badge-warning"
	default:
		return "badge-danger"
	}
}

func requestContentType(interaction Interaction) string {
	return interaction.Request.Headers.Get("Content-Type")
}

func responseContentType(interaction Interaction) string {
	return interaction.Response.Headers.Get("Content-Type")
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.000")
}

func duration(t0, t1 time.Time) string {
	return fmt.Sprintf("%.3f", t1.Sub(t0).Seconds())
}

const itemsPerPage = 30

func index(recorder Recorder) func(http.ResponseWriter, *http.Request) {
	t := template.Must(template.New("index.html").Funcs(template.FuncMap{
		"statusText":          http.StatusText,
		"statusBadgeClass":    statusBadgeClass,
		"requestContentType":  requestContentType,
		"responseContentType": responseContentType,
		"formatTime":          formatTime,
		"duration":            duration,
		"offsetIndex":         offsetIndex,
	}).Parse(embedded.FSMustString(false, "/assets/index.html")))

	return func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("action") == "clear" {
			recorder.Clear()
		}

		length := recorder.Length()

		index, err := strconv.Atoi(r.URL.Query().Get("index"))
		if err != nil || index >= recorder.Length() {
			index = -1
		}
		lastPage := ((length - 1) / itemsPerPage) + 1
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		} else if page > lastPage {
			page = lastPage
		}

		interactions := recorder.GetInteractions((page-1)*itemsPerPage, itemsPerPage)
		var interaction *Interaction
		if index >= 0 {
			i := recorder.GetInteraction(index)
			interaction = &i
		}
		var numPages int
		if length == 0 {
			numPages = 1
		} else {
			numPages = lastPage
		}
		prevPage := -1
		if page > 1 {
			prevPage = page - 1
		}
		nextPage := -1
		if page < numPages {
			nextPage = page + 1
		}
		if err := t.ExecuteTemplate(w, "index.html", indexData{
			Index:        index,
			Page:         page,
			NumPages:     numPages,
			Interaction:  interaction,
			Interactions: interactions,
			AllPages:     createPages(numPages),
			PrevPage:     prevPage,
			NextPage:     nextPage,
		}); err != nil {
			log.Print("Failed to render index template: " + err.Error())
		}
	}
}

func data(recorder Recorder) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		index, err := strconv.Atoi(r.URL.Query().Get("index"))
		if err != nil {
			http.Error(w, "invalid index", http.StatusBadRequest)
			return
		}

		if index >= recorder.Length() {
			http.Error(w, "invalid index", http.StatusBadRequest)
			return
		}

		i := recorder.GetInteraction(index)
		var body []byte
		if _, ok := r.URL.Query()["request"]; ok {
			body = i.Request.Body
		} else {
			body = i.Response.Body
		}

		switch r.URL.Query().Get("format") {
		case "json":
			buf := bytes.Buffer{}
			if err := json.Indent(&buf, body, "", "  "); err != nil {
				http.Error(w, "failed to format JSON: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			io.Copy(w, &buf)
			return
		}

		contentType := r.URL.Query().Get("content")
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		io.Copy(w, bytes.NewReader(body))
	}
}

func createPages(numPages int) []int {
	retval := make([]int, numPages)
	for i := 0; i < numPages; i++ {
		retval[i] = i + 1
	}
	return retval
}

func offsetIndex(index, page int) int {
	return (page-1)*itemsPerPage + index
}
