package steg

import (
	"fmt"
	"image"
)

// Report describes one verification or recovery run.
type Report struct {
	BlocksX      int     `json:"blocksX"`
	BlocksY      int     `json:"blocksY"`
	BlockSize    int     `json:"blockSize"`
	ImgWidth     int     `json:"imgWidth"`
	ImgHeight    int     `json:"imgHeight"`
	BorderPx     int     `json:"unwatermarkedBorderPx"`
	Total        int     `json:"totalBlocks"`
	Flagged      int     `json:"flaggedBlocks"`
	TamperMap    []bool  `json:"-"`
	TamperGrid   [][]int `json:"tamperMap"`
	CarrierLost  int     `json:"carrierLostBlocks"`
	Recovered    int     `json:"recoveredBlocks"`
	Unrecovered  int     `json:"unrecoveredBlocks"`
	FlaggedIdx   []int   `json:"-"`
	RecoveredIdx []int   `json:"-"`
}

func newReport(g *grid, p Params) *Report {
	b := g.img.Bounds()
	border := 0
	if g.bs > 0 {
		if r := b.Dx() - g.bx*g.bs; r > border {
			border = r
		}
		if r := b.Dy() - g.by*g.bs; r > border {
			border = r
		}
	}
	return &Report{
		BlocksX:   g.bx,
		BlocksY:   g.by,
		BlockSize: p.BlockSize,
		ImgWidth:  b.Dx(),
		ImgHeight: b.Dy(),
		BorderPx:  border,
		Total:     g.count(),
		TamperMap: make([]bool, g.count()),
	}
}

func (r *Report) buildGrid() {
	r.TamperGrid = make([][]int, r.BlocksY)
	for y := 0; y < r.BlocksY; y++ {
		row := make([]int, r.BlocksX)
		for x := 0; x < r.BlocksX; x++ {
			if r.TamperMap[y*r.BlocksX+x] {
				row[x] = 1
			}
		}
		r.TamperGrid[y] = row
	}
}

// analyze classifies every block of g. A block is flagged as tampered when its recomputed
// authentication code matches NEITHER embedded copy, so a single destroyed carrier does not
// create a false positive.
func analyze(g *grid, key string, p Params) *Report {
	n := g.count()
	perms := carrierPerms(key, n, p)
	rep := newReport(g, p)
	for i := 0; i < n; i++ {
		crc := blockAuth(g, i, key)
		matched := false
		for c := 0; c < p.Copies; c++ {
			bits := readPayload(g, perms[c][i], copyChannel[c], p)
			if authFromBits(bits) == crc {
				matched = true
				break
			}
		}
		if !matched {
			rep.TamperMap[i] = true
			rep.FlaggedIdx = append(rep.FlaggedIdx, i)
		}
	}
	rep.Flagged = len(rep.FlaggedIdx)
	rep.buildGrid()
	return rep
}

// Verify reports which blocks of w are inconsistent with their embedded authentication
// data. w must be the watermarked image (possibly tampered); a lossless pipeline is assumed.
func Verify(w *image.RGBA, key string, p Params) (*Report, error) {
	if w == nil {
		return nil, fmt.Errorf("nil image")
	}
	if key == "" {
		return nil, fmt.Errorf("key must not be empty")
	}
	p, err := p.validate()
	if err != nil {
		return nil, err
	}
	g := newGrid(w, p.BlockSize)
	if g.count() == 0 {
		return nil, fmt.Errorf("image smaller than one %dx%d block", p.BlockSize, p.BlockSize)
	}
	return analyze(g, key, p), nil
}

// Recover rebuilds the flagged blocks of w from the embedded recovery data. The returned
// image has the same dimensions as w. Payloads are read for every flagged block before any
// reconstruction is written, so restoring one block can never corrupt the data another
// block still needs.
func Recover(w *image.RGBA, key string, p Params) (*image.RGBA, *Report, error) {
	if w == nil {
		return nil, nil, fmt.Errorf("nil image")
	}
	if key == "" {
		return nil, nil, fmt.Errorf("key must not be empty")
	}
	p, err := p.validate()
	if err != nil {
		return nil, nil, err
	}
	out := copyImage(w)
	g := newGrid(out, p.BlockSize)
	n := g.count()
	if n == 0 {
		return nil, nil, fmt.Errorf("image smaller than one %dx%d block", p.BlockSize, p.BlockSize)
	}
	rep := analyze(g, key, p)
	perms := carrierPerms(key, n, p)

	type fix struct {
		block int
		bits  []byte
	}
	fixes := make([]fix, 0, len(rep.FlaggedIdx))
	for _, i := range rep.FlaggedIdx {
		restored := false
		var firstLost bool
		for c := 0; c < p.Copies; c++ {
			carrier := perms[c][i]
			if rep.TamperMap[carrier] { // that carrier's payload is gone
				if c == 0 {
					firstLost = true
				}
				continue
			}
			fixes = append(fixes, fix{block: i, bits: readPayload(g, carrier, copyChannel[c], p)})
			restored = true
			break
		}
		if restored {
			rep.Recovered++
			rep.RecoveredIdx = append(rep.RecoveredIdx, i)
		} else {
			rep.Unrecovered++
			if firstLost {
				rep.CarrierLost++
			}
		}
	}
	for _, f := range fixes {
		decodeRecoveryInto(g, f.block, f.bits)
	}
	return out, rep, nil
}
