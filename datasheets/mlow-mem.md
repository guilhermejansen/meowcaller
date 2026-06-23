# Datasheet: `mlow/mem`

The embedded heap-memory window and table-base pointers used to replicate exact
pointer arithmetic when reading the runtime-built CDF tables, plus the fixed cosine
approximation table for the LSF root search. Media layer; a data/table provider that
the pulse/pitch/gains and LSF decode paths read through.

**Validation vector:** `smpl_tables.json` — the runtime-built CDF tables this module
backs (the decode paths that consume the heap window and cosine table are pinned by
it). The heap window loads from `smpl_cc_blob.json`.

> The heap window is stored as a zlib-compressed protobuf `HeapWindow`
> (`tables.proto`), loaded via `smpl_tables_blob::load_blob_prost`. meowcaller
> embeds the same `smpl_cc_blob.bin` and decodes it through the shared `tables.proto`
> schema (compiled to `mlow/internal/tables`).

**Reference pinned at:** `41095d4e6ba4610e054e9ede3af1d5e88a83faee` (`wacore/src/voip/mlow/smpl_mem.rs`, `wacore/src/voip/mlow/tables.proto`, `wacore/src/voip/mlow/silk_lsf_cos_tab.rs`)

## Reference source (verbatim — authoritative)

### Heap-memory window and table-base pointers

```rust
//! MLow heap window for the pitch lag/contour reads (`smpl_pitch.rs` Group D), the only consumer
//! still addressing the WASM heap by absolute pointer. Groups A/B/C/E moved to the logical `CcTables`
//! (`cc_seed.bin`), as did the p6!=0 LR gain/filter/weight tables. The contour window proper (pcfg
//! header + per-contour records, lag/frac/delta CDFs, contour_map, delta-lag bounds) is BUILT from
//! `pitch_seed.bin` at load (`build_smpl_mem`), so nothing of it is stored. Everything sits at the
//! original absolute addresses so the existing pointer arithmetic still lands.

use std::sync::OnceLock;

/// `tables.proto` Region.
#[derive(Clone, PartialEq, prost::Message)]
struct SmplMemRegion {
    #[prost(uint32, tag = "1")]
    base: u32,
    #[prost(bytes = "vec", tag = "2")]
    data: Vec<u8>,
}

/// `tables.proto` HeapWindow; the runtime window built by `build_smpl_mem`.
#[derive(Clone, PartialEq, prost::Message)]
pub(crate) struct SmplMem {
    #[prost(message, repeated, tag = "1")]
    regions: Vec<SmplMemRegion>,
    #[prost(uint32, tag = "2")]
    pub(crate) g_cc: u32,
    #[prost(uint32, tag = "3")]
    pub(crate) g_nrg: u32,
    #[prost(uint32, tag = "4")]
    pub(crate) g_pitch: u32,
    #[prost(uint32, tag = "5")]
    pub(crate) g_clk: u32,
}

static SMPL_MEM: OnceLock<SmplMem> = OnceLock::new();

/// Parse the full heap-window JSON dump into a `SmplMem` (the generator's carve source and the
/// byte-identical oracle while Group D still reads the heap).
#[cfg(test)]
pub(crate) fn parse_smpl_mem_json(s: &str) -> SmplMem {
    #[derive(serde::Deserialize)]
    struct RawRegion {
        base: u32,
        b64: String,
    }
    #[derive(serde::Deserialize)]
    struct Raw {
        regions: Vec<RawRegion>,
        g_cc: u32,
        g_nrg: u32,
        g_pitch: u32,
        clk: u32,
    }
    use base64::Engine;
    let raw: Raw = serde_json::from_str(s).expect("smpl_cc_blob.json must parse");
    let engine = base64::engine::general_purpose::STANDARD;
    let regions = raw
        .regions
        .into_iter()
        .map(|r| SmplMemRegion {
            base: r.base,
            data: engine.decode(r.b64).expect("smpl_cc_blob region b64"),
        })
        .collect();
    SmplMem {
        regions,
        g_cc: raw.g_cc,
        g_nrg: raw.g_nrg,
        g_pitch: raw.g_pitch,
        g_clk: raw.clk,
    }
}

// Fixed WASM-build globals for the Group-D heap layout. The window is built at these absolute
// addresses so `smpl_pitch.rs`'s pointer arithmetic lands unchanged.
const G_CLK: u32 = 0xb9f9a8;
const G_PITCH: u32 = 0xb9d378;
const PCFG: u32 = G_CLK.wrapping_add(0x5704);
// pcfg header pointers (@pcfg+0x56e0): num_contours, contour_map, lag_cdf, frac_base, then three the
// consumer never reads, then delta_cdf. Fixed addresses for this build.
const HDR_CONTOUR_MAP: u32 = 0xe7c10;
const HDR_LAG_CDF: u32 = 0xbaa7b0;
const HDR_FRAC_BASE: u32 = 0xbaa9be;
const HDR_DELTA_CDF: u32 = 0xbab13e;
const HDR_UNUSED: [u32; 3] = [0xe7d20, 0xe7ef0, 0xe8096];
const DELTA_BOUNDS_ADDR: u32 = 0xe7ef0;
const NUM_CONTOURS: usize = 217;

/// Build the pitch lag/contour window from the seed. Reproduces the carved window byte-for-byte at
/// every address the consumer reads.
fn build_smpl_mem() -> SmplMem {
    let seed: super::smpl_pitch_seed::PitchSeed =
        super::smpl_tables_blob::load_blob_prost(include_bytes!("testdata/pitch_seed.bin"));
    let w = seed.build_contour_window();

    let mut regions = Vec::with_capacity(6);
    let push = |regions: &mut Vec<SmplMemRegion>, base: u32, data: Vec<u8>| {
        regions.push(SmplMemRegion { base, data });
    };

    // region 0 @pcfg+0x1d38: 217 records (lags[8] | seglens[8] | seg_count, 0x44 each), a 4-byte
    // gap (=NUM_BLOCKTRACKS), then the 8-pointer header.
    let mut r0 = Vec::with_capacity(NUM_CONTOURS * 0x44 + 4 + 32);
    for (lags, seglens) in &w.records {
        let sc = lags.len();
        for i in 0..8 {
            r0.extend_from_slice(&(*lags.get(i).unwrap_or(&0) as i32).to_le_bytes());
        }
        for i in 0..8 {
            r0.extend_from_slice(&(*seglens.get(i).unwrap_or(&0) as i32).to_le_bytes());
        }
        r0.extend_from_slice(&(sc as i32).to_le_bytes());
    }
    r0.extend_from_slice(&187u32.to_le_bytes());
    for h in [
        NUM_CONTOURS as u32,
        HDR_CONTOUR_MAP,
        HDR_LAG_CDF,
        HDR_FRAC_BASE,
        HDR_UNUSED[0],
        HDR_UNUSED[1],
        HDR_UNUSED[2],
        HDR_DELTA_CDF,
    ] {
        r0.extend_from_slice(&h.to_le_bytes());
    }
    push(&mut regions, PCFG.wrapping_add(0x1d38), r0);

    // CDF tables (each its own region; the dead inter-table gaps the carve included are unread).
    let u16_bytes =
        |v: &[u32]| -> Vec<u8> { v.iter().flat_map(|&x| (x as u16).to_le_bytes()).collect() };
    push(&mut regions, HDR_LAG_CDF, u16_bytes(&w.lag_cdf));
    let frac: Vec<u32> = w.frac_cmfs.iter().flatten().copied().collect();
    push(&mut regions, HDR_FRAC_BASE, u16_bytes(&frac));
    let delta: Vec<u32> = w.delta_cmfs.iter().flatten().copied().collect();
    push(&mut regions, HDR_DELTA_CDF, u16_bytes(&delta));

    // contour_map[217].
    push(&mut regions, HDR_CONTOUR_MAP, w.contour_map.clone());

    // delta-lag bounds: firstblock_range u8 pairs, then the carve's 2 trailing pad bytes.
    let mut bounds: Vec<u8> = w
        .firstblock_range
        .iter()
        .flat_map(|&[lo, hi]| [lo as u8, hi as u8])
        .collect();
    bounds.extend_from_slice(&[0, 0]);
    push(&mut regions, DELTA_BOUNDS_ADDR, bounds);

    SmplMem {
        regions,
        g_cc: 0,
        g_nrg: 0,
        g_pitch: G_PITCH,
        g_clk: G_CLK,
    }
}

pub(crate) fn load_smpl_mem() -> &'static SmplMem {
    SMPL_MEM.get_or_init(build_smpl_mem)
}

impl SmplMem {
    /// Region containing `[addr, addr+n)` and the byte offset of `addr` within it, or `None`.
    fn region_for(&self, addr: u32, n: usize) -> Option<(&[u8], usize)> {
        for r in &self.regions {
            if addr >= r.base && (addr - r.base) as usize + n <= r.data.len() {
                return Some((&r.data, (addr - r.base) as usize));
            }
        }
        None
    }

    pub(crate) fn u8(&self, addr: u32) -> u8 {
        self.region_for(addr, 1).map_or(0, |(d, off)| d[off])
    }

    pub(crate) fn u16(&self, addr: u32) -> u16 {
        self.region_for(addr, 2)
            .map_or(0, |(d, off)| u16::from_le_bytes([d[off], d[off + 1]]))
    }

    pub(crate) fn i16(&self, addr: u32) -> i16 {
        self.u16(addr) as i16
    }

    pub(crate) fn u32(&self, addr: u32) -> u32 {
        self.region_for(addr, 4).map_or(0, |(d, off)| {
            u32::from_le_bytes([d[off], d[off + 1], d[off + 2], d[off + 3]])
        })
    }

    pub(crate) fn i32(&self, addr: u32) -> i32 {
        self.u32(addr) as i32
    }

    /// Materialize the n-entry cumulative u16 CDF at WASM address `addr` (for `decode_cdf`). Entries
    /// outside the window read as 0, matching the WASM's out-of-region fallback.
    pub(crate) fn cdf_at(&self, addr: u32, n: usize) -> Vec<u16> {
        (0..n)
            .map(|i| self.u16(addr.wrapping_add((i as u32) * 2)))
            .collect()
    }

    /// Raw `[addr, addr+2n)` byte slice when the whole CDF window sits inside one region (the common
    /// case), so callers can read it in place instead of allocating a `Vec<u16>`. `None` at a region
    /// boundary or out of range, where the zero-fill semantics of `cdf_at` must apply; callers fall
    /// back to `cdf_at` there.
    pub(crate) fn cdf_bytes(&self, addr: u32, n: usize) -> Option<&[u8]> {
        let (data, off) = self.region_for(addr, n * 2)?;
        Some(&data[off..off + n * 2])
    }
}

/// Load the full heap dump from the (gitignored) `smpl_cc_blob.json` oracle, or `None` if absent
/// (CI has only the carved `.bin`; the byte-identical gates run where the JSON lives).
#[cfg(test)]
pub(crate) fn try_load_full_heap() -> Option<SmplMem> {
    let s = std::fs::read_to_string("src/voip/mlow/testdata/smpl_cc_blob.json").ok()?;
    Some(parse_smpl_mem_json(&s))
}
```

### Heap-window protobuf schema

```proto
syntax = "proto3";

package mlow.tables;

// The embedded MLow heap window: heap byte regions at their absolute addresses
// plus the table-base pointers. Carries the residual pitch-contour window (the
// only reads still served from the heap). Stored zlib-compressed.
message HeapWindow {
  repeated Region regions = 1;
  uint32 g_cc = 2;
  uint32 g_nrg = 3;
  uint32 g_pitch = 4;
  uint32 g_clk = 5;
}

// One contiguous heap region: its absolute base address and bytes.
message Region {
  uint32 base = 1;
  bytes data = 2;
}
```

### LSF cosine approximation table

```rust
// SILK cosine approximation table (silk_LSFCosTab_FIX_Q12, Q12, 129 entries) for the A2NLSF root
// search.
#[rustfmt::skip]
const SILK_LSF_COS_TAB_FIX_Q12: [i32; 129] = [
    8192, 8190, 8182, 8170, 8152, 8130, 8104, 8072,
    8034, 7994, 7946, 7896, 7840, 7778, 7714, 7644,
    7568, 7490, 7406, 7318, 7226, 7128, 7026, 6922,
    6812, 6698, 6580, 6458, 6332, 6204, 6070, 5934,
    5792, 5648, 5502, 5352, 5198, 5040, 4880, 4718,
    4552, 4382, 4212, 4038, 3862, 3684, 3502, 3320,
    3136, 2948, 2760, 2570, 2378, 2186, 1990, 1794,
    1598, 1400, 1202, 1002, 802, 602, 402, 202,
    0, -202, -402, -602, -802, -1002, -1202, -1400,
    -1598, -1794, -1990, -2186, -2378, -2570, -2760, -2948,
    -3136, -3320, -3502, -3684, -3862, -4038, -4212, -4382,
    -4552, -4718, -4880, -5040, -5198, -5352, -5502, -5648,
    -5792, -5934, -6070, -6204, -6332, -6458, -6580, -6698,
    -6812, -6922, -7026, -7128, -7226, -7318, -7406, -7490,
    -7568, -7644, -7714, -7778, -7840, -7896, -7946, -7994,
    -8034, -8072, -8104, -8130, -8152, -8170, -8182, -8190,
    -8192,
];
```

## Go envelope (signatures only)

```go
package mlow

type smplMemRegion struct {
	base uint32
	data []byte
}

type SmplMem struct {
	regions []smplMemRegion
	GCC     uint32
	GNrg    uint32
	GPitch  uint32
	GClk    uint32
}

func LoadSmplMem() *SmplMem

func (m *SmplMem) regionFor(addr uint32, n int) (data []byte, off int, ok bool)
func (m *SmplMem) U8(addr uint32) uint8
func (m *SmplMem) U16(addr uint32) uint16
func (m *SmplMem) I16(addr uint32) int16
func (m *SmplMem) U32(addr uint32) uint32
func (m *SmplMem) I32(addr uint32) int32
func (m *SmplMem) CDFAt(addr uint32, n int) []uint16

var silkLSFCosTabFIXQ12 = [129]int32{
	// 129 entries, see verbatim source
}
```

## Implementation suggestions (guidance, not authoritative)

- `region_for` returns `Option<(&[u8], usize)>`; the idiomatic Go form is a
  multi-value return with a trailing `ok bool` (or a `nil` slice). Every accessor must
  preserve the "out of region reads as 0" fallback — that is observable behavior the
  CDF materialization relies on.
- Width mapping: `u8`/`u16`/`u32` → `uint8`/`uint16`/`uint32`; `i16`/`i32` are the
  signed reinterpretations of the same bytes (`int16(m.U16(addr))`, `int32(m.U32(addr))`),
  not separate reads.
- All multi-byte reads are little-endian; use `encoding/binary.LittleEndian.Uint16` /
  `Uint32` over the offset slice rather than hand-assembling shifts.
- The blob loads once and is shared immutably; `sync.OnceLock` maps to a `sync.Once`
  (or a package-level `var ... = loadSmplMem()` if eager init is acceptable). The
  decoded regions are read-only after init.
- The base64 region payloads use standard encoding (`base64.StdEncoding`). TODO(human):
  confirm the JSON field names (`base`/`b64` per region; top-level `g_cc`, `g_nrg`,
  `g_pitch`, `clk`) match what your struct tags expect, since `clk` maps to the
  `GClk` field, not `g_clk`.
- The cosine table is a fixed 129-entry `int32` array (Q12, symmetric around index 64
  where the value is 0); a package-level array literal is fine — no runtime init.
```