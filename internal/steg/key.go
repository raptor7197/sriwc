package steg

import (
	"hash/fnv"
	"strconv"
)

// Geometry and payload constants. Frozen: see .hermes_spec.md.
const (
	blockSize     = 8 // host/carrier block edge in pixels
	copies        = 2 // redundant carriers per block
	planesPerCopy = 4 // LSB planes used in the copy's channel
	maxCopies     = 2 // copy0 -> channel R, copy1 -> channel G

	quadrantsPerChannel = 16                                                       // (blockSize/2)^2
	recoveryChannels    = 3                                                        // R, G, B
	bitsPerQuadMean     = 5                                                        // quadrant mean quantised to 0..31
	recoveryBits        = quadrantsPerChannel * recoveryChannels * bitsPerQuadMean // 240
	authBits            = 8
	payloadBits         = recoveryBits + authBits // 248
)

// copyChannel maps copy index -> the colour channel its payload lives in (R, G).
var copyChannel = [maxCopies]int{0, 1}

// keyHash is a deterministic 64-bit hash of key + ":" + tag.
func keyHash(key, tag string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	_, _ = h.Write([]byte(":"))
	_, _ = h.Write([]byte(tag))
	return h.Sum64()
}

// carrierPerms returns, for every copy, the block that carries block i's payload
// (perms[c][i]).
//
// The mapping is deliberately NOT a uniform random permutation. It is a large key-derived
// cyclic shift of the block grid: copy 0 moves a block's data forward by ~1/4 of the grid,
// copy 1 by ~3/4. With a uniform permutation, a contiguous tamper covering only 10% of the
// image lands BOTH of a block's carriers inside the tampered area for ~1% of all blocks,
// which manufactures false positives and leaves those blocks unrecoverable. With the shift
// mapping a contiguous region smaller than ~1/4 of the grid can contain at most one
// carrier of any block, so detection stays exact and every block inside the region is
// restorable from its other copy. Trade-off: the mapping is easier to guess than a full
// permutation for an attacker who knows the scheme (see README "Known limits").
func carrierPerms(key string, n int, p Params) [][]int {
	perms := make([][]int, p.Copies)
	used := make(map[int]bool, p.Copies)
	for c := 0; c < p.Copies; c++ {
		s := carrierShift(key, n, c, p.Copies)
		for used[s] && n > 1 { // two copies must never share one carrier
			s = (s + 1) % n
		}
		used[s] = true
		pm := make([]int, n)
		for i := range pm {
			pm[i] = (i + s) % n
		}
		perms[c] = pm
	}
	return perms
}

// carrierShift derives the per-copy displacement from the key.
func carrierShift(key string, n, copy, copies int) int {
	if n <= 1 {
		return 0
	}
	if copies <= 1 {
		copies = 1
	}
	num := int64(copy*2 + 1) // 1, 3, 5 ... over
	den := int64(2 * copies) // 2, 4, 6 ...
	base := int(num * int64(n) / den)
	jitter := 0
	if n >= 16 {
		if span := uint64(n / 8); span > 0 {
			jitter = int(keyHash(key, "p"+strconv.Itoa(copy)) % span)
		}
	}
	s := (base + jitter) % n
	if s == 0 {
		s = 1
	}
	return s
}
