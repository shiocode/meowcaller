# meowcaller — execution plan

A **clean-room, pure-Go** WhatsApp 1:1 calling library. No WASM bridge, no cgo for
the protocol, no inherited code from earlier attempts. It is a faithful port of
the validated Rust reference, built **module-by-module and verified byte-exact
against shared known-answer vectors (KATs)**.

> This is a plan to execute, not finished code. Every step below has a concrete
> "done when" tied to a test vector, so progress is unambiguous and a wrong port
> cannot masquerade as a working one.

---

## 1. Goal and non-goals

**Goal:** a Go module that can place and receive a real WhatsApp 1:1 **audio**
call end to end — signaling, keying, transport, and the MLow codec — matching the
reference bit-for-bit where vectors exist.

**Non-goals (for v1):** video, group calls, the neural "companion" post-filter,
and any reliance on the WhatsApp WASM at runtime.

**Hard rule:** do not copy from the old `dublin` / `meowmeow` calling code. It is
the unvalidated prior attempt. Start from the Rust reference and the KATs only.

---

## 2. The single source of truth

The Rust reference is **`whatsapp-rust`** (path on this machine:
`/Users/purpshell/Documents/Programming/whatsapp-rust-voip`). It is KAT-pinned and
known to work. The relevant trees:

- `wacore/src/voip/` — the platform-agnostic core to port:
  - `mlow/` — the MLow/SMPL CELP codec (~31 modules) + `testdata/*.json` KATs.
  - `e2e_srtp.rs`, `hbh_srtp.rs`, `sframe.rs`, `warp.rs` — crypto.
  - `stun.rs`, `rtp.rs`, `rtcp.rs`, `relay_parse.rs`, `ssrc.rs` — transport.
  - `stanza.rs` — call signaling builders.
- `src/voip/` — the orchestration (session, transport loop, audio).
- `examples/voip.rs` — the end-to-end driver (loopback / listen / call).

Lineage to keep in mind: **WhatsApp WASM → byte-exact Go reference → zapo-caller
(TS) → whatsapp-rust (Rust)**. The Rust is the most complete and the one with
checked-in vectors, so it is the porting target.

**Protocol documentation** (the "why" behind the code) lives in the wacrg repo
(`/Users/purpshell/Documents/Programming/wacrg/docs/`): `codec/mlow/`,
`keying/srtp-key-schedule.md`, `keying/sframe-media-e2ee.md`,
`signaling/stanza-reference.md`, `transport/warp-stun-relay.md`. Read the relevant
page before porting a subsystem.

---

## 3. Methodology — KAT-driven port

For **every** module:

1. Read the Rust module and its test (`#[cfg(test)]`), and the wacrg doc.
2. Copy the Rust's vector JSON into `meowcaller/<pkg>/testdata/` verbatim.
3. Port the logic to idiomatic Go (mirror the Rust's structure and names).
4. Write a Go test that loads the same vector and asserts **byte/float-exact**.
5. The module is "done" only when its vector passes. No vector → not done.

This is non-negotiable for the DSP: a wrong-but-self-consistent codec passes a
round-trip but produces garbage against real frames. The vectors are the guard.

The Rust ships these KATs in `wacore/src/voip/mlow/testdata/`:
`rc_vectors.json`, `toc_vectors.json`, `smpl_tables.json` (the heap ROM),
`lsf_vectors.json`, `lsf_quant_io.json`, `lsf_cb_dump.json`, `pitch_vectors.json`,
`c_pitch_full.json`, `pulse_vectors.json`, `gains_vectors.json`,
`sigmode_ground_truth.json`, `vad_ground_truth.json`, `gennoise_vectors.json`,
`e2e_vectors.json`, and `inbound_capture_frames.json` (real captured frames).
The crypto/signaling modules have inline `#[test]` KATs in their `.rs` files.

---

## 4. Project layout

Mirror the Rust so a reviewer can diff module-for-module:

```
meowcaller/
  go.mod                      module github.com/<you>/meowcaller
  PLAN.md                     this file
  internal/kat/kat.go         shared JSON-vector loader for tests
  mlow/                       the codec
    rangecoder.go  toc.go  mem.go  lpc.go  lsf.go  lsf_quant.go
    pitch.go  pulse.go  gains.go  synth.go  postfilter.go  vad.go
    noise.go  red.go  decoder.go  encoder.go  tables.go
    testdata/*.json           copied from the Rust
  crypto/                     keying
    hkdf.go  e2e_srtp.go  hbh_srtp.go  sframe.go  warp.go
  signal/                     call signaling
    stanza.go                 the <call> builders
  transport/                  media plane
    stun.go  rtp.go  rtcp.go  ssrc.go  relay.go
  session/                    orchestration
    session.go  pipeline.go
  cmd/meowcall/main.go        loopback / listen / call driver
```

Pure-Go deps only. Transport may use `pion/dtls`, `pion/sctp` (pure Go). The
codec must be dependency-free.

---

## 5. Port order

Dependency-ordered. Each line: **module — Rust source — KAT — effort**. Do them in
order; later modules need earlier ones.

### Phase A — codec foundation (small, unblock everything)

1. **rangecoder** — `mlow/rangecoder.rs` — `rc_vectors.json` — small.
   The CELT/Opus range coder (`ec_dec`/`ec_enc`). A verified Go implementation
   already exists in wacrg `impl/mlow/rangecoder.go` (decoder + matched encoder,
   round-trip tested) — adapt it and add the `rc_vectors.json` check.
2. **toc** — `mlow/toc.rs` — `toc_vectors.json` (256 cases) — small.
   The smpl TOC bit layout. *(A correct, KAT-passing Go port already exists — see
   Appendix A; drop it in as the first module.)*
3. **mem (heap ROM)** — `mlow/smpl_mem.rs` + `smpl_tables.json` — medium.
   The ~49 KB constant ROM (LTP codebooks, pulse/gain CDFs, pitch contours) +
   typed accessors (`i16@addr`, `cdf@addr`). **Load-bearing** for every decode
   step. Verify table reads at known addresses against the Rust.

### Phase B — receive DSP (the audible path)

4. **lpc** — `mlow/smpl_lpc.rs` — `lsf_vectors.json` — large.
   FFT-based LPC analysis (512-pt RFFT → power spectrum → cosine transform →
   reflection coeffs → LPC + bandwidth expansion). Needs a portable mixed-radix
   FFT. Replaces any naive time-domain autocorrelation.
5. **lsf_quant** — `mlow/smpl_lsf_quant.rs` + `lsf_cb_dump.json` —
   `lsf_quant_io.json` — large. 2-stage LSF VQ (Mahalanobis shortlist + RD beam,
   per-coeff stage-2 with constraint tables).
6. **lsf decode** — `mlow/smpl_decode.rs` — covered by `lsf_*` — medium.
   Stage-1 grid selector + LSF→LPC reconstruction.
7. **pitch** — `mlow/smpl_pitch.rs` — `pitch_vectors.json`, `c_pitch_full.json` —
   large. Contour map, per-segment fractional lags (Q6), LTP filter index.
8. **pulse** — `mlow/smpl_pulse.rs` — `pulse_vectors.json` — large.
   PVQ-style algebraic pulse coding (triangular CDF prior, recursive split,
   run-length magnitudes, signs).
9. **gains** — `mlow/smpl_gains.rs` — `gains_vectors.json` — medium.
10. **synth** — `mlow/smpl_synth.rs` — checked via end-to-end — medium.
    CELP synthesis filter + LTP.
11. **postfilter / vad / noise / perc** — `smpl_postfilter.rs`,
    `smpl_harm_postfilter.rs`, `smpl_harmcomb.rs`, `smpl_vad.rs`,
    `smpl_gennoise.rs`, `smpl_perc.rs`, `smpl_nrgres.rs` —
    `vad_ground_truth.json`, `gennoise_vectors.json` — medium each.
12. **red** — `mlow/red.rs` — (RFC 2198 split) — small.
13. **decoder (integration)** — `mlow/decoder.rs` — **`e2e_vectors.json` +
    `inbound_capture_frames.json`** — integration. **This is the milestone: real
    captured frames → correct PCM.**

### Phase C — send DSP

14. **encoder + analysis** — `mlow/encode.rs`, `mlow/analysis.rs`,
    `smpl_pitch_enc.rs`, `smpl_signal_mode.rs` — `sigmode_ground_truth.json` +
    encoder vectors — large. (The old dublin encoder never emitted voiced frames;
    do not reuse it.)

### Phase D — crypto (small, KAT-verified by the Rust `#[test]`s)

15. **hkdf** — a thin HKDF-SHA256 helper (stdlib `golang.org/x/crypto/hkdf`).
16. **e2e_srtp** — `e2e_srtp.rs` — `HKDF-SHA256(callKey, info=LID, 46)` →
    RFC-3711 AES-CM labels 0x00/0x01/0x02; RTP IV = salt right-aligned, SSRC^@4-7,
    index^@8-13. **Confirmed correct in wacrg.**
17. **hbh_srtp** — `hbh_srtp.rs` — two-stage `wa_sfu_kdf` (uplink/downlink labels).
    Port the Rust's exact label set; do not assume the dublin "10-type" model.
18. **sframe** — `sframe.rs` — **AES-128-GCM**, 16-byte LE-counter nonce,
    `HKDF(salt=callKey[0:16], ikm=callKey[16:32], info="e2e sframe key"+pid, 32)`,
    wire `[ct||tag||varint-header]` (header not AAD). Verify against its `#[test]`.
19. **warp** — `warp.rs` — MI tag `HMAC-SHA1(warp_auth_key, pkt||roc_be32)[:4]`;
    `warp_auth_key = HKDF("", callKey, "warp auth key", 32)`.

### Phase E — signaling

20. **stanza** — `stanza.rs` — builders for offer/preaccept/accept/transport/
    relaylatency/heartbeat/terminate/mute_v2/reject. **Port the exact child order**
    (offer: privacy→audio8k→audio16k→net→capability→enc|destination→encopt→
    device-identity; server returns 439 if wrong), the capability blobs
    (`01 05 f7 09 e4 bb 13` offer/accept, `…07` preaccept), `keygen=2`, and
    `encode_latency = 0x02000000 + rtt_ms`. Mirror the Rust's unit tests.
    *Open discrepancy to settle with a capture: accept `<net medium>` — the Rust
    uses `2`; a prior capture-based attempt used `3`. Use the Rust's `2` and flag.*

### Phase F — transport + orchestration

21. **stun** — `stun.rs` — relay dialect (magic `0x2112a442`, ping `0x0801`/pong
    `0x0802`/allocate `0x0003`, FINGERPRINT `CRC32 ^ 0x5354554e`, MESSAGE-INTEGRITY
    by relay key; protobuf `0x40xx` allocate attrs).
22. **rtp / rtcp** — `rtp.rs`, `rtcp.rs` — WARP profile (first byte `0x90`, PT
    120/121, ext `0xdebe`, 16/20-byte headers).
23. **ssrc** — `ssrc.rs` — `HKDF(salt=slot_word_le32, ikm=call_id, info=LID, 4)`
    as LE u32 (derived, not random).
24. **relay** — `src/voip/transport.rs` — the DTLS/SCTP relay media loop (pion).
    This is the part the Rust itself defers; it is the last mile to a live call.
25. **session/pipeline** — wire signaling → keying → transport → codec → PCM.

---

## 6. Milestones (definition of "working")

- **M1 — Decode a real frame.** Phases A+B done: `inbound_capture_frames.json`
  decodes to PCM matching `e2e_vectors.json`. Proves the receive codec.
- **M2 — Full receive path.** + crypto (D) + transport (F up to relay): take a
  live inbound media packet, unwrap SRTP/SFrame/WARP, decode to audio.
- **M3 — Send path.** + encoder (C): mic PCM → MLow → SRTP/SFrame/WARP → relay.
- **M4 — End-to-end call.** + signaling (E) + orchestration: place/answer a real
  1:1 call with two-way audio (the Rust `examples/voip.rs` is the behavioral
  reference for the flow).

---

## 7. Verification harness (build this first)

`internal/kat/kat.go`: a tiny helper to load a `testdata/*.json` vector into a Go
struct, used by every module test. Keeps the KAT pattern one-line per test:

```go
var vs []TocVector
kat.Load(t, "testdata/toc_vectors.json", &vs)
for _, v := range vs { /* assert ParseTOC(v.B) == v */ }
```

Run `go test ./...` after every module. Green is the only acceptable state.

---

## 8. Pitfalls (learned the hard way)

- **Trust the vectors, not prose or function names.** Reverse-engineered names
  (in the WASM and in any analysis notes) are frequently wrong; a passing KAT is
  the only proof.
- **MLow is split-band CELP, not MDCT.** Any MDCT/transform code seen in the WASM
  is the *standard-Opus fallback* path, selected when the TOC top bits are `11`.
  The MLow path is LPC + pitch/LTP + algebraic pulse + synthesis.
- **SFrame is AES-GCM, not AES-CTR.** (AES-CTR is the SRTP cipher.)
- **Don't reuse the old codec.** It passes a sine round-trip and fails on real
  frames (its LPC front-end and pulse coding are wrong). Port fresh from the Rust.

---

## Appendix A — worked example (the pattern, already proven)

The smpl TOC module was ported and passes `toc_vectors.json` for all 256 byte
values. Use it as the template for every other module (port + copy vector + KAT
test). `mlow/toc.go`:

```go
package mlow

// SmplTOC is the decoded MLow "smpl_toc" (first byte of a bare MLow frame).
// (b & 0xC0) == 0xC0 marks a STANDARD Opus/CELT frame, routed to stock libopus.
// Bit layout: bit7=SID, bit6=VAD, bit5=rate(0->16k,1->32k), bits4:3=frame_ms
// index into {10,20,60,120}, bit2=flag2, bit1=voiced-enable, bit0=flag0.
type SmplTOC struct {
	StdOpus    bool
	SID        bool
	VAD        bool
	SampleRate int
	FrameMs    int
	Voiced     bool
	Active     bool
	Flag2      bool
	Flag0      bool
}

func standardOpusFrameMs(b byte) int {
	c := b >> 3
	switch {
	case c < 12:
		return []int{10, 20, 40, 60}[c&3]
	case c < 16:
		return []int{10, 20}[(c-12)&1]
	default:
		switch c & 3 {
		case 0:
			return 3 // 2.5ms rounded up
		case 1:
			return 5
		case 2:
			return 10
		default:
			return 20
		}
	}
}

func ParseSmplTOC(b byte) SmplTOC {
	if b&0xC0 == 0xC0 {
		return SmplTOC{StdOpus: true, SampleRate: 16000, FrameMs: standardOpusFrameMs(b)}
	}
	bit1 := (b>>1)&1 != 0
	vad := (b>>6)&1 != 0
	sr := 16000
	if b&0x20 != 0 {
		sr = 32000
	}
	return SmplTOC{
		SID:        b>>7 != 0,
		VAD:        vad,
		SampleRate: sr,
		FrameMs:    []int{10, 20, 60, 120}[(b>>3)&3],
		Voiced:     vad && bit1,
		Active:     vad || bit1,
		Flag2:      (b>>2)&1 != 0,
		Flag0:      b&1 != 0,
	}
}
```

and the KAT test `mlow/toc_test.go` (copy `toc_vectors.json` from the Rust's
`testdata/` first):

```go
package mlow

import (
	"encoding/json"
	"os"
	"testing"
)

type tocVector struct {
	B      byte `json:"b"`
	Std    bool `json:"std"`
	SID    bool `json:"sid"`
	VAD    bool `json:"vad"`
	SR     int  `json:"sr"`
	Ms     int  `json:"ms"`
	Voiced bool `json:"voiced"`
	Active bool `json:"active"`
	F2     bool `json:"f2"`
	F0     bool `json:"f0"`
}

func TestParseSmplTOCAgainstKAT(t *testing.T) {
	raw, err := os.ReadFile("testdata/toc_vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var vs []tocVector
	if err := json.Unmarshal(raw, &vs); err != nil {
		t.Fatal(err)
	}
	for _, v := range vs {
		g := ParseSmplTOC(v.B)
		if g.StdOpus != v.Std || g.SID != v.SID || g.VAD != v.VAD || g.SampleRate != v.SR ||
			g.FrameMs != v.Ms || g.Voiced != v.Voiced || g.Active != v.Active ||
			g.Flag2 != v.F2 || g.Flag0 != v.F0 {
			t.Errorf("byte 0x%02x: got %+v want %+v", v.B, g, v)
		}
	}
}
```

This is exactly the loop to repeat for `rangecoder`, `lpc`, `pitch`, `pulse`, …
until `e2e_vectors.json` decodes real frames. That is the whole project.
