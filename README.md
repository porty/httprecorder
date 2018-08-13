HTTP Recorder
===

Records HTTP sessions via HTTP middleware, can view via provided handler.

![image](https://user-images.githubusercontent.com/1373315/43304186-f58d820c-9127-11e8-93a8-99ed6d30d514.png)

## Usage

Have a look at [the simple-recorder example](https://github.com/porty/httprecorder/tree/master/examples/simple-recorder) or do something like:

```go
// Create an in memory recorder of max 100 requests
recorder := httprecorder.NewMemoryRecorder(100)

// host the UI outside of the middleware
go http.ListenAndServe("localhost:9001", httprecorder.UIHandler(recorder))

// Wrap some handlers with the provided middleware
mux := http.NewServeMux()
mux.HandleFunc("/", ...)
handler := httprecorder.Middleware(recorder)(mux)
http.ListenAndServe(addr, handler)
```

## Development

The HTML template is embedded in to the source using [esc](https://github.com/mjibson/esc)

```shell
# install esc if you haven't already got it
go get -u github.com/mjibson/esc

# create static_generated.go
esc -o embedded/static_generated.go -pkg embedded assets
```
