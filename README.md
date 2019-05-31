Forked from <https://godoc.org/github.com/tim-st/go-zim>.

# go-zim
Package `zim` implements reading support for the ZIM File Format.

Documentation at <https://godoc.org/github.com/tim-st/go-zim>.

Download and install package `zim` with `go get -u github.com/tim-st/go-zim/...`

If you want to try the `zimserver` tool, install it with `go install github.com/tim-st/go-zim/cmd/zimserver`

If you want to extract sentences or texts from a Wikipedia ZIM file use `zimtext` tool, install it with `go install github.com/tim-st/go-zim/cmd/zimtext`

You can download a ZIM file for testing [here](https://download.kiwix.org/zim/).

# reMarkable support
Build with
```GOARCH=arm GOOS=linux go build -o zimserver github.com/dps/go-zim/cmd/zimserver```
