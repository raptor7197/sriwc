package steg

import "image"

// grid gives block-addressed access to an image. Only whole blocks are watermarked: the
// grid covers floor(w/blockSize) x floor(h/blockSize) blocks, and any leftover border strip
// (up to blockSize-1 px wide) carries no payload and is never flagged. That keeps the image
// dimensions exactly as uploaded instead of padding and cropping, which would move payload
// bits into pixels that disappear.
type grid struct {
	img    *image.RGBA
	bs     int
	bx, by int
}

func newGrid(img *image.RGBA, bs int) *grid {
	b := img.Bounds()
	return &grid{img: img, bs: bs, bx: b.Dx() / bs, by: b.Dy() / bs}
}

func (g *grid) count() int { return g.bx * g.by }

func (g *grid) origin(i int) (int, int) {
	return (i % g.bx) * g.bs, (i / g.bx) * g.bs
}

// offset is the byte offset of the k-th pixel (raster order) within block i.
func (g *grid) offset(i, k int) int {
	x0, y0 := g.origin(i)
	b := g.img.Bounds()
	return g.img.PixOffset(b.Min.X+x0+k%g.bs, b.Min.Y+y0+k/g.bs)
}

func (g *grid) pix(i, k, ch int) uint8 { return g.img.Pix[g.offset(i, k)+ch] }

func (g *grid) setPix(i, k, ch int, v uint8) { g.img.Pix[g.offset(i, k)+ch] = v }

// copyImage returns a deep copy of img.
func copyImage(img *image.RGBA) *image.RGBA {
	b := img.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := 0; y < b.Dy(); y++ {
		copy(out.Pix[y*out.Stride:y*out.Stride+b.Dx()*4], img.Pix[y*img.Stride:y*img.Stride+b.Dx()*4])
	}
	return out
}
