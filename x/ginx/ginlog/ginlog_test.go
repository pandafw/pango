package ginlog

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pandafw/pango/str"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func assertContains(t *testing.T, msg string, body string, ss ...string) {
	for _, s := range ss {
		if !str.Contains(body, s) {
			t.Errorf(`%s response body not contains %q`, msg, s)
		}
	}
}

func TestTextLog(t *testing.T) {
	buffer := new(bytes.Buffer)
	router := gin.New()

	writer := io.MultiWriter(buffer, os.Stdout)
	router.Use(New(writer, DefaultTextLogFormat).Handler())

	router.Any("/example", func(c *gin.Context) {})

	buffer.Reset()
	performRequest(router, "GET", "/example?a=100")
	assertContains(t, "GET /example?a=100", buffer.String(), "200", "GET", "/example", "a=100")

	buffer.Reset()
	performRequest(router, "POST", "/example")
	assertContains(t, "POST /example", buffer.String(), "200", "POST", "/example")

	buffer.Reset()
	performRequest(router, "PUT", "/example")
	assertContains(t, "PUT /example", buffer.String(), "200", "PUT", "/example")

	buffer.Reset()
	performRequest(router, "DELETE", "/example")
	assertContains(t, "DELETE /example", buffer.String(), "200", "DELETE", "/example")

	buffer.Reset()
	performRequest(router, "PATCH", "/example")
	assertContains(t, "PATCH /example", buffer.String(), "200", "PATCH", "/example")

	buffer.Reset()
	performRequest(router, "HEAD", "/example")
	assertContains(t, "HEAD /example", buffer.String(), "200", "HEAD", "/example")

	buffer.Reset()
	performRequest(router, "OPTIONS", "/example")
	assertContains(t, "OPTIONS /example", buffer.String(), "200", "OPTIONS", "/example")

	buffer.Reset()
	performRequest(router, "GET", "/notfound")
	assertContains(t, "GET /notfound", buffer.String(), "404", "GET", "/notfound")
}

func assertJsonResult(t *testing.T, result map[string]interface{}, sc int, method string, url string) {
	if result["status"] != float64(sc) {
		t.Errorf("status = %v, want %v", result["status"], sc)
	}
	if result["method"] != method {
		t.Errorf("method = %v, want %v", result["method"], sc)
	}
	if result["url"] != url {
		t.Errorf("url = %v, want %v", result["url"], sc)
	}
}

func TestJSONLog(t *testing.T) {
	result := make(map[string]interface{})
	buffer := new(bytes.Buffer)
	router := gin.New()

	writer := io.MultiWriter(buffer, os.Stdout)
	router.Use(New(writer, DefaultJSONLogFormat).Handler())

	router.Any("/example", func(c *gin.Context) {})

	buffer.Reset()
	performRequest(router, "GET", "/example?a=100")
	json.Unmarshal(buffer.Bytes(), &result)
	assertJsonResult(t, result, 200, "GET", "/example?a=100")

	buffer.Reset()
	performRequest(router, "POST", "/example")
	json.Unmarshal(buffer.Bytes(), &result)
	assertJsonResult(t, result, 200, "POST", "/example")

	buffer.Reset()
	performRequest(router, "PUT", "/example")
	json.Unmarshal(buffer.Bytes(), &result)
	assertJsonResult(t, result, 200, "PUT", "/example")

	buffer.Reset()
	performRequest(router, "DELETE", "/example")
	json.Unmarshal(buffer.Bytes(), &result)
	assertJsonResult(t, result, 200, "DELETE", "/example")

	buffer.Reset()
	performRequest(router, "PATCH", "/example")
	json.Unmarshal(buffer.Bytes(), &result)
	assertJsonResult(t, result, 200, "PATCH", "/example")

	buffer.Reset()
	performRequest(router, "HEAD", "/example")
	json.Unmarshal(buffer.Bytes(), &result)
	assertJsonResult(t, result, 200, "HEAD", "/example")

	buffer.Reset()
	performRequest(router, "OPTIONS", "/example")
	json.Unmarshal(buffer.Bytes(), &result)
	assertJsonResult(t, result, 200, "OPTIONS", "/example")

	buffer.Reset()
	performRequest(router, "GET", "/notfound")
	json.Unmarshal(buffer.Bytes(), &result)
	assertJsonResult(t, result, 404, "GET", "/notfound")
}

func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
