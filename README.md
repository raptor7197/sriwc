# stegno — Self-Recoverable Image Watermarking for Content Restoration

Detect tampering in an image and **reconstruct the damaged regions** from recovery data that
was embedded in the image itself, before it was attacked.

- Backend: Go 1.27, standard library only (`image/png`, `net/http`). No external Go modules.
- Frontend: Vite + React 18 + Tailwind CSS.
- Lossless pipeline: PNG in, PNG out. JPEG recompression destroys the LSB payload by design.

## How it works

Every 8x8 pixel block carries its own compressed description inside *other* blocks:

| Piece | Detail |
|---|---|
| Recovery data | the block reduced to 16 quadrants (2x2 px each) x 3 channels, each quadrant mean quantised to 5 bits = 240 bits |
| Authentication | first 8 bits of HMAC-SHA256 over the key, block index and protected samples = 8 bits |
| Payload | 248 bits per block |
| Carriers | two key-derived cyclic offsets near 1/4 and 3/4 of the block grid; copy 0 is written into the low 4 bits of **R**, copy 1 into the low 4 bits of **G** |
| Verification | recompute the keyed tag of each block; a block is flagged when its tag matches **neither** embedded copy (dual-copy voting keeps a single destroyed carrier from creating a false positive) |
| Recovery | rebuild each flagged block from the first carrier that is *not itself flagged*; blocks whose two carriers are both destroyed are reported as `unrecovered` instead of being silently guessed |

Measured distortion is about 34 dB PSNR (the price of 4 LSB planes in two channels);
reconstruction is about 23-40 dB PSNR depending on texture.

Lineage: Fridrich & Goljan (1999) self-correcting images; Lin, Hsieh, Huang, Chang (Pattern
Recognition 2005) hierarchical block-mapping tamper detection and recovery; Chan & Chang
(2007) keyed-permutation self-recovery; Zhang, Wang, Qian, Feng (IEEE TIFS 2011) reference
sharing / multiple-carrier self-embedding.

## Run

```bash
# 1. backend + API on :8080 (serves the built frontend when web/dist exists)
cd /home/krxsna/dev/stegno-project
bash scripts/run.sh              # builds web/dist if missing, then serves on :8080

# 2. frontend dev server with hot reload on :5173 (proxies /api -> :8080)
cd web && npm install && npm run dev
```

Container paths (no system Go/node): Go lives in `~/.hermes/tools/go-1.27.2-linux-x64/bin`,
node in `~/.hermes/tools/node-26.7.0-linux-x64/bin`. `scripts/run.sh` sets both.

## CLI

```bash
go run ./cmd/stegno analyze -in testdata/sample.png -x 120 -y 80 -w 220 -h 140 -mode black
go run ./cmd/stegno embed  -in testdata/sample.png -out /tmp/w.png -key mykey
go run ./cmd/stegno verify -in /tmp/w.png -key mykey
go run ./cmd/stegno recover -in /tmp/tampered.png -out /tmp/rec.png -key mykey
```

## API

`GET /api/health`, `POST /api/embed|tamper|verify|recover|analyze` — see `.hermes_spec.md`
for the exact request/response contract. `/api/analyze` runs the whole pipeline in one call
and returns every stage image plus the metrics (imperceptibility PSNR/SSIM, detection
TPR/FPR, recovery PSNR/SSIM).

## Known limits (honest list)

- PNG only. Any lossy recompression, resizing, rotation or colour shift destroys the payload.
- An attacker who knows the key can forge a valid watermark; change detection uses a keyed
  truncated HMAC, not a public-key signature.
- A tampered block whose **both** carriers were also tampered cannot be reconstructed; the
  report counts these in `unrecoveredBlocks`.
- Blocks modified only inside their low 4 bits (an extremely subtle edit) may leave their
  carrier looking intact while its payload is already garbage.
- Authentication is 8 bits, so a random forged block has roughly a 1/256 chance per tested
  carrier of escaping detection. Use a high-entropy key; weak keys can be dictionary-tested.
- A right/bottom remainder smaller than one 8x8 block is kept unchanged and unauthenticated;
  `unwatermarkedBorderPx` reports its maximum width.
