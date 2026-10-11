package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sampleImage(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 128, 128))
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			img.SetRGBA(x, y, color.RGBA{uint8(35 + x/2), uint8(80 + y/2), uint8(100 + (x+y)/4), 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func request(t *testing.T, h http.Handler, path string, fields map[string]string, files map[string][]byte) map[string]json.RawMessage {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := mw.WriteField(k, v); err != nil {
			t.Fatal(err)
		}
	}
	for k, data := range files {
		fw, err := mw.CreateFormFile(k, k+".png")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodPost, path, &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("%s returned %d: %s", path, w.Code, w.Body.String())
	}
	var v map[string]json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	return v
}

func pngFromURL(t *testing.T, raw json.RawMessage) []byte {
	t.Helper()
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(s, "data:image/png;base64,") {
		t.Fatalf("not a PNG data URL: %.40s", s)
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(s, "data:image/png;base64,"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := png.Decode(bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}
	return data
}

func number(t *testing.T, raw json.RawMessage, field string) float64 {
	t.Helper()
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatal(err)
	}
	var n float64
	if err := json.Unmarshal(obj[field], &n); err != nil {
		t.Fatalf("%s: %v", field, err)
	}
	return n
}

func TestHealthAndCORS(t *testing.T) {
	h := New(Config{})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if w.Code != 200 || strings.TrimSpace(w.Body.String()) != `{"status":"ok","version":"1.0"}` {
		t.Fatalf("health: %d %s", w.Code, w.Body.String())
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("CORS origin missing")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodOptions, "/api/embed", nil))
	if w.Code != http.StatusNoContent || w.Header().Get("Access-Control-Allow-Methods") != "GET, POST, OPTIONS" || w.Header().Get("Access-Control-Allow-Headers") != "*" {
		t.Fatalf("OPTIONS: %d %v", w.Code, w.Header())
	}
}

func TestPipeline(t *testing.T) {
	h := New(Config{})
	original := sampleImage(t)
	embedded := request(t, h, "/api/embed", map[string]string{"key": "demo"}, map[string][]byte{"image": original})
	watermarked := pngFromURL(t, embedded["watermarked"])
	verified := request(t, h, "/api/verify", map[string]string{"key": "demo"}, map[string][]byte{"image": watermarked})
	if n := number(t, verified["report"], "flaggedBlocks"); n != 0 {
		t.Fatalf("clean image flagged: %v", n)
	}
	pngFromURL(t, verified["overlay"])

	region := map[string]string{"key": "demo", "x": "32", "y": "32", "w": "32", "h": "32", "mode": "black"}
	tampered := request(t, h, "/api/tamper", region, map[string][]byte{"image": watermarked})
	bad := pngFromURL(t, tampered["tampered"])
	checked := request(t, h, "/api/verify", map[string]string{"key": "demo"}, map[string][]byte{"image": bad})
	if n := number(t, checked["report"], "flaggedBlocks"); n <= 0 {
		t.Fatalf("tamper not flagged: %v", n)
	}
	recovered := request(t, h, "/api/recover", map[string]string{"key": "demo"}, map[string][]byte{"image": bad, "original": original})
	pngFromURL(t, recovered["recovered"])
	if n := number(t, recovered["metrics"], "psnr"); n <= 25 {
		t.Fatalf("recovery PSNR too low: %v", n)
	}
	withoutOriginal := request(t, h, "/api/recover", map[string]string{"key": "demo"}, map[string][]byte{"image": bad})
	if string(withoutOriginal["metrics"]) != "null" {
		t.Fatalf("unexpected metrics: %s", withoutOriginal["metrics"])
	}

	analysis := request(t, h, "/api/analyze", region, map[string][]byte{"image": original})
	for _, field := range []string{"original", "watermarked", "tampered", "overlay", "recovered"} {
		pngFromURL(t, analysis[field])
	}
	var detection map[string]json.RawMessage
	if err := json.Unmarshal(analysis["detection"], &detection); err != nil {
		t.Fatal(err)
	}
	if n := number(t, analysis["detection"], "tpr"); n <= 0.9 {
		t.Fatalf("TPR too low: %v", n)
	}
	for _, field := range []string{"fpr", "flagged", "total", "groundTruthBlocks"} {
		if _, ok := detection[field]; !ok {
			t.Fatalf("missing detection.%s", field)
		}
	}
	if n := number(t, analysis["recovery"], "psnr"); n <= 25 {
		t.Fatalf("analysis recovery PSNR too low: %v", n)
	}
	for _, field := range []string{"psnr", "ssim"} {
		number(t, analysis["imperceptibility"], field)
	}
	for _, field := range []string{"ssim", "recoveredBlocks", "unrecoveredBlocks"} {
		number(t, analysis["recovery"], field)
	}
	var reports map[string]json.RawMessage
	if err := json.Unmarshal(analysis["report"], &reports); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"reportBefore", "reportAfter", "tamperMapBefore", "tamperMapAfter", "flaggedAfterRecovery"} {
		if _, ok := reports[field]; !ok {
			t.Fatalf("missing report.%s", field)
		}
	}
	if n := number(t, analysis["report"], "flaggedAfterRecovery"); n < 0 {
		t.Fatalf("invalid flaggedAfterRecovery: %v", n)
	}
}

func TestErrorsAndStatic(t *testing.T) {
	h := New(Config{})
	for _, path := range []string{"/", "/api/missing"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != 404 || !strings.Contains(w.Body.String(), `"error"`) {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body.String())
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/embed", nil))
	if w.Code != 405 || !strings.Contains(w.Body.String(), `"error"`) {
		t.Fatalf("method: %d %s", w.Code, w.Body.String())
	}
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("image", "sample.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write(sampleImage(t)); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodPost, "/api/analyze", bytes.NewReader(buf.Bytes()))
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 400 {
		t.Fatalf("empty analyze key: %d %s", w.Code, w.Body.String())
	}
	h = New(Config{MaxUploadBytes: 100})
	r = httptest.NewRequest(http.MethodPost, "/api/embed", bytes.NewReader(buf.Bytes()))
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 413 {
		t.Fatalf("upload limit: %d %s", w.Code, w.Body.String())
	}
}

func TestStaticFiles(t *testing.T) {
	base := t.TempDir()
	web := filepath.Join(base, "public")
	if err := os.Mkdir(web, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "secret"), []byte("SECRET"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(web, "index.html"), []byte("<h1>Stegno</h1>"), 0600); err != nil {
		t.Fatal(err)
	}
	h := New(Config{WebDir: web})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "Stegno") {
		t.Fatalf("static: %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/missing.html", nil))
	if w.Code != 404 || !strings.Contains(w.Body.String(), `"error"`) {
		t.Fatalf("missing static: %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/../secret", nil))
	if w.Code != 404 {
		t.Fatalf("path escape: %d %s", w.Code, w.Body.String())
	}
}
