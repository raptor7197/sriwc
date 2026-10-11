package steg

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
)

// crc8 is the authentication helper from the rejected revision-A draft. It is retained as
// an inert reference for the design history; production authentication uses keyed HMAC-SHA256
// in blockAuth below.
func crc8(data []byte) byte {
	var c byte = 0xFF
	for _, v := range data {
		c ^= v
		for i := 0; i < 8; i++ {
			if c&0x80 != 0 {
				c = c<<1 ^ 0x07
			} else {
				c <<= 1
			}
		}
	}
	return c
}

// blockAuth returns a keyed 8-bit authentication tag for block i. HMAC prevents an attacker
// who does not know the key from rewriting the public authentication bits after forging the
// image. The block index is included, so moving an intact block to another location does not
// preserve its tag. Eight bits is a deliberate payload constraint, not cryptographic-strength
// authentication: a random forged block still has a 1/256 chance of matching one copy.
//
// R and G contribute their high nibbles because their low nibbles carry the watermark. B is
// not used as a carrier, so all eight B bits are authenticated; this also detects low-nibble
// edits to B that the earlier CRC-only draft missed.
func blockAuth(g *grid, i int, key string) byte {
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write([]byte("stegno-auth-v1"))
	var pos [8]byte
	binary.LittleEndian.PutUint64(pos[:], uint64(i))
	_, _ = mac.Write(pos[:])
	buf := make([]byte, 0, g.bs*g.bs*3)
	for k := 0; k < g.bs*g.bs; k++ {
		off := g.offset(i, k)
		buf = append(buf, g.img.Pix[off]>>4, g.img.Pix[off+1]>>4, g.img.Pix[off+2])
	}
	_, _ = mac.Write(buf)
	return mac.Sum(nil)[0]
}

// encodePayload builds the payload of block i as a bit slice (MSB first): 240 recovery bits
// (16 quadrants x 3 channels x 5 bits, channel-major) followed by 8 authentication bits.
func encodePayload(g *grid, i int, p Params, key string) []byte {
	bits := make([]byte, 0, payloadBits)
	q := g.bs / 2
	for ch := 0; ch < recoveryChannels; ch++ {
		for qy := 0; qy < q; qy++ {
			for qx := 0; qx < q; qx++ {
				sum := 0
				for dy := 0; dy < 2; dy++ {
					for dx := 0; dx < 2; dx++ {
						sum += int(g.pix(i, (qy*2+dy)*g.bs+qx*2+dx, ch))
					}
				}
				mv := (sum / 4) >> 3 // rounded-down mean, quantised to 5 bits
				if mv > 31 {
					mv = 31
				}
				for b := bitsPerQuadMean - 1; b >= 0; b-- {
					bits = append(bits, byte((mv>>uint(b))&1))
				}
			}
		}
	}
	a := blockAuth(g, i, key)
	for b := authBits - 1; b >= 0; b-- {
		bits = append(bits, (a>>uint(b))&1)
	}
	return bits
}

// authFromBits reads the trailing 8 authentication bits.
func authFromBits(bits []byte) byte {
	var a byte
	for _, b := range bits[recoveryBits:payloadBits] {
		a = a<<1 | b
	}
	return a
}

// writePayload stores bits in the low PlanesPerCopy bits of channel ch of carrier block ci.
func writePayload(g *grid, ci, ch int, bits []byte, p Params) {
	pp := p.PlanesPerCopy
	for b, bit := range bits {
		off := g.offset(ci, b/pp) + ch
		mask := uint8(1) << uint(b%pp)
		if bit != 0 {
			g.img.Pix[off] |= mask
		} else {
			g.img.Pix[off] &^= mask
		}
	}
}

// readPayload is the inverse of writePayload.
func readPayload(g *grid, ci, ch int, p Params) []byte {
	pp := p.PlanesPerCopy
	bits := make([]byte, payloadBits)
	for b := 0; b < payloadBits; b++ {
		v := g.img.Pix[g.offset(ci, b/pp)+ch]
		bits[b] = (v >> uint(b%pp)) & 1
	}
	return bits
}

// decodeRecoveryInto rebuilds block dst from the 240 recovery bits.
func decodeRecoveryInto(g *grid, dst int, bits []byte) {
	q := g.bs / 2
	idx := 0
	for ch := 0; ch < recoveryChannels; ch++ {
		for qy := 0; qy < q; qy++ {
			for qx := 0; qx < q; qx++ {
				mv := 0
				for b := 0; b < bitsPerQuadMean; b++ {
					mv = mv<<1 | int(bits[idx])
					idx++
				}
				v := uint8(mv*8 + 3) // bucket centre, max 31*8+3 = 251
				for dy := 0; dy < 2; dy++ {
					for dx := 0; dx < 2; dx++ {
						g.setPix(dst, (qy*2+dy)*g.bs+qx*2+dx, ch, v)
					}
				}
			}
		}
	}
}
