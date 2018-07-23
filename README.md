HTTP Recorder
===

Records HTTP sessions via HTTP middleware, can view via provided handler.

## Development

The HTML template is embedded in to the source using [esc](https://github.com/mjibson/esc)

```shell
# install esc if you haven't already got it
go get -u github.com/mjibson/esc

# create static_generated.go
esc -o static_generated.go -pkg httprecorder assets
```
