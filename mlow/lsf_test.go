package mlow

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

// TestLoadSmplTables is the KAT for the LSF table asset: the protobuf blob embedded
// at the package root must decode to exactly the tables captured in the reference
// JSON dump (testdata/smpl_tables.json), proving the zlib+protobuf round-trip is
// lossless and the Go port reads the same bytes the Rust reference generated.
func TestLoadSmplTables(t *testing.T) {
	raw, err := os.ReadFile("testdata/smpl_tables.json")
	if err != nil {
		t.Fatalf("read smpl_tables.json: %v", err)
	}
	var want SmplTables // unknown gain_* keys are ignored by encoding/json
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatalf("parse smpl_tables.json: %v", err)
	}

	got := LoadSmplTables()
	if got == nil {
		t.Fatal("LoadSmplTables returned nil")
	}
	if !reflect.DeepEqual(*got, want) {
		t.Fatalf("blob tables differ from JSON capture:\n sel:    %v\n grid:   %v\n extra:  %v\n stage2 dims: %d",
			len(got.LsfSel) == len(want.LsfSel),
			reflect.DeepEqual(got.LsfGrid, want.LsfGrid),
			reflect.DeepEqual(got.LsfExtra, want.LsfExtra),
			len(got.LsfStage2))
	}
}
