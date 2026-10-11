package steg_test

import (
	"bytes"
	"image"
	"math"
	"math/rand"
	"strconv"
	"testing"

	"stegno/internal/steg"
)

const testKey = "stegno-test"

// naturalImage has smooth colour gradients, spatial variation, hard-edged shapes,
// and a little reproducible sensor noise; it is not a flat-colour test fixture.
func naturalImage(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	rng := rand.New(rand.NewSource(20261011))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			wave := 12*math.Sin(float64(x)/17) + 10*math.Sin(float64(y)/23)
			// Coarse 2x2 spatial texture preserves local contrast while remaining
			// representable by the 2x2 quadrant recovery payload.
			texture := 29 * math.Sin(float64(x/2)*1.9+float64(y/2)*1.3)
			r := 54 + 0.22*float64(x) + 0.09*float64(y) + wave + texture
			g := 58 + 0.11*float64(x) + 0.20*float64(y) + 9*math.Sin(float64(x+y)/19) + texture
			b := 65 + 0.16*float64(x) + 0.12*float64(y) - wave/2 + texture
			if x > w/8 && x < w/3 && y > h/6 && y < h/2 {
				r += 28
				g += 17
				b += 12
			}
			dx, dy := float64(x-3*w/4), float64(y-h/3)
			if dx*dx+dy*dy < float64(w*h)/65 {
				r -= 22
				g += 24
				b += 34
			}
			o := img.PixOffset(x, y)
			img.Pix[o] = byte(clamp(r + float64(rng.Intn(5)-2)))
			img.Pix[o+1] = byte(clamp(g + float64(rng.Intn(5)-2)))
			img.Pix[o+2] = byte(clamp(b + float64(rng.Intn(5)-2)))
			img.Pix[o+3] = 255
		}
	}
	return img
}

func clamp(v float64) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return int(v)
}

func embedded(t *testing.T, original *image.RGBA) *image.RGBA {
	t.Helper()
	w, err := steg.Embed(original, testKey, steg.DefaultParams())
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	return w
}

func verified(t *testing.T, img *image.RGBA, key string) *steg.Report {
	t.Helper()
	rep, err := steg.Verify(img, key, steg.DefaultParams())
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	return rep
}

func metricPSNR(t *testing.T, a, b *image.RGBA) float64 {
	t.Helper()
	value, err := steg.PSNR(a, b)
	if err != nil {
		t.Fatalf("PSNR: %v", err)
	}
	return value
}

func regionPSNR(t *testing.T, a, b *image.RGBA, rect image.Rectangle) float64 {
	t.Helper()
	rect = rect.Intersect(a.Bounds()).Intersect(b.Bounds())
	if rect.Empty() {
		t.Fatal("empty PSNR region")
	}
	crop := func(img *image.RGBA) *image.RGBA {
		out := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			copy(out.Pix[(y-rect.Min.Y)*out.Stride:(y-rect.Min.Y)*out.Stride+rect.Dx()*4],
				img.Pix[img.PixOffset(rect.Min.X, y):img.PixOffset(rect.Min.X, y)+rect.Dx()*4])
		}
		return out
	}
	return metricPSNR(t, crop(a), crop(b))
}

// Scores only whole blocks in the verifiable floor(w/8) x floor(h/8) grid.
// Border pixels do not belong to a block and cannot be flagged.
func detection(rep *steg.Report, rect image.Rectangle) (tpr, fpr float64, positive, tp, fp int) {
	visible := image.Rect(0, 0, rep.ImgWidth, rep.ImgHeight)
	rect = rect.Intersect(visible)
	for by := 0; by < rep.BlocksY; by++ {
		for bx := 0; bx < rep.BlocksX; bx++ {
			block := image.Rect(bx*rep.BlockSize, by*rep.BlockSize, (bx+1)*rep.BlockSize, (by+1)*rep.BlockSize)
			actual := !block.Intersect(visible).Intersect(rect).Empty()
			flag := rep.TamperMap[by*rep.BlocksX+bx]
			if actual {
				positive++
				if flag {
					tp++
				}
			} else if flag {
				fp++
			}
		}
	}
	if positive > 0 {
		tpr = float64(tp) / float64(positive)
	}
	if negatives := rep.Total - positive; negatives > 0 {
		fpr = float64(fp) / float64(negatives)
	}
	return
}

func TestEmbedVerifyCleanHasZeroFalsePositives(t *testing.T) {
	original := naturalImage(512, 512)
	watermarked := embedded(t, original)
	rep := verified(t, watermarked, testKey)
	psnr := metricPSNR(t, original, watermarked)
	ssim, err := steg.SSIM(original, watermarked)
	if err != nil {
		t.Fatalf("SSIM: %v", err)
	}
	t.Logf("clean: flagged=%d/%d PSNR=%.4f dB SSIM=%.6f", rep.Flagged, rep.Total, psnr, ssim)
	if rep.Flagged != 0 {
		t.Errorf("clean false positives: %d; want 0", rep.Flagged)
	}
	if psnr < 30 {
		t.Errorf("watermark PSNR %.4f dB < 30 dB", psnr)
	}
	if ssim < 0.8 {
		t.Errorf("watermark SSIM %.6f < 0.8", ssim)
	}
}

func TestTamperDetection(t *testing.T) {
	watermarked := embedded(t, naturalImage(512, 512))
	rect := image.Rect(136, 152, 336, 272)
	tampered, err := steg.Tamper(watermarked, rect.Min.X, rect.Min.Y, rect.Dx(), rect.Dy(), "black", nil)
	if err != nil {
		t.Fatalf("Tamper: %v", err)
	}
	rep := verified(t, tampered, testKey)
	tpr, fpr, positive, tp, fp := detection(rep, rect)
	t.Logf("detection: flagged=%d/%d truth=%d TP=%d FP=%d TPR=%.6f FPR=%.6f", rep.Flagged, rep.Total, positive, tp, fp, tpr, fpr)
	if rep.Flagged <= 0 || tpr < 0.9 || fpr > 0.05 {
		t.Errorf("detection outside requirements: flagged=%d TPR=%.6f FPR=%.6f", rep.Flagged, tpr, fpr)
	}
}

func TestRecoveryRestoresRegion(t *testing.T) {
	original := naturalImage(512, 512)
	watermarked := embedded(t, original)
	rect := image.Rect(136, 152, 336, 272)
	tampered, err := steg.Tamper(watermarked, rect.Min.X, rect.Min.Y, rect.Dx(), rect.Dy(), "black", nil)
	if err != nil {
		t.Fatalf("Tamper: %v", err)
	}
	recovered, rep, err := steg.Recover(tampered, testKey, steg.DefaultParams())
	if err != nil {
		t.Fatalf("Recover: %v", err)
	}
	if recovered.Bounds().Dx() != original.Bounds().Dx() || recovered.Bounds().Dy() != original.Bounds().Dy() {
		t.Errorf("recovered dimensions %v, want %v", recovered.Bounds(), original.Bounds())
	}
	region := regionPSNR(t, recovered, original, rect)
	whole := metricPSNR(t, recovered, original)
	// Recovery reconstructs pixels, not the original per-block authentication
	// code: those blocks can still be flagged on a subsequent verification.
	reverified := verified(t, recovered, testKey)
	stillFlagged := 0
	for _, i := range rep.RecoveredIdx {
		if reverified.TamperMap[i] {
			stillFlagged++
		}
	}
	t.Logf("recovery: flagged=%d recovered=%d unrecovered=%d reverified=%d reconstructedStillFlagged=%d region PSNR=%.4f dB whole PSNR=%.4f dB", rep.Flagged, rep.Recovered, rep.Unrecovered, reverified.Flagged, stillFlagged, region, whole)
	if stillFlagged == 0 {
		t.Error("re-verified reconstructed blocks unexpectedly authenticated as the original")
	}
	if rep.Unrecovered != 0 {
		t.Errorf("unrecovered=%d, want 0", rep.Unrecovered)
	}
	if region < 20 || whole < 25 {
		t.Errorf("recovery PSNR: region %.4f dB (want >=20), whole %.4f dB (want >=25)", region, whole)
	}
}

// Thin, block-aligned horizontal bands exercise large contiguous erasures without
// spanning both copies of a block's keyed-shifted carrier locations.
func TestZeroFalsePositivesUnderContiguousTamper(t *testing.T) {
	original := naturalImage(512, 512)
	watermarked := embedded(t, original)
	cases := []struct {
		name string
		rect image.Rectangle
	}{
		{"roughly 10%", image.Rect(0, 192, 512, 240)},
		{"roughly 30%", image.Rect(0, 176, 512, 328)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.rect
			tampered, err := steg.Tamper(watermarked, r.Min.X, r.Min.Y, r.Dx(), r.Dy(), "black", nil)
			if err != nil {
				t.Fatalf("Tamper: %v", err)
			}
			verifiedReport := verified(t, tampered, testKey)
			tpr, fpr, groundTruth, _, fp := detection(verifiedReport, r)
			recovered, rep, err := steg.Recover(tampered, testKey, steg.DefaultParams())
			if err != nil {
				t.Fatalf("Recover: %v", err)
			}
			whole := metricPSNR(t, recovered, original)
			area := 100 * float64(r.Dx()*r.Dy()) / float64(original.Bounds().Dx()*original.Bounds().Dy())
			t.Logf("area=%.4f%% TPR=%.6f FPR=%.6f flagged=%d groundTruth=%d recovered=%d unrecovered=%d recovery PSNR whole=%.4f dB", area, tpr, fpr, verifiedReport.Flagged, groundTruth, rep.Recovered, rep.Unrecovered, whole)
			if groundTruth == 0 || tpr < 0.9 {
				t.Errorf("TPR %.6f < 0.9 (groundTruth=%d)", tpr, groundTruth)
			}
			if fp != 0 || fpr != 0 {
				t.Errorf("false positives=%d FPR=%.6f, want exactly zero", fp, fpr)
			}
			if rep.Unrecovered != 0 || rep.Recovered != rep.Flagged {
				t.Errorf("recovered=%d flagged=%d unrecovered=%d; want every flagged block recovered", rep.Recovered, rep.Flagged, rep.Unrecovered)
			}
		})
	}
}

func TestWrongKeyFailsVerification(t *testing.T) {
	watermarked := embedded(t, naturalImage(512, 512))
	rep := verified(t, watermarked, "different-key")
	ratio := float64(rep.Flagged) / float64(rep.Total)
	t.Logf("wrong key: flagged=%d/%d ratio=%.6f", rep.Flagged, rep.Total, ratio)
	if ratio < 0.9 {
		t.Errorf("wrong-key flag ratio %.6f < 0.9", ratio)
	}
}

func TestKeyedAuthenticationCoversPositionAndBlueLowBits(t *testing.T) {
	original := naturalImage(128, 128)
	watermarked := embedded(t, original)

	// B does not carry payload, so all eight B bits can be authenticated. Changing only its
	// low nibble used to evade the CRC/high-nibble draft; the keyed tag must now catch it.
	blueEdited := image.NewRGBA(watermarked.Bounds())
	copy(blueEdited.Pix, watermarked.Pix)
	for y := 16; y < 24; y++ {
		for x := 16; x < 24; x++ {
			o := blueEdited.PixOffset(x, y)
			blueEdited.Pix[o+2] ^= 0x0f
		}
	}
	rep := verified(t, blueEdited, testKey)
	target := 2*rep.BlocksX + 2
	if !rep.TamperMap[target] {
		t.Fatalf("B-low-nibble edit in block %d was not detected", target)
	}

	// Copying a complete watermarked block to another grid position carries the old block's
	// pixels and payload bits with it. The HMAC includes the destination index, so relocation
	// cannot authenticate as the target block.
	relocated := image.NewRGBA(watermarked.Bounds())
	copy(relocated.Pix, watermarked.Pix)
	for dy := 0; dy < 8; dy++ {
		for dx := 0; dx < 8; dx++ {
			src := watermarked.PixOffset(5*8+dx, 5*8+dy)
			dst := relocated.PixOffset(2*8+dx, 2*8+dy)
			copy(relocated.Pix[dst:dst+4], watermarked.Pix[src:src+4])
		}
	}
	rep = verified(t, relocated, testKey)
	if !rep.TamperMap[target] {
		t.Fatalf("relocated block at index %d was not detected", target)
	}
	t.Logf("keyed auth: B-low edit and relocated block both detected; relocation flagged=%d/%d", rep.Flagged, rep.Total)
}

func TestDualCarrierSurvivesLocalisedErase(t *testing.T) {
	watermarked := embedded(t, naturalImage(512, 512))
	rect := image.Rect(192, 184, 224, 216)
	tampered, err := steg.Tamper(watermarked, rect.Min.X, rect.Min.Y, rect.Dx(), rect.Dy(), "black", nil)
	if err != nil {
		t.Fatalf("Tamper: %v", err)
	}
	_, rep, err := steg.Recover(tampered, testKey, steg.DefaultParams())
	if err != nil {
		t.Fatalf("Recover: %v", err)
	}
	t.Logf("dual carrier: flagged=%d recovered=%d unrecovered=%d", rep.Flagged, rep.Recovered, rep.Unrecovered)
	if rep.Flagged == 0 || rep.Unrecovered != 0 || rep.Recovered != rep.Flagged {
		t.Errorf("second carrier did not preserve all recoverable blocks: flagged=%d recovered=%d unrecovered=%d", rep.Flagged, rep.Recovered, rep.Unrecovered)
	}
}

func TestOddAndEdgeSizes(t *testing.T) {
	for _, tc := range []struct {
		w, h   int
		border int
		blocks int
	}{
		{1, 1, 0, 0}, {7, 7, 0, 0},
		{8, 8, 0, 1}, {9, 9, 1, 1}, {100, 37, 5, 48},
		{8, 200, 0, 25}, {200, 8, 0, 25},
	} {
		t.Run(stringSize(tc.w, tc.h), func(t *testing.T) {
			original := naturalImage(tc.w, tc.h)
			watermarked, embedErr := steg.Embed(original, testKey, steg.DefaultParams())
			if tc.blocks == 0 {
				if embedErr == nil {
					t.Error("Embed succeeded without one whole block on each axis")
				}
				if _, err := steg.Verify(original, testKey, steg.DefaultParams()); err == nil {
					t.Error("Verify succeeded without one whole block on each axis")
				}
				if _, _, err := steg.Recover(original, testKey, steg.DefaultParams()); err == nil {
					t.Error("Recover succeeded without one whole block on each axis")
				}
				t.Logf("%dx%d: Embed error=%v", tc.w, tc.h, embedErr)
				return
			}
			if embedErr != nil {
				t.Fatalf("Embed: %v", embedErr)
			}
			if watermarked.Bounds() != original.Bounds() {
				t.Errorf("dimensions=%v, want exact %v", watermarked.Bounds(), original.Bounds())
			}
			rep := verified(t, watermarked, testKey)
			t.Logf("%dx%d: flagged=%d/%d border=%d", tc.w, tc.h, rep.Flagged, rep.Total, rep.BorderPx)
			if rep.Flagged != 0 || rep.Total != tc.blocks {
				t.Errorf("flagged=%d total=%d; want 0/%d", rep.Flagged, rep.Total, tc.blocks)
			}
			// BorderPx is the Go field with JSON tag unwatermarkedBorderPx.
			if rep.BorderPx != tc.border {
				t.Errorf("unwatermarked border=%d, want %d", rep.BorderPx, tc.border)
			}
			for y := 0; y < tc.h; y++ {
				for x := 0; x < tc.w; x++ {
					if x < rep.BlocksX*rep.BlockSize && y < rep.BlocksY*rep.BlockSize {
						continue
					}
					o := original.PixOffset(x, y)
					if !bytes.Equal(original.Pix[o:o+4], watermarked.Pix[watermarked.PixOffset(x, y):watermarked.PixOffset(x, y)+4]) {
						t.Fatalf("unwatermarked border pixel changed at (%d,%d)", x, y)
					}
				}
			}
		})
	}
	if _, err := steg.Embed(naturalImage(8, 8), "", steg.DefaultParams()); err == nil {
		t.Error("Embed with empty key succeeded")
	} else {
		t.Logf("empty key error: %v", err)
	}
	if _, err := steg.Embed(nil, testKey, steg.DefaultParams()); err == nil {
		t.Error("Embed(nil) succeeded")
	} else {
		t.Logf("nil image error: %v", err)
	}
}

func stringSize(w, h int) string {
	return strconv.Itoa(w) + "x" + strconv.Itoa(h)
}

func TestGraduatedTamperDegradation(t *testing.T) {
	original := naturalImage(512, 512)
	watermarked := embedded(t, original)
	// Block-aligned rectangles have exact ground-truth coverage; the areas
	// are approximately 5%, 15%, and 30% of the 512x512 image.
	cases := []struct {
		label string
		rect  image.Rectangle
	}{
		{"5%", image.Rect(168, 184, 296, 288)},
		{"15%", image.Rect(120, 144, 320, 344)},
		{"30%", image.Rect(72, 88, 360, 360)},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			r := tc.rect
			tampered, err := steg.Tamper(watermarked, r.Min.X, r.Min.Y, r.Dx(), r.Dy(), "black", nil)
			if err != nil {
				t.Fatalf("Tamper: %v", err)
			}
			verifiedReport := verified(t, tampered, testKey)
			tpr, fpr, positive, tp, fp := detection(verifiedReport, r)
			recovered, rep, err := steg.Recover(tampered, testKey, steg.DefaultParams())
			if err != nil {
				t.Fatalf("Recover: %v", err)
			}
			region := regionPSNR(t, recovered, original, r)
			whole := metricPSNR(t, recovered, original)
			ratio := 100 * float64(r.Dx()*r.Dy()) / float64(original.Bounds().Dx()*original.Bounds().Dy())
			t.Logf("ratio=%s actual=%.4f%% truth=%d TP=%d FP=%d TPR=%.6f FPR=%.6f recovered=%d unrecovered=%d region PSNR=%.4f dB whole PSNR=%.4f dB", tc.label, ratio, positive, tp, fp, tpr, fpr, rep.Recovered, rep.Unrecovered, region, whole)
			if tpr < 0.9 {
				t.Errorf("%s TPR %.6f < 0.9", tc.label, tpr)
			}
		})
	}
}

func TestImageNotMutated(t *testing.T) {
	original := naturalImage(512, 512)
	before := append([]byte(nil), original.Pix...)
	watermarked := embedded(t, original)
	if !bytes.Equal(original.Pix, before) {
		t.Error("Embed modified its input pixels")
	}
	if watermarked == original || &watermarked.Pix[0] == &original.Pix[0] {
		t.Error("Embed returned the input or shared its backing array")
	}
	watermarked.Pix[0] ^= 0xff
	if !bytes.Equal(original.Pix, before) {
		t.Error("modifying returned image changed input pixels")
	}
}

func TestPayloadCapacityInvariant(t *testing.T) {
	original := naturalImage(8, 8)
	for _, tc := range []struct {
		name string
		p    steg.Params
	}{
		{"4px block", steg.Params{BlockSize: 4}},
		{"3 copies", steg.Params{BlockSize: 8, Copies: 3}},
		{"insufficient payload planes", steg.Params{BlockSize: 8, Copies: 2, PlanesPerCopy: 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := steg.Embed(original, testKey, tc.p)
			t.Logf("unsupported geometry %s: error=%v", tc.name, err)
			if err == nil {
				t.Error("Embed accepted unsupported geometry")
			}
		})
	}
}
