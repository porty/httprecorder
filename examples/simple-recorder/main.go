package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/porty/httprecorder"
)

func main() {
	recorder := httprecorder.NewMemoryRecorder(100)

	mux := http.NewServeMux()
	mux.HandleFunc("/generateResponse", generateResponse)
	mux.HandleFunc("/", index)

	http.Handle("/recorder/", http.StripPrefix("/recorder", httprecorder.UIHandler(recorder)))
	http.Handle("/", httprecorder.Middleware(recorder)(mux))

	log.Print("Connect to http://localhost:9000/")
	if err := http.ListenAndServe("localhost:9000", nil); err != nil {
		panic(err)
	}
}

func generateResponse(w http.ResponseWriter, r *http.Request) {
	statusCode, _ := strconv.Atoi(r.URL.Query().Get("statusCode"))
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	responseType := r.URL.Query().Get("responseType")
	switch responseType {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		fmt.Fprint(w, `{"users":[{"id":123,"name":"Bazza","groups":["jenkins","linux"]},{"id":321,"name":"Shazza","groups":["linux","root"]}]}`)
	case "html":
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(statusCode)
		fmt.Fprint(w, `<html><head><title>This is HTML</title></head><body><p>This is <b>HTML!</b></p><img src="https://i.imgur.com/WI1qXi7.gif"></body></html>`)
	case "gif":
		resp, err := http.Get("https://i.imgur.com/WI1qXi7.gif")
		if err != nil {
			http.Error(w, "failed to download gif: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		if resp.Header.Get("Content-Type") != "image/gif" {
			http.Error(w, "the llama gif I thought was available isn't available any more, time to update the example", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/gif")
		w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
		w.WriteHeader(statusCode)
		io.Copy(w, resp.Body)
	default:
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(statusCode)
		fmt.Fprint(w, "this is some text")
	}
}

const indexText = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>HTTP Recorder</title>
</head>
<body>
	<form action="generateResponse">
		<h2>Status Code</h2>
		<input name="statusCode" placeholder="status code" value="200">

		<h2>Response Type</h2>
		<label><input type="radio" name="responseType" value="json" checked> JSON</label><br>
		<label><input type="radio" name="responseType" value="html"> HTML</label><br>
		<label><input type="radio" name="responseType" value="gif"> GIF</label><br>
		<label><input type="radio" name="responseType" value=""> plain text</label><br>
		<input type="submit" value="OK">
	</form>
	<hr>
	<a href="/recorder">The recorder UI</a>
</body>
</html>
`

func index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, indexText)
}
