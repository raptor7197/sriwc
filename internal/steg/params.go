// Package steg implements a self-recoverable fragile image watermark: recovery data for
// every block is embedded into keyed carrier blocks, so tampered regions can be detected
// and reconstructed from the surviving embedded copies.
//
// Scheme (see .hermes_spec.md and .hermes/plans/):
//
//	block 8x8 px; per-block payload = 248 bits
//	  240 recovery bits = 16 quadrants x 3 channels x 5 bits (quadrant means)
//	    8 auth bits     = truncated, position-bound HMAC-SHA256 tag
//	2 redundant copies, carrier of copy0 = shift0[i] (embedded in R planes 0..3),
//	                        carrier of copy1 = shift1[i] (embedded in G planes 0..3)
//
// The pipeline is lossless (PNG) by design; JPEG recompression destroys LSB watermarks.
package steg

import "fmt"

// Params controls the embedding geometry.
type Params struct {
	BlockSize     int `json:"blockSize"`
	Copies        int `json:"copies"`
	PlanesPerCopy int `json:"planesPerCopy"`
}

// DefaultParams returns the supported geometry.
func DefaultParams() Params {
	return Params{BlockSize: blockSize, Copies: copies, PlanesPerCopy: planesPerCopy}
}

func (p Params) normalized() Params {
	d := DefaultParams()
	if p.BlockSize == 0 {
		p.BlockSize = d.BlockSize
	}
	if p.Copies == 0 {
		p.Copies = d.Copies
	}
	if p.PlanesPerCopy == 0 {
		p.PlanesPerCopy = d.PlanesPerCopy
	}
	return p
}

func (p Params) validate() (Params, error) {
	p = p.normalized()
	if p.BlockSize != blockSize {
		return p, fmt.Errorf("unsupported block size %d (only %d is supported)", p.BlockSize, blockSize)
	}
	if p.Copies < 1 || p.Copies > maxCopies {
		return p, fmt.Errorf("copies must be between 1 and %d, got %d", maxCopies, p.Copies)
	}
	if p.PlanesPerCopy < 1 || p.PlanesPerCopy > 8 {
		return p, fmt.Errorf("planesPerCopy must be between 1 and 8, got %d", p.PlanesPerCopy)
	}
	if p.blockCapacity() < payloadBits {
		return p, fmt.Errorf("carrier capacity %d bits < payload %d bits", p.blockCapacity(), payloadBits)
	}
	return p, nil
}

// blockCapacity is the number of payload bits one carrier block can hold per copy.
func (p Params) blockCapacity() int { return p.BlockSize * p.BlockSize * p.PlanesPerCopy }
