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

// smplTablesBlob is the runtime LSF CDF table set (func 3559 output) as a
// zlib-compressed SmplLsfTables protobuf — the byte-identical blob the reference
// embeds (its testdata/smpl_tables.bin). It keeps the reference's filename and
// lives at the package root, not testdata, because it is a production asset —
// mirroring smpl_cc_blob.bin in mem.go.
//
//go:embed smpl_tables.bin
var smplTablesBlob []byte

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
		zr, err := zlib.NewReader(bytes.NewReader(smplTablesBlob))
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
// previous internal frame. PrevLagSamples, PrevLagblk and PrevLagidx are encoder-only
// (pitch-search/lag-predictor continuity) and unused by the decoder.
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
	PrevLagblk     int32
	PrevLagidx     int32
}

// SmplAdvanceLsfState advances the LSF predictor mirror exactly as the
// encode/decode path does for an internal frame with the given stage-1 selector:
// on a no-match (intf 0, or stage1 differs from the previous frame) it resets the
// four pitch/LTP predictor fields to -1, then records PrevStage1/PrevMatch. The
// encoder analysis runs this so its PrevLag tracks what the entropy encoder will
// compute (driving the abs-vs-delta lag pick).
func SmplAdvanceLsfState(st *SmplLsfState, intf int, stage1 int32) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/c697c36ffa7875c304ceea9154be30b66cada914/wacore/src/voip/mlow/smpl_decode.rs#L192-L205
	m := intf != 0 && stage1 == st.PrevStage1
	if !m {
		st.PrevGainIdx = -1
		st.PrevFiltIdx = -1
		st.PrevLag = -1
		st.PrevFracLag = -1
		st.PrevLagblk = -1
		st.PrevLagidx = -1
	}
	st.PrevStage1 = stage1
	st.PrevMatch = m
	st.HavePrev = true
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
	var idx SmplLsfIndices

	// Read 1 — stage-1 selector. Frame 0 uses dedicated row 0; later frames pick
	// row 2 if the previous stage-1 was nonzero, else row 1.
	sel := 0
	if intf != 0 {
		if st.PrevStage1 != 0 {
			sel = 2
		} else {
			sel = 1
		}
	}
	stage1 := dec.DecodeCDF(t.LsfSel[sel])
	idx.Stage1 = stage1

	// match := (not the first frame) && stage1 == prev. On a no-match the four
	// pitch/LTP predictor fields reset to -1, recorded BEFORE PrevStage1 is updated.
	m := intf != 0 && stage1 == st.PrevStage1
	if !m {
		st.PrevGainIdx = -1
		st.PrevFiltIdx = -1
		st.PrevLag = -1
		st.PrevFracLag = -1
	}
	st.PrevStage1 = stage1

	// Read 2 — stage-1 grid. Outer select on match, inner on the current stage1.
	var gridCDF []uint16
	switch {
	case m && stage1 != 0:
		gridCDF = t.LsfGrid.Match1
	case m:
		gridCDF = t.LsfGrid.Match1Alt
	case stage1 != 0:
		gridCDF = t.LsfGrid.Match0Alt
	default:
		gridCDF = t.LsfGrid.Match0
	}
	grid := dec.DecodeCDF(gridCDF)
	idx.Grid = grid
	st.PrevMatch = m
	st.HavePrev = true

	// Read 3 — 16 stage-2 residuals, each coeff k from LsfStage2[stage1][config][grid][k].
	st2 := t.LsfStage2[int(stage1)][config][int(grid)]
	for k := 0; k < 16; k++ {
		c := st2[k]
		idx.Stage2[k] = dec.DecodeCDF(c)
		idx.StageNraw[k] = int32(len(c)) - 2
	}

	// Read 4 — the 3-symbol "extra" LSF CDF, which always fires for the 1:1 path.
	idx.Extra = dec.DecodeCDF(t.LsfExtra)
	return idx
}
