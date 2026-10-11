// Command stegno generates demo images and runs the watermarking pipeline.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"math"
	"math/rand"
	"os"
	"strings"

	"stegno/internal/imgio"
	"stegno/internal/steg"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "stegno:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: stegno <gen|embed|tamper|verify|recover|analyze> [flags]")
	}
	name := args[0]
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	in := fs.String("in", "", "input PNG or JPEG")
	out := fs.String("out", "", "output PNG (verify: red overlay)")
	key := fs.String("key", "stegno", "watermark key")
	width := fs.Int("w", 768, "gen: image width; tamper/analyze: rectangle width")
	height := fs.Int("h", 512, "gen: image height; tamper/analyze: rectangle height")
	kind := fs.String("kind", "gradient", "gen: gradient, testsrc or photo")
	x := fs.Int("x", 120, "tamper rectangle x")
	y := fs.Int("y", 80, "tamper rectangle y")
	mode := fs.String("mode", "black", "tamper mode: black, white, noise or copy")
	fillPath := fs.String("fill", "", "copy-mode fill image")
	origPath := fs.String("orig", "", "optional reference image for quality metrics")
	if name != "gen" && name != "embed" && name != "tamper" && name != "verify" && name != "recover" && name != "analyze" {
		return fmt.Errorf("unknown command %q (use gen, embed, tamper, verify, recover or analyze)", name)
	}
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %s", strings.Join(fs.Args(), " "))
	}
	if (name == "gen" || name == "embed" || name == "tamper" || name == "recover") && *out == "" {
		return errors.New("-out is required")
	}
	if name != "gen" && *in == "" {
		return errors.New("-in is required")
	}
	if name == "gen" {
		img, err := generate(*width, *height, *kind)
		if err != nil {
			return err
		}
		return imgio.SavePNG(*out, img)
	}
	if (name == "embed" || name == "tamper" || name == "recover") && *out == *in {
		return errors.New("-out must differ from -in")
	}
	src, err := imgio.Load(*in)
	if err != nil {
		return fmt.Errorf("load %s: %w", *in, err)
	}
	params := steg.DefaultParams()
	if name == "embed" || name == "verify" || name == "recover" || name == "analyze" {
		if *key == "" {
			return errors.New("key must not be empty")
		}
	}
	switch name {
	case "embed":
		img, err := steg.Embed(src, *key, params)
		if err != nil {
			return err
		}
		return imgio.SavePNG(*out, img)
	case "tamper":
		fill, err := loadOptional(*fillPath)
		if err != nil {
			return err
		}
		img, err := steg.Tamper(src, *x, *y, *width, *height, *mode, fill)
		if err != nil {
			return err
		}
		return imgio.SavePNG(*out, img)
	case "verify":
		report, err := steg.Verify(src, *key, params)
		if err != nil {
			return err
		}
		if *out != "" {
			if err := imgio.SavePNG(*out, steg.Overlay(src, report)); err != nil {
				return err
			}
		}
		return printJSON(report)
	case "recover":
		img, report, err := steg.Recover(src, *key, params)
		if err != nil {
			return err
		}
		if err := imgio.SavePNG(*out, img); err != nil {
			return err
		}
		result := struct {
			Report  *steg.Report `json:"report"`
			Metrics *quality     `json:"metrics"`
		}{Report: report}
		if *origPath != "" {
			ref, err := loadOptional(*origPath)
			if err != nil {
				return err
			}
			q, err := measure(ref, img)
			if err != nil {
				return err
			}
			result.Metrics = &q
		}
		return printJSON(result)
	case "analyze":
		ref := src
		if *origPath != "" {
			ref, err = loadOptional(*origPath)
			if err != nil {
				return err
			}
		}
		watermarked, err := steg.Embed(src, *key, params)
		if err != nil {
			return err
		}
		imperceptibility, err := measure(ref, watermarked)
		if err != nil {
			return err
		}
		tampered, err := steg.Tamper(watermarked, *x, *y, *width, *height, *mode, nil)
		if err != nil {
			return err
		}
		before, err := steg.Verify(tampered, *key, params)
		if err != nil {
			return err
		}
		restored, recoveryReport, err := steg.Recover(tampered, *key, params)
		if err != nil {
			return err
		}
		after, err := steg.Verify(restored, *key, params)
		if err != nil {
			return err
		}
		recoveryQuality, err := measure(ref, restored)
		if err != nil {
			return err
		}
		detection := score(before, *x, *y, *width, *height)
		result := struct {
			Image            string         `json:"image"`
			Imperceptibility quality        `json:"imperceptibility"`
			Detection        detectionScore `json:"detection"`
			Recovery         struct {
				PSNR              float64 `json:"psnr"`
				SSIM              float64 `json:"ssim"`
				RecoveredBlocks   int     `json:"recoveredBlocks"`
				UnrecoveredBlocks int     `json:"unrecoveredBlocks"`
			} `json:"recovery"`
			StillFlaggedAfterRecovery int            `json:"stillFlaggedAfterRecovery"`
			ReportBefore              reportSnapshot `json:"reportBefore"`
			ReportAfter               reportSnapshot `json:"reportAfter"`
		}{Image: *in, Imperceptibility: imperceptibility, Detection: detection,
			StillFlaggedAfterRecovery: after.Flagged, ReportBefore: snapshot(before), ReportAfter: snapshot(after)}
		result.Recovery.PSNR = recoveryQuality.PSNR
		result.Recovery.SSIM = recoveryQuality.SSIM
		result.Recovery.RecoveredBlocks = recoveryReport.Recovered
		result.Recovery.UnrecoveredBlocks = recoveryReport.Unrecovered
		return printJSON(result)
	}
	return nil
}

type quality struct {
	PSNR float64 `json:"psnr"`
	SSIM float64 `json:"ssim"`
}

func measure(a, b *image.RGBA) (quality, error) {
	psnr, err := steg.PSNR(a, b)
	if err != nil {
		return quality{}, err
	}
	ssim, err := steg.SSIM(a, b)
	return quality{psnr, ssim}, err
}

// reportSnapshot keeps the complete block classification as sparse indices rather than
// printing two 96x64 grids in the CLI's one-line analyze result. The indices are row-major
// and reconstruct the same tamper map as Report.TamperGrid.
type reportSnapshot struct {
	BlocksX        int   `json:"blocksX"`
	BlocksY        int   `json:"blocksY"`
	BlockSize      int   `json:"blockSize"`
	ImgWidth       int   `json:"imgWidth"`
	ImgHeight      int   `json:"imgHeight"`
	Total          int   `json:"totalBlocks"`
	Flagged        int   `json:"flaggedBlocks"`
	FlaggedIndices []int `json:"flaggedIndices"`
	CarrierLost    int   `json:"carrierLostBlocks"`
	Recovered      int   `json:"recoveredBlocks"`
	Unrecovered    int   `json:"unrecoveredBlocks"`
}

func snapshot(r *steg.Report) reportSnapshot {
	return reportSnapshot{r.BlocksX, r.BlocksY, r.BlockSize, r.ImgWidth, r.ImgHeight,
		r.Total, r.Flagged, append([]int{}, r.FlaggedIdx...), r.CarrierLost, r.Recovered, r.Unrecovered}
}

type detectionScore struct {
	TPR               float64 `json:"tpr"`
	FPR               float64 `json:"fpr"`
	Flagged           int     `json:"flagged"`
	Total             int     `json:"total"`
	GroundTruthBlocks int     `json:"groundTruthBlocks"`
}

func score(rep *steg.Report, x, y, w, h int) detectionScore {
	s := detectionScore{Flagged: rep.Flagged, Total: rep.Total}
	truePositive, falsePositive := 0, 0
	// Intersect the rectangle with each grid tile, including partial edge tiles.
	for by := 0; by < rep.BlocksY; by++ {
		for bx := 0; bx < rep.BlocksX; bx++ {
			x0, y0 := bx*rep.BlockSize, by*rep.BlockSize
			x1, y1 := min(x0+rep.BlockSize, rep.ImgWidth), min(y0+rep.BlockSize, rep.ImgHeight)
			groundTruth := x < x1 && x+w > x0 && y < y1 && y+h > y0
			flagged := rep.TamperMap[by*rep.BlocksX+bx]
			if groundTruth {
				s.GroundTruthBlocks++
				if flagged {
					truePositive++
				}
			} else if flagged {
				falsePositive++
			}
		}
	}
	if s.GroundTruthBlocks > 0 {
		s.TPR = float64(truePositive) / float64(s.GroundTruthBlocks)
	}
	if s.Total > s.GroundTruthBlocks {
		s.FPR = float64(falsePositive) / float64(s.Total-s.GroundTruthBlocks)
	}
	return s
}

func loadOptional(path string) (*image.RGBA, error) {
	if path == "" {
		return nil, nil
	}
	img, err := imgio.Load(path)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", path, err)
	}
	return img, nil
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	return enc.Encode(v)
}

func generate(w, h int, kind string) (*image.RGBA, error) {
	if w < 8 || h < 8 || w > 8192 || h > 8192 {
		return nil, fmt.Errorf("image dimensions must be between 8 and 8192, got %dx%d", w, h)
	}
	if kind != "gradient" && kind != "testsrc" && kind != "photo" {
		return nil, fmt.Errorf("unknown image kind %q (use gradient, testsrc or photo)", kind)
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	rng := rand.New(rand.NewSource(20261011))
	for y := 0; y < h; y++ {
		v := float64(y) / float64(h)
		for x := 0; x < w; x++ {
			u := float64(x) / float64(w)
			noise := float64(rng.Intn(9) - 4)
			var r, g, b float64
			switch kind {
			case "gradient":
				r = 35 + 182*u + 20*math.Sin(11*v+5*u)
				g = 48 + 145*v + 32*math.Sin(6*u+3*v)
				b = 65 + 106*(1-u) + 32*v
				if math.Pow((u-.28)/.13, 2)+math.Pow((v-.39)/.19, 2) < 1 {
					r, g, b = 216+18*u, 100+32*v, 74
				}
				if u > .59 && u < .86 && v > .29 && v < .73 {
					r, g, b = 62+75*u, 153-50*v, 186-44*u
				}
				if math.Abs(v-(.72-.28*u)) < .024 {
					r, g, b = 237, 215, 98
				}
			case "testsrc":
				bars := [][3]float64{{225, 225, 225}, {219, 212, 38}, {40, 209, 210}, {45, 195, 77}, {199, 50, 200}, {202, 49, 46}, {42, 43, 207}}
				c := bars[min(int(u*float64(len(bars))), len(bars)-1)]
				r, g, b = c[0]*(.7+.3*v), c[1]*(.7+.3*v), c[2]*(.7+.3*v)
				if v > .73 {
					r, g, b = 33+170*u, 35+130*u, 38+195*(1-u)
				}
				if math.Pow((u-.5)/.15, 2)+math.Pow((v-.52)/.18, 2) < 1 {
					r, g, b = 245, 236-45*u, 105+65*v
				}
			case "photo":
				// A deterministic, layered landscape with a soft sky, ridges, lake and sun.
				r, g, b = 58+154*v, 120+115*v, 175+50*v
				if math.Pow((u-.75)/.085, 2)+math.Pow((v-.22)/.13, 2) < 1 {
					r, g, b = 253, 224, 160
				}
				ridge := .48 + .09*math.Sin(11*u) + .04*math.Sin(29*u)
				if v > ridge {
					r, g, b = 55+42*u, 85+39*v, 98+35*u
				}
				if v > .61+.055*math.Sin(17*u+1) {
					r, g, b = 54+64*v, 91+66*u, 71+46*v
				}
				if v > .77 {
					r, g, b = 41+47*v+15*math.Sin(80*v+15*u), 103+31*u, 139+34*u
				}
				if u > .13 && u < .16 && v > .44 && v < .74 {
					r, g, b = 54, 42, 38
				}
				if math.Pow((u-.145)/.095, 2)+math.Pow((v-.41)/.14, 2) < 1 {
					r, g, b = 32+18*u, 77+20*v, 42
				}
			}
			i := img.PixOffset(x, y)
			img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = clamp(r+noise), clamp(g+noise), clamp(b+noise), 255
		}
	}
	return img, nil
}

func clamp(v float64) uint8 {
	return uint8(math.Max(0, math.Min(255, v)))
}
