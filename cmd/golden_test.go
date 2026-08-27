package cmd

// Golden-output integration tests: run the real export handlers against the
// checked-in step-ca badger database fixture (../db) and assert that the full
// JSON emission matches the golden results in ../test_results/ record-for-record.
// Both fixtures are git-ignored, so every test here skips cleanly when they are
// absent (e.g. in CI) and runs fully on a developer machine where both exist.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// decodeJSONRecords unmarshals a JSON array of records into []map[string]any.
// UseNumber keeps large integers (e.g. RSA PublicKey.N in the x509 fixture, which
// overflows float64) as json.Number instead of failing the decode.
func decodeJSONRecords(t *testing.T, label string, data []byte) []map[string]any {
	t.Helper()

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	var records []map[string]any
	if err := dec.Decode(&records); err != nil {
		t.Fatalf("%s: unmarshal JSON: %v\nprefix: %.200q", label, err, data)
	}
	return records
}

// stripValidity removes the "Validity" field from each record so golden
// comparisons are not affected by time-dependent validity classification:
// whether a certificate is currently Valid or Expired depends on time.Now() at
// emission time and would otherwise make the test flaky as certificates age.
func stripValidity(records []map[string]any) {
	for _, r := range records {
		delete(r, "Validity")
	}
}

// compareGoldenRecords unmarshals gotJSON (captured CLI output) and the golden
// file at goldenPath into record slices, strips the time-dependent "Validity"
// field from both sides, and asserts they are identical record-for-record. The
// first differing record is reported with a JSON snippet of each side.
func compareGoldenRecords(t *testing.T, label, gotJSON, goldenPath string) {
	t.Helper()

	goldenBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Skipf("golden file %s not present: %v", goldenPath, err)
	}

	got := decodeJSONRecords(t, label+"/emitted", []byte(gotJSON))
	want := decodeJSONRecords(t, label+"/golden", goldenBytes)

	stripValidity(got)
	stripValidity(want)

	if len(got) != len(want) {
		t.Fatalf("%s: record count = %d, want %d (golden %s)", label, len(got), len(want), goldenPath)
	}

	for i := range want {
		if !reflect.DeepEqual(got[i], want[i]) {
			gotSnippet, _ := json.Marshal(got[i])
			wantSnippet, _ := json.Marshal(want[i])
			t.Fatalf("%s: record %d differs.\ngot:  %.500q\nwant: %.500q", label, i, gotSnippet, wantSnippet)
		}
	}
}

// TestExportSshMainGolden runs `step-badger sshCerts -e -r --emit=json` against
// the real database fixture and asserts the emitted JSON matches the golden
// result in test_results/sshCerts.json, ignoring the "Validity" field.
func TestExportSshMainGolden(t *testing.T) {
	db := fixtureDB(t)

	resetConfig()
	config.showExpired = true // -e
	config.showRevoked = true // -r
	// showValid stays at its default (true), matching the command line.
	config.emitSshFormat.Value = FORMAT_JSON // --emit=json

	out := emitNoPanic(t, "exportSshMain/golden", func() {
		exportSshMain([]string{db})
	})
	if out == "" {
		t.Fatal("exportSshMain produced no output")
	}

	compareGoldenRecords(t, "sshCerts", out, filepath.Join("..", "test_results", "sshCerts.json"))
}

// TestExportX509MainGolden runs `step-badger x509Certs -e -r --emit=json` against
// the real database fixture and asserts the emitted JSON matches the golden
// result in test_results/x509Certs.json, ignoring the "Validity" field.
func TestExportX509MainGolden(t *testing.T) {
	db := fixtureDB(t)

	resetConfig()
	config.showExpired = true // -e
	config.showRevoked = true // -r
	// showValid stays at its default (true), matching the command line.
	config.emitX509Format.Value = FORMAT_JSON // --emit=json

	out := emitNoPanic(t, "exportX509Main/golden", func() {
		exportX509Main([]string{db})
	})
	if out == "" {
		t.Fatal("exportX509Main produced no output")
	}

	compareGoldenRecords(t, "x509Certs", out, filepath.Join("..", "test_results", "x509Certs.json"))
}
