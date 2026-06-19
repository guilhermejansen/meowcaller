package mlow

import (
	"bytes"
	"compress/zlib"
	_ "embed"
	"io"
	"sync"

	"google.golang.org/protobuf/proto"

	"github.com/purpshell/meowcaller/mlow/internal/tables"
)

// smplLsfTablesBlob is the runtime LSF CDF table set (func 3559 output) as a
// zlib-compressed SmplLsfTables protobuf — the byte-identical blob the reference
// embeds (testdata/smpl_tables.bin). It lives at the package root, not testdata,
// because it is a production asset, mirroring smpl_cc_blob.bin in mem.go.
//
//go:embed smpl_lsf_tables.bin
var smplLsfTablesBlob []byte

// LsfGrid holds the four stage-1 grid CDFs, selected by (match, stage1!=0).
type LsfGrid struct {
	Match1    []uint16 `json:"match1"`
	Match1Alt []uint16 `json:"match1_alt"`
	Match0    []uint16 `json:"match0"`
	Match0Alt []uint16 `json:"match0_alt"`
}

// SmplTables is the runtime-built CDF table set the LSF decode reads. The smpl LSF
// coding is Meta-specific (not stock SILK CB1): a 2-way stage-1 codebook selector,
// a stage-1 grid index, then 16 stage-2 residuals keyed by
// LsfStage2[stage1][config][grid][coeff]. The gain CDFs the decoder uses come from
// the heap window (SmplMem g_nrg), not these fields. The json tags exist only so
// the KAT can parse the captured smpl_tables.json dump for cross-checking.
type SmplTables struct {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/c697c36ffa7875c304ceea9154be30b66cada914/wacore/src/voip/mlow/smpl_decode.rs#L23-L31
	LsfSel    [][]uint16       `json:"lsf_sel"`
	LsfGrid   LsfGrid          `json:"lsf_grid"`
	LsfStage2 [][][][][]uint16 `json:"lsf_stage2"` // [stage1][config][grid][coeff] -> cumulative CDF
	LsfExtra  []uint16         `json:"lsf_extra"`
}

var (
	smplTablesOnce sync.Once
	smplTables     *SmplTables
)

// LoadSmplTables inflates and protobuf-decodes the embedded LSF table blob once and
// returns the shared, read-only set. The u16 CDF entries are stored widened to u32
// on the wire (protobuf has no u16) and narrowed back here.
func LoadSmplTables() *SmplTables {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/c697c36ffa7875c304ceea9154be30b66cada914/wacore/src/voip/mlow/smpl_decode.rs#L35-L43
	smplTablesOnce.Do(func() {
		zr, err := zlib.NewReader(bytes.NewReader(smplLsfTablesBlob))
		if err != nil {
			panic("mlow: open lsf table blob: " + err.Error())
		}
		raw, err := io.ReadAll(zr)
		if err != nil {
			panic("mlow: inflate lsf table blob: " + err.Error())
		}
		_ = zr.Close()
		var pb tables.SmplLsfTables
		if err := proto.Unmarshal(raw, &pb); err != nil {
			panic("mlow: decode lsf table blob: " + err.Error())
		}
		smplTables = pbToSmplTables(&pb)
	})
	return smplTables
}

// cdfU16 narrows a wire CDF (u32) back to the native u16 cumulative table.
func cdfU16(v []uint32) []uint16 {
	out := make([]uint16, len(v))
	for i, x := range v {
		out[i] = uint16(x)
	}
	return out
}

// pbToSmplTables converts the decoded protobuf message into the runtime table set,
// reconstructing the literal [stage1][config][grid][coeff] nesting.
func pbToSmplTables(p *tables.SmplLsfTables) *SmplTables {
	t := &SmplTables{}
	for _, c := range p.GetLsfSel() {
		t.LsfSel = append(t.LsfSel, cdfU16(c.GetV()))
	}
	g := p.GetLsfGrid()
	t.LsfGrid = LsfGrid{
		Match1:    cdfU16(g.GetMatch1()),
		Match1Alt: cdfU16(g.GetMatch1Alt()),
		Match0:    cdfU16(g.GetMatch0()),
		Match0Alt: cdfU16(g.GetMatch0Alt()),
	}
	for _, s1 := range p.GetLsfStage2() {
		var c1 [][][][]uint16
		for _, s2 := range s1.GetConfig() {
			var c2 [][][]uint16
			for _, s3 := range s2.GetGrid() {
				var c3 [][]uint16
				for _, cd := range s3.GetCoeff() {
					c3 = append(c3, cdfU16(cd.GetV()))
				}
				c2 = append(c2, c3)
			}
			c1 = append(c1, c2)
		}
		t.LsfStage2 = append(t.LsfStage2, c1)
	}
	t.LsfExtra = cdfU16(p.GetLsfExtra())
	return t
}

// SmplLsfState is the cross-internal-frame decoder state. The LSF block resets the
// pitch/LTP predictor fields to -1 whenever the stage-1 selector does not match the
// previous internal frame. PrevLagSamples is encoder-only (pitch-search continuity
// bias) and unused by the decoder.
type SmplLsfState struct {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/c697c36ffa7875c304ceea9154be30b66cada914/wacore/src/voip/mlow/smpl_decode.rs#L169-L186
	PrevStage1     int32
	PrevMatch      bool
	HavePrev       bool
	PrevGainIdx    int32
	PrevFiltIdx    int32
	PrevLag        int32
	PrevFracLag    int32
	PrevLagSamples float32
}

// SmplAdvanceLsfState advances the LSF predictor mirror exactly as the
// encode/decode path does for an internal frame with the given stage-1 selector:
// on a no-match (intf 0, or stage1 differs from the previous frame) it resets the
// four pitch/LTP predictor fields to -1, then records PrevStage1/PrevMatch. The
// encoder analysis runs this so its PrevLag tracks what the entropy encoder will
// compute (driving the abs-vs-delta lag pick).
func SmplAdvanceLsfState(st *SmplLsfState, intf int, stage1 int32) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/c697c36ffa7875c304ceea9154be30b66cada914/wacore/src/voip/mlow/smpl_decode.rs#L192-L205
	// TODO
	// agent suggestion: m := intf != 0 && stage1 == st.PrevStage1; if !m reset the
	//   four Prev{Gain,Filt,Lag,FracLag}Idx/Lag fields to -1; then set PrevStage1,
	//   PrevMatch=m, HavePrev=true.
	// human input:
	panic("mlow: SmplAdvanceLsfState not yet implemented (scaffold)")
}

// SmplLsfIndices is the decoded per-internal-frame LSF index set. StageNraw[k] is
// the raw symbol count for coefficient k (len(cdf)-2), carried for the dequantizer.
type SmplLsfIndices struct {
	Stage1    int32
	Grid      int32
	Stage2    [16]int32
	StageNraw [16]int32
	Extra     int32
}

// DecodeSmplLsf decodes the LSF block of one internal frame (the first block of the
// frame body). config is the smpl config (0/1); intf is the internal-frame index
// (0,1,2) within the 60 ms packet. It mutates st, applying the no-match predictor
// reset in place exactly where the reference does.
//
// The four reads, in order: (1) the stage-1 selector — intf 0 uses dedicated row 0,
// later frames pick row 1/2 by the previous frame's stage-1; (2) the stage-1 grid,
// whose CDF is selected by (match, current stage1!=0); (3) 16 stage-2 residuals,
// each coeff k from its own CDF LsfStage2[stage1][config][grid][k]; (4) the 3-symbol
// "extra" LSF CDF, which always fires for our 1:1 path.
func DecodeSmplLsf(
	dec *RangeDecoder,
	t *SmplTables,
	st *SmplLsfState,
	config int,
	intf int,
) SmplLsfIndices {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/c697c36ffa7875c304ceea9154be30b66cada914/wacore/src/voip/mlow/smpl_decode.rs#L218-L291
	// TODO
	// agent suggestion: mirror decode_smpl_lsf read-for-read — sel pick → DecodeCDF
	//   on LsfSel[sel]; compute m := intf!=0 && stage1==st.PrevStage1 and reset the
	//   predictor on !m (BEFORE recording st.PrevStage1=stage1); grid CDF pick on
	//   (m, stage1!=0); set st.PrevMatch=m, st.HavePrev=true; loop 16 coeffs filling
	//   Stage2[k]=DecodeCDF(c) and StageNraw[k]=len(c)-2; Extra=DecodeCDF(LsfExtra).
	//   Index tables with int(stage1)/int(grid) at the use site.
	// human input:
	panic("mlow: DecodeSmplLsf not yet implemented (scaffold)")
}
