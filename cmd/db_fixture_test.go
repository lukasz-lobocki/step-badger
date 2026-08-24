package cmd

// Integration tests that run the real export/table handlers against the
// checked-in step-ca badger database fixture (db/ at the repo root). The fixture
// is git-ignored so it is absent in CI; every test here skips cleanly when it is
// missing and runs fully on a developer machine where the 89 MB snapshot exists.

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// fixtureOnce + fixturePath ensure the (costly) copy of the 89 MB fixture happens
// at most once per test run and is shared by all fixture-backed tests. They must be
// package-level (not local to fixtureDB) so every call reads the same result.
var (
	fixtureOnce sync.Once
	fixturePath string
)

// fixtureDB returns an isolated, writable copy of the db/ test fixture, created
// once per test run and shared by all fixture tests. Copying (rather than opening
// db/ in place) honours badger's single-writer limitation documented in README: it
// never mutates the source snapshot and avoids lock-file churn in the repo dir. The
// calling test is skipped if the fixture is not present.
func fixtureDB(t *testing.T) string {
	t.Helper()

	fixtureOnce.Do(func() {
		// Tests run with working dir = the package dir (cmd/), so the fixture at
		// the repo root is ../db relative to here.
		const src = "../db"
		if _, err := os.Stat(filepath.Join(src, "MANIFEST")); err != nil {
			return // leave fixturePath empty -> caller skips
		}
		dst, err := os.MkdirTemp("", "step-badger-fixture-*")
		if err != nil {
			t.Logf("create fixture temp dir: %v", err)
			return
		}
		entries, err := os.ReadDir(src)
		if err != nil {
			t.Logf("read fixture dir: %v", err)
			return
		}
		for _, e := range entries {
			if !e.Type().IsRegular() {
				continue
			}
			data, err := os.ReadFile(filepath.Join(src, e.Name()))
			if err != nil {
				t.Logf("read fixture file %s: %v", e.Name(), err)
				continue
			}
			if err := os.WriteFile(filepath.Join(dst, e.Name()), data, 0o600); err != nil {
				t.Logf("write fixture file %s: %v", e.Name(), err)
				return
			}
		}
		fixturePath = dst
	})

	if fixturePath == "" {
		t.Skip("step-ca db fixture (../db) not present; skipping integration test")
	}
	return fixturePath
}

// TestDbTableMainFixture runs the dbTable subcommand against real buckets,
// covering the large-record JSON marshal path that synthetic seeds cannot.
func TestDbTableMainFixture(t *testing.T) {
	db := fixtureDB(t)

	for _, tc := range []struct {
		name   string
		bucket string
	}{
		{"ssh_certs", "ssh_certs"},
		{"x509_certs", "x509_certs"},
		{"revoked_x509_certs", "revoked_x509_certs"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resetConfig()
			out := emitNoPanic(t, "dbTableMain/"+tc.bucket, func() {
				dbTableMain([]string{db, tc.bucket})
			})
			if strings.TrimSpace(out) == "" {
				t.Fatalf("dbTableMain(%q) produced no output", tc.bucket)
			}
			// Records marshal as a JSON array of {Bucket,Key,Value}; Bucket/Key/Value
			// are []byte so they appear base64-encoded (see TestDbTableMain).
			if !strings.HasPrefix(strings.TrimSpace(out), "[") {
				t.Errorf("expected JSON array output, got prefix: %.120q", out)
			}
			wantBucket := base64.StdEncoding.EncodeToString([]byte(tc.bucket))
			if !strings.Contains(out, wantBucket) {
				t.Errorf("output does not echo bucket %q (base64 %q)", tc.bucket, wantBucket)
			}
		})
	}
}

// TestExportSshMainFixture runs the full SSH export against the real database in
// every supported format. With showValid/showExpired/showRevoked all on, the whole
// 1636-cert population is parsed, revocation-looked-up, and emitted.
func TestExportSshMainFixture(t *testing.T) {
	db := fixtureDB(t)

	for _, tc := range []struct {
		name   string
		format string
	}{
		{"table", FORMAT_TABLE},
		{"json", FORMAT_JSON},
		{"markdown", FORMAT_MARKDOWN},
		{"plain", FORMAT_PLAIN},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resetConfig()
			config.showValid = true
			config.showExpired = true
			config.showRevoked = true
			config.emitSshFormat.Value = tc.format

			out := emitNoPanic(t, "exportSshMain/"+tc.format, func() {
				exportSshMain([]string{db})
			})
			if strings.TrimSpace(out) == "" {
				t.Fatalf("exportSshMain(%q) produced no output", tc.format)
			}
		})
	}
}

// TestExportX509MainFixture runs the full X.509 export against the real database in
// every supported format, including openssl. This exercises getX509Revocation and
// getX509CertificateData "record found" branches across ~1200 real certs.
func TestExportX509MainFixture(t *testing.T) {
	db := fixtureDB(t)

	for _, tc := range []struct {
		name   string
		format string
	}{
		{"table", FORMAT_TABLE},
		{"json", FORMAT_JSON},
		{"markdown", FORMAT_MARKDOWN},
		{"openssl", FORMAT_OPENSSL},
		{"plain", FORMAT_PLAIN},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resetConfig()
			config.showValid = true
			config.showExpired = true
			config.showRevoked = true
			config.emitX509Format.Value = tc.format

			out := emitNoPanic(t, "exportX509Main/"+tc.format, func() {
				exportX509Main([]string{db})
			})
			if strings.TrimSpace(out) == "" {
				t.Fatalf("exportX509Main(%q) produced no output", tc.format)
			}
		})
	}
}

// TestCheckLogginglevelLevels drives checkLogginglevel across every valid level to
// cover the >=1 info-emission branch. The >MAX_LOGGING_LEVEL branch calls
// logError.Fatalln (os.Exit) and is intentionally not exercised in-process.
func TestCheckLogginglevelLevels(t *testing.T) {
	for _, level := range []int{0, 1, 2, MAX_LOGGING_LEVEL} {
		t.Run("level_"+strconv.Itoa(level), func(t *testing.T) {
			loggingLevel = level
			emitNoPanic(t, "checkLogginglevel", func() {
				checkLogginglevel([]string{"sshCerts", "./db"})
			})
			loggingLevel = 0
		})
	}
}
