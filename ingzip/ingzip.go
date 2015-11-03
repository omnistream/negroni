package ingzip

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"strings"
)

// These compression constants are copied from the compress/gzip package.
const (
	encodingGzip = "gzip"

	headerContentEncoding = "Content-Encoding"
	headerContentLength   = "Content-Length"
	headerSecWebSocketKey = "Sec-WebSocket-Key"
)

// Gzip returns a handler which will handle the Gzip compression in ServeHTTP.
// Valid values for level are identical to those in the compress/gzip package.
func InGzip() *handler {
	h := &handler{}
	return h
}

// handler struct contains the ServeHTTP method
type handler struct {
}

// ServeHTTP wraps the http.ResponseWriter with a gzip.Writer.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// Skip compression if the client doesn't accept gzip encoding.
	if !strings.Contains(r.Header.Get(headerContentEncoding), encodingGzip) {
		next(w, r)
		return
	}

	// Skip compression if client attempt WebSocket connection
	if len(r.Header.Get(headerSecWebSocketKey)) > 0 {
		next(w, r)
		return
	}

	// Read compressed Body
	if r.Body == nil {
		next(w, r)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		next(w, r)
		return
	}

	raw := bytes.NewBuffer(data)
	unz, err := gzip.NewReader(raw)
	if err != nil {
		next(w, r)
		return
	}
	buf, err := ioutil.ReadAll(unz)
	if err != nil {
		next(w, r)
		return
	}
	unz.Close()

	//replace body of current request with un-compressed version
	r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))

	//remove content-legth and content-encoding
	r.Header.Del(headerContentLength)
	r.Header.Del(headerContentEncoding)

	// Call the next handler supplying the gzipResponseWriter instead of
	// the original.
	next(w, r)

}
