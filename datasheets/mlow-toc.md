# Datasheet: `mlow/toc`

Parses the first byte of a media frame to decide how to decode it and whether it
is a standard-Opus frame to route elsewhere. Media layer; every inbound frame
starts here.

**Validation vector:** `toc_vectors.json` (256 entries, one per byte value). Copy
it verbatim into `mlow/testdata/`.

**Reference pinned at:** `41095d4e6ba4610e054e9ede3af1d5e88a83faee` (`wacore/src/voip/mlow/toc.rs`)

## Reference source (verbatim — authoritative)

```rust
//! MLow "smpl_toc" — the first byte of a bare MLow frame (WASM func 3544). The smpl TOC is only
//! valid when `(b & 0xC0) != 0xC0`; `(b & 0xC0) == 0xC0` marks a STANDARD Opus/CELT TOC instead,
//! which we route to stock libopus. Ported from the Go reference (`mlow_toc.go`).
//!
//! Bit layout (LSB = bit0): bit7=SID(DTX/CNG), bit6=VAD, bit5=internal rate(0→16k,1→32k),
//! bits4:3→frame size index into {10,20,60,120}ms, bit2=flag2, bit1=voiced-enable, bit0=flag0.

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub(crate) struct MlowToc {
    pub std_opus: bool,
    pub sid: bool,
    pub vad: bool,
    pub sample_rate: i32,
    pub frame_ms: i32,
    pub voiced: bool,
    pub active: bool,
    pub flag2: bool,
    pub flag0: bool,
}

/// Frame duration (ms) of a standard Opus packet from its TOC config field `b>>3` (RFC 6716
/// Table 2). 2.5 ms is rounded up — the smpl path only needs an output length for CNG frames.
fn standard_opus_frame_ms(b: u8) -> i32 {
    let config = b >> 3;
    if config < 12 {
        [10, 20, 40, 60][(config & 3) as usize] // SILK NB/MB/WB
    } else if config < 16 {
        [10, 20][((config - 12) & 1) as usize] // Hybrid
    } else {
        match config & 3 {
            0 => 3, // 2.5 ms rounded up
            1 => 5,
            2 => 10,
            _ => 20,
        }
    }
}

/// Parse the smpl TOC byte (WASM func 3544).
pub(crate) fn parse_mlow_toc(b: u8) -> MlowToc {
    if b & 0xC0 == 0xC0 {
        return MlowToc {
            std_opus: true,
            sid: false,
            vad: false,
            sample_rate: 16000,
            frame_ms: standard_opus_frame_ms(b),
            voiced: false,
            active: false,
            flag2: false,
            flag0: false,
        };
    }
    let bit1 = (b >> 1) & 1 != 0;
    let vad = (b >> 6) & 1 != 0;
    MlowToc {
        std_opus: false,
        sid: b >> 7 != 0,
        vad,
        sample_rate: if b & 0x20 != 0 { 32000 } else { 16000 },
        frame_ms: [10, 20, 60, 120][((b >> 3) & 3) as usize],
        voiced: vad && bit1,
        active: vad || bit1,
        flag2: (b >> 2) & 1 != 0,
        flag0: b & 1 != 0,
    }
}
```

Test (verbatim): exhaustive over all 256 byte values against
`testdata/toc_vectors.json`, asserting every field (`std, sid, vad, sr, ms, voiced,
active, f2, f0`).

```rust
#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::Value;

    #[test]
    fn toc_matches_go_all_256() {
        let recs: Value =
            serde_json::from_str(include_str!("testdata/toc_vectors.json")).expect("toc_vectors");
        let arr = recs.as_array().unwrap();
        assert_eq!(arr.len(), 256);
        for rec in arr {
            let b = rec["b"].as_u64().unwrap() as u8;
            let t = parse_mlow_toc(b);
            assert_eq!(t.std_opus, rec["std"].as_bool().unwrap(), "std b=0x{b:02x}");
            // ... asserts every field: sid, vad, sr, ms, voiced, active, f2, f0
        }
    }
}
```

## Go envelope (signatures only)

```go
package mlow

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

func ParseSmplTOC(b byte) SmplTOC
```

## Implementation suggestions (guidance, not authoritative)

- `i32` fields → Go `int`. Pure function, no receiver, no error return.
- Every array index in the reference is masked to its valid range, so plain slice
  indexing is safe (no bounds checks needed).
- Sample rate is `16000`/`32000` (Hz); frame duration is in ms.
- KAT covers the entire 256-value input space, so once it is green there is no
  unobserved input — a safe first module to land.
