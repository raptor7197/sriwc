package api

import (
	"fmt"
	"image"
	"math"
	"net/http"
	"strconv"

	"stegno/internal/imgio"
	"stegno/internal/steg"
)

type metrics struct {
	PSNR float64 `json:"psnr"`
	SSIM float64 `json:"ssim"`
}

type region struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

func upload(r *http.Request, field string, required bool) (*image.RGBA, error) {
	f, _, err := r.FormFile(field)
	if err != nil {
		if !required && err == http.ErrMissingFile {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w", field, err)
	}
	defer f.Close()
	img, format, err := imgio.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", field, err)
	}
	if format != "png" && format != "jpeg" {
		return nil, fmt.Errorf("%s: use PNG or JPEG", field)
	}
	return img, nil
}

func imageURL(w http.ResponseWriter, img *image.RGBA) (string, bool) {
	url, err := imgio.DataURL(img)
	if err != nil {
		jsonError(w, 500, "encode PNG: "+err.Error())
		return "", false
	}
	return url, true
}

func measure(a, b *image.RGBA) (metrics, error) {
	psnr, err := steg.PSNR(a, b)
	if err != nil {
		return metrics{}, err
	}
	ssim, err := steg.SSIM(a, b)
	if err != nil {
		return metrics{}, err
	}
	if math.IsNaN(psnr) || math.IsInf(psnr, 0) || math.IsNaN(ssim) || math.IsInf(ssim, 0) {
		return metrics{}, fmt.Errorf("metrics are not finite")
	}
	return metrics{psnr, ssim}, nil
}

func key(r *http.Request) string {
	if k := r.FormValue("key"); k != "" {
		return k
	}
	return "stegno"
}

func parseRegion(r *http.Request) (region, error) {
	var v region
	for _, item := range []struct {
		name string
		dest *int
	}{{"x", &v.X}, {"y", &v.Y}, {"w", &v.W}, {"h", &v.H}} {
		text := r.FormValue(item.name)
		if text == "" {
			return v, fmt.Errorf("missing %s", item.name)
		}
		n, err := strconv.Atoi(text)
		if err != nil {
			return v, fmt.Errorf("invalid %s: %w", item.name, err)
		}
		*item.dest = n
	}
	return v, nil
}

func (s *server) embed(w http.ResponseWriter, r *http.Request) {
	if !s.parse(w, r) {
		return
	}
	img, err := upload(r, "image", true)
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}
	p := steg.DefaultParams()
	marked, err := steg.Embed(img, key(r), p)
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}
	m, err := measure(img, marked)
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}
	url, ok := imageURL(w, marked)
	if !ok {
		return
	}
	jsonResponse(w, struct {
		Watermarked string      `json:"watermarked"`
		PSNR        float64     `json:"psnr"`
		SSIM        float64     `json:"ssim"`
		Params      steg.Params `json:"params"`
	}{url, m.PSNR, m.SSIM, p})
}

func (s *server) tamper(w http.ResponseWriter, r *http.Request) {
	if !s.parse(w, r) {
		return
	}
	img, err := upload(r, "image", true)
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}
	v, err := parseRegion(r)
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}
	fill, err := upload(r, "fill", false)
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}
	changed, err := steg.Tamper(img, v.X, v.Y, v.W, v.H, r.FormValue("mode"), fill)
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}
	url, ok := imageURL(w, changed)
	if !ok {
		return
	}
	jsonResponse(w, struct {
		Tampered string `json:"tampered"`
		Region   region `json:"region"`
	}{url, v})
}

func (s *server) verify(w http.ResponseWriter, r *http.Request) {
	if !s.parse(w, r) {
		return
	}
	img, err := upload(r, "image", true)
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}
	rep, err := steg.Verify(img, key(r), steg.DefaultParams())
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}
	url, ok := imageURL(w, steg.Overlay(img, rep))
	if !ok {
		return
	}
	jsonResponse(w, struct {
		Report  *steg.Report `json:"report"`
		Overlay string       `json:"overlay"`
	}{rep, url})
}

func (s *server) recoverImage(w http.ResponseWriter, r *http.Request) {
	if !s.parse(w, r) {
		return
	}
	img, err := upload(r, "image", true)
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}
	original, err := upload(r, "original", false)
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}
	result, rep, err := steg.Recover(img, key(r), steg.DefaultParams())
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}
	var m *metrics
	if original != nil {
		measured, err := measure(original, result)
		if err != nil {
			jsonError(w, 400, err.Error())
			return
		}
		m = &measured
	}
	url, ok := imageURL(w, result)
	if !ok {
		return
	}
	jsonResponse(w, struct {
		Recovered string       `json:"recovered"`
		Report    *steg.Report `json:"report"`
		Metrics   *metrics     `json:"metrics"`
	}{url, rep, m})
}

func (s *server) analyze(w http.ResponseWriter, r *http.Request) {
	if !s.parse(w, r) {
		return
	}
	if r.FormValue("key") == "" {
		jsonError(w, 400, "key must not be empty")
		return
	}
	original, err := upload(r, "image", true)
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}
	v, err := parseRegion(r)
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}
	fill, err := upload(r, "fill", false)
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}
	p := steg.DefaultParams()
	marked, err := steg.Embed(original, r.FormValue("key"), p)
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}
	changed, err := steg.Tamper(marked, v.X, v.Y, v.W, v.H, r.FormValue("mode"), fill)
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}
	before, err := steg.Verify(changed, r.FormValue("key"), p)
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}
	restored, recoveryReport, err := steg.Recover(changed, r.FormValue("key"), p)
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}
	after, err := steg.Verify(restored, r.FormValue("key"), p)
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}
	imperceptibility, err := measure(original, marked)
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}
	recoveryMetrics, err := measure(original, restored)
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}

	groundTruth, truePositive, falsePositive := 0, 0, 0
	// A tile is ground truth iff its image-space rectangle intersects the requested region.
	for by := 0; by < before.BlocksY; by++ {
		for bx := 0; bx < before.BlocksX; bx++ {
			hit := bx*p.BlockSize < v.X+v.W && (bx+1)*p.BlockSize > v.X && by*p.BlockSize < v.Y+v.H && (by+1)*p.BlockSize > v.Y
			flagged := before.TamperMap[by*before.BlocksX+bx]
			if hit {
				groundTruth++
				if flagged {
					truePositive++
				}
			} else if flagged {
				falsePositive++
			}
		}
	}
	detection := struct {
		TPR               float64 `json:"tpr"`
		FPR               float64 `json:"fpr"`
		Flagged           int     `json:"flagged"`
		Total             int     `json:"total"`
		GroundTruthBlocks int     `json:"groundTruthBlocks"`
	}{Flagged: before.Flagged, Total: before.Total, GroundTruthBlocks: groundTruth}
	if groundTruth > 0 {
		detection.TPR = float64(truePositive) / float64(groundTruth)
	}
	if other := before.Total - groundTruth; other > 0 {
		detection.FPR = float64(falsePositive) / float64(other)
	}

	urls := make([]string, 5)
	for i, img := range []*image.RGBA{original, marked, changed, steg.Overlay(changed, before), restored} {
		var ok bool
		urls[i], ok = imageURL(w, img)
		if !ok {
			return
		}
	}
	jsonResponse(w, struct {
		Original         string  `json:"original"`
		Watermarked      string  `json:"watermarked"`
		Tampered         string  `json:"tampered"`
		Overlay          string  `json:"overlay"`
		Recovered        string  `json:"recovered"`
		Imperceptibility metrics `json:"imperceptibility"`
		Detection        any     `json:"detection"`
		Recovery         any     `json:"recovery"`
		Report           any     `json:"report"`
	}{urls[0], urls[1], urls[2], urls[3], urls[4], imperceptibility, detection,
		struct {
			PSNR              float64 `json:"psnr"`
			SSIM              float64 `json:"ssim"`
			RecoveredBlocks   int     `json:"recoveredBlocks"`
			UnrecoveredBlocks int     `json:"unrecoveredBlocks"`
		}{recoveryMetrics.PSNR, recoveryMetrics.SSIM, recoveryReport.Recovered, recoveryReport.Unrecovered},
		struct {
			ReportBefore         *steg.Report `json:"reportBefore"`
			ReportAfter          *steg.Report `json:"reportAfter"`
			TamperMapBefore      [][]int      `json:"tamperMapBefore"`
			TamperMapAfter       [][]int      `json:"tamperMapAfter"`
			FlaggedAfterRecovery int          `json:"flaggedAfterRecovery"`
		}{before, after, before.TamperGrid, after.TamperGrid, after.Flagged},
	})
}
