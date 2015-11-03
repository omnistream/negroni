package ingzip

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	gzipTestString              = "Foobar Wibble Content"
	gzipNoContent               = "no-content"
	gzipTestWebSocketKey        = "Test"
	gzipInvalidCompressionLevel = 11
)

func testHTTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		body, _ := ioutil.ReadAll(r.Body)
		fmt.Fprintf(w, string(body))
	} else {
		fmt.Fprintf(w, string(gzipNoContent))
	}
}

func getCompressed() []byte {
	var b bytes.Buffer
	g := gzip.NewWriter(&b)
	g.Write([]byte(gzipTestString))
	g.Close()
	return b.Bytes()
}

func getUnCompressed() []byte {
	return []byte(gzipTestString)
}

func Test_ServeHTTP_Compressed_get(t *testing.T) {
	gzipHandler := InGzip()
	w := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "http://localhost/foobar", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set(headerContentEncoding, encodingGzip)

	gzipHandler.ServeHTTP(w, req, testHTTPHandler)

	body, _ := ioutil.ReadAll(w.Body)

	log.Println("BODY: ", string(body), string(body) != gzipNoContent)
	if string(body) != gzipNoContent {
		t.Fail()
	}
}

func Test_ServeHTTP_Compressed_post(t *testing.T) {
	gzipHandler := InGzip()
	w := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "http://localhost/foobar", bytes.NewReader(getCompressed()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set(headerContentEncoding, encodingGzip)

	gzipHandler.ServeHTTP(w, req, testHTTPHandler)

	body, _ := ioutil.ReadAll(w.Body)

	log.Println("BODY: ", string(body), string(body) != gzipTestString)
	if string(body) != gzipTestString {
		t.Fail()
	}
}

func Test_ServeHTTP_NoCompression(t *testing.T) {
	gzipHandler := InGzip()
	w := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "http://localhost/foobar", bytes.NewReader(getUnCompressed()))
	if err != nil {
		t.Fatal(err)
	}

	gzipHandler.ServeHTTP(w, req, testHTTPHandler)

	body := w.Body.String()
	log.Println("BODY: ", string(body))
	if body != gzipTestString {
		t.Fail()
	}
}

func Test_ServeHTTP_CompressionWithNoGzipHeader(t *testing.T) {
	gzipHandler := InGzip()
	w := httptest.NewRecorder()

	var compressed = getUnCompressed()
	req, err := http.NewRequest("POST", "http://localhost/foobar", bytes.NewReader(compressed))
	if err != nil {
		t.Fatal(err)
	}

	gzipHandler.ServeHTTP(w, req, testHTTPHandler)

	body := w.Body.String()
	log.Println("BODY: ", string(body), string(body) != string(compressed))
	if body != string(compressed) {
		t.Fail()
	}
}

func Test_ServeHTTP_WebSocketConnection(t *testing.T) {
	gzipHandler := InGzip()
	w := httptest.NewRecorder()

	var compressed = getUnCompressed()
	req, err := http.NewRequest("POST", "http://localhost/foobar", bytes.NewReader(compressed))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set(headerContentEncoding, encodingGzip)
	req.Header.Set(headerSecWebSocketKey, gzipTestWebSocketKey)

	gzipHandler.ServeHTTP(w, req, testHTTPHandler)

	body := w.Body.String()
	log.Println("BODY: ", string(body), string(body) != string(compressed))
	if body != string(compressed) {
		t.Fail()
	}
}

func Benchmark_ServeHTTPCompressed(b *testing.B) {

	b.StopTimer()
	b.ReportAllocs()

	gzipHandler := InGzip()
	req, err := http.NewRequest("POST", "http://localhost/foobar", bytes.NewReader(getCompressed()))
	if err != nil {
		b.Fatal(err)
	}
	req.Header.Set(headerContentEncoding, encodingGzip)

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		gzipHandler.ServeHTTP(w, req, testHTTPHandler)
	}

}

func Benchmark_ServeHTTPUnCompressed(b *testing.B) {

	b.StopTimer()
	b.ReportAllocs()

	gzipHandler := InGzip()
	req, err := http.NewRequest("POST", "http://localhost/foobar", bytes.NewReader(getUnCompressed()))
	if err != nil {
		b.Fatal(err)
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		gzipHandler.ServeHTTP(w, req, testHTTPHandler)
	}
}
