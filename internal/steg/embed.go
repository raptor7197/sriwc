package steg

import (
	"fmt"
	"image"
)

// Embed returns a watermarked copy of src. src is not modified and the result has exactly
// the same dimensions. Only whole 8x8 blocks are watermarked; a leftover border strip of up
// to 7 px (when the dimensions are not multiples of the block size) keeps its original
// pixels and is outside the verifiable region.
func Embed(src *image.RGBA, key string, p Params) (*image.RGBA, error) {
	if src == nil {
		return nil, fmt.Errorf("nil image")
	}
	if key == "" {
		return nil, fmt.Errorf("key must not be empty")
	}
	p, err := p.validate()
	if err != nil {
		return nil, err
	}
	g := newGrid(src, p.BlockSize)
	n := g.count()
	if n == 0 {
		return nil, fmt.Errorf("image smaller than one %dx%d block: %dx%d", p.BlockSize, p.BlockSize, src.Bounds().Dx(), src.Bounds().Dy())
	}

	out := copyImage(src)
	og := newGrid(out, p.BlockSize)
	perms := carrierPerms(key, n, p)

	// Payloads come from the unmodified content (they read only the high nibbles, which
	// embedding never touches), so the carrier LSBs written below cannot influence them.
	payloads := make([][]byte, n)
	for i := 0; i < n; i++ {
		payloads[i] = encodePayload(og, i, p, key)
	}
	for i := 0; i < n; i++ {
		for c := 0; c < p.Copies; c++ {
			writePayload(og, perms[c][i], copyChannel[c], payloads[i], p)
		}
	}
	return out, nil
}
