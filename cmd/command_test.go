package cmd

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/smallstep/nosql"
	"github.com/smallstep/nosql/database"
	"golang.org/x/crypto/ssh"
)

// TestMain wires the package globals that init() would normally populate so
// unit tests can exercise emit/column logic without running cobra. Note that
// initChoices() here installs *fresh* tChoice values: they are NOT the pointers
// cobra bound to the flags in command_*.go's init(), because these tests never
// parse a command line. Set config.<choice>.Value directly to drive behaviour.
func TestMain(m *testing.M) {
	initLoggers()
	initChoices()
	os.Exit(m.Run())
}

// resetConfig restores a known-good default configuration before each test so
// the shared global is deterministic regardless of test ordering.
func resetConfig() {
	config = tConfig{
		emitSshFormat:  newChoice([]string{FORMAT_TABLE, FORMAT_JSON, FORMAT_MARKDOWN, FORMAT_PLAIN}, FORMAT_TABLE),
		emitX509Format: newChoice([]string{FORMAT_TABLE, FORMAT_JSON, FORMAT_MARKDOWN, FORMAT_OPENSSL, FORMAT_PLAIN}, FORMAT_TABLE),
		sortOrder:      newChoice([]string{SORT_START, SORT_FINISH}, SORT_FINISH),
		timeFormat:     newChoice([]string{TIME_ISO, TIME_SHORT}, TIME_ISO),
		serialFormat:   newChoice([]string{SERIAL_DEC, SERIAL_HEX}, SERIAL_DEC),
		showValid:      true,
	}
}

// captureStdout redirects os.Stdout for the duration of fn and returns what was
// written. Used to assert emit output without depending on test stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	fn()

	_ = w.Close()
	os.Stdout = old
	<-done
	return buf.String()
}

// emitNoPanic runs fn capturing stdout and converts any panic into a test
// failure instead of crashing the whole binary.
func emitNoPanic(t *testing.T, name string, fn func()) string {
	t.Helper()
	out := captureStdout(t, func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("%s panicked: %v", name, r)
			}
		}()
		fn()
	})
	return out
}

// newSSHSigner returns an ed25519 SSH signer.
func newSSHSigner(t *testing.T) ssh.Signer {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate ed25519: %v", err)
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatalf("ssh signer: %v", err)
	}
	return signer
}

// newSSHUserPub returns an ed25519 SSH public key (for cert.Key).
func newSSHUserPub(t *testing.T) ssh.PublicKey {
	t.Helper()
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate ed25519: %v", err)
	}
	key, err := ssh.NewPublicKey(pub)
	if err != nil {
		t.Fatalf("ssh pubkey: %v", err)
	}
	return key
}

// makeSSHCert builds a signed SSH user certificate.
func makeSSHCert(t *testing.T, serial uint64, validAfter, validBefore time.Time) ssh.Certificate {
	t.Helper()
	signer := newSSHSigner(t)
	cert := &ssh.Certificate{
		CertType:        ssh.UserCert,
		Key:             newSSHUserPub(t),
		Serial:          serial,
		ValidPrincipals: []string{"alice"},
		ValidAfter:      uint64(validAfter.Unix()),
		ValidBefore:     uint64(validBefore.Unix()),
	}
	if err := cert.SignCert(rand.Reader, signer); err != nil {
		t.Fatalf("sign cert: %v", err)
	}
	return *cert
}

// makeX509Cert builds a self-signed X.509 certificate with the given validity.
func makeX509Cert(t *testing.T, serial int64, notBefore, notAfter time.Time) *x509.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(serial),
		Subject:               pkix.Name{CommonName: "test"},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse cert: %v", err)
	}
	return cert
}

// seedSSHDB writes valid/expired/revoked SSH certs into a fresh badger DB,
// closes it, and returns the directory so export funcs can reopen it.
func seedSSHDB(t *testing.T, now time.Time) string {
	t.Helper()
	dir := t.TempDir()
	db, err := nosql.New("badgerv2", dir, database.WithValueDir(dir))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	recs := []struct {
		cert   ssh.Certificate
		revoke bool
	}{
		{makeSSHCert(t, 100, now.Add(-time.Hour), now.Add(time.Hour)), false},        // valid
		{makeSSHCert(t, 200, now.Add(-48*time.Hour), now.Add(-24*time.Hour)), false}, // expired
		{makeSSHCert(t, 300, now.Add(-time.Hour), now.Add(time.Hour)), true},         // revoked
	}
	for _, r := range recs {
		key := []byte(strconv.FormatUint(r.cert.Serial, 10))
		if err := db.Set([]byte("ssh_certs"), key, r.cert.Marshal()); err != nil {
			t.Fatalf("set ssh cert: %v", err)
		}
		if r.revoke {
			rev := `{"ProvisionerID":"p1","ReasonCode":1,"Reason":"unspecified","RevokedAt":"` +
				now.Add(-time.Minute).UTC().Format(time.RFC3339) + `"}`
			if err := db.Set([]byte("revoked_ssh_certs"), key, []byte(rev)); err != nil {
				t.Fatalf("set revocation: %v", err)
			}
		}
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close db: %v", err)
	}
	return dir
}

// seedX509DB writes valid/expired X.509 certs into a fresh badger DB, closes it,
// and returns the directory so export funcs can reopen it.
func seedX509DB(t *testing.T, now time.Time) string {
	t.Helper()
	dir := t.TempDir()
	db, err := nosql.New("badgerv2", dir, database.WithValueDir(dir))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	certs := []*x509.Certificate{
		makeX509Cert(t, 1000, now.Add(-time.Hour), now.Add(time.Hour)),        // valid
		makeX509Cert(t, 2000, now.Add(-48*time.Hour), now.Add(-24*time.Hour)), // expired
	}
	for _, c := range certs {
		if err := db.Set([]byte("x509_certs"), []byte(c.SerialNumber.String()), c.Raw); err != nil {
			t.Fatalf("set x509 cert: %v", err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close db: %v", err)
	}
	return dir
}

func TestEscapeMarkdown(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "hello", "hello"},
		{"empty", "", ""},
		{"pipe not escaped", "a|b", "a|b"},
		{"brackets", "[x](y)", `\[x\]\(y\)`},
		{"hashes", "# title", `\# title`},
		{"backtick", "`code`", "\\`code\\`"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeMarkdown(tt.in); got != tt.want {
				t.Errorf("escapeMarkdown(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestTChoiceSet(t *testing.T) {
	tests := []struct {
		name    string
		allowed []string
		def     string
		set     string
		wantErr bool
	}{
		{"valid", []string{"a", "b"}, "a", "b", false},
		{"invalid", []string{"a", "b"}, "a", "c", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newChoice(tt.allowed, tt.def)
			err := c.Set(tt.set)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Set(%q) error = %v, wantErr %v", tt.set, err, tt.wantErr)
			}
			if !tt.wantErr && c.Value != tt.set {
				t.Errorf("Value = %q, want %q", c.Value, tt.set)
			}
		})
	}
}

func TestValidityColors(t *testing.T) {
	for _, s := range []string{VALID_STR, EXPIRED_STR, REVOKED_STR} {
		if _, ok := validityColors[s]; !ok {
			t.Errorf("missing color for %q", s)
		}
	}
}

func TestAlignChars(t *testing.T) {
	for i := ALIGN_LEFT; i <= ALIGN_RIGHT; i++ {
		if _, ok := alignChars[i]; !ok {
			t.Errorf("missing align char for %d", i)
		}
	}
}

func TestMakePEM(t *testing.T) {
	raw := []byte{0x30, 0x82, 0x01, 0x00}
	got := makePEM(raw)
	want := "-----BEGIN CERTIFICATE-----\n" + string(raw) + "\n-----END CERTIFICATE-----"
	if got != want {
		t.Errorf("makePEM = %q, want %q", got, want)
	}
}

func TestParseValueToSshCertificate(t *testing.T) {
	cert := makeSSHCert(t, 42, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	marshaled := cert.Marshal()

	t.Run("valid", func(t *testing.T) {
		got := parseValueToSshCertificate(marshaled)
		if got.Serial != 42 {
			t.Errorf("Serial = %d, want 42", got.Serial)
		}
		if len(got.ValidPrincipals) == 0 || got.ValidPrincipals[0] != "alice" {
			t.Errorf("ValidPrincipals = %v, want [alice]", got.ValidPrincipals)
		}
	})

	t.Run("plain public key panics", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("expected panic for plain public key")
			}
		}()
		parseValueToSshCertificate(newSSHUserPub(t).Marshal())
	})

	t.Run("garbage panics", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("expected panic for garbage")
			}
		}()
		parseValueToSshCertificate([]byte("garbage"))
	})
}

func TestParseValueToX509Certificate(t *testing.T) {
	cert := makeX509Cert(t, 777, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))

	t.Run("valid", func(t *testing.T) {
		got := parseValueToX509Certificate(cert.Raw)
		if got.SerialNumber.Cmp(big.NewInt(777)) != 0 {
			t.Errorf("Serial = %s, want 777", got.SerialNumber)
		}
	})

	t.Run("garbage panics", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("expected panic for garbage")
			}
		}()
		parseValueToX509Certificate([]byte("garbage"))
	})
}

func TestParseRevocationAndData(t *testing.T) {
	t.Run("revocation empty", func(t *testing.T) {
		if got := parseValueToCertificateRevocation(nil); got.ProvisionerID != "" {
			t.Errorf("ProvisionerID = %q, want empty", got.ProvisionerID)
		}
	})

	t.Run("revocation populated", func(t *testing.T) {
		raw := []byte(`{"ProvisionerID":"p9","ReasonCode":2,"Reason":"keyCompromise"}`)
		got := parseValueToCertificateRevocation(raw)
		if got.ProvisionerID != "p9" || got.ReasonCode != 2 || got.Reason != "keyCompromise" {
			t.Errorf("unexpected revocation: %+v", got)
		}
	})

	t.Run("data empty", func(t *testing.T) {
		if got := parseValueToX509CertificateData(nil); got.Provisioner.ID != "" {
			t.Errorf("Provisioner.ID = %q, want empty", got.Provisioner.ID)
		}
	})

	t.Run("data populated", func(t *testing.T) {
		raw := []byte(`{"Provisioner":{"ID":"i1","Name":"n1","Type":"x509"}}`)
		got := parseValueToX509CertificateData(raw)
		if got.Provisioner.ID != "i1" || got.Provisioner.Name != "n1" || got.Provisioner.Type != "x509" {
			t.Errorf("unexpected data: %+v", got)
		}
	})
}

// TestExportSshMain runs the full SSH export against a seeded badger DB, covering
// valid/expired/revoked selection and both sort orders.
func TestExportSshMain(t *testing.T) {
	resetConfig()
	config.showExpired = true
	config.showRevoked = true
	config.sortOrder.Value = SORT_FINISH

	dir := seedSSHDB(t, time.Now())

	out := emitNoPanic(t, "exportSshMain/finish", func() { exportSshMain([]string{dir}) })
	if len(strings.TrimSpace(out)) == 0 {
		t.Error("exportSshMain produced no output")
	}

	config.sortOrder.Value = SORT_START
	out = emitNoPanic(t, "exportSshMain/start", func() { exportSshMain([]string{dir}) })
	if len(strings.TrimSpace(out)) == 0 {
		t.Error("exportSshMain (start sort) produced no output")
	}
}

// TestExportX509Main runs the full X.509 export against a seeded badger DB.
func TestExportX509Main(t *testing.T) {
	resetConfig()
	config.showExpired = true
	config.sortOrder.Value = SORT_FINISH

	dir := seedX509DB(t, time.Now())

	out := emitNoPanic(t, "exportX509Main/finish", func() { exportX509Main([]string{dir}) })
	if len(strings.TrimSpace(out)) == 0 {
		t.Error("exportX509Main produced no output")
	}

	config.sortOrder.Value = SORT_START
	out = emitNoPanic(t, "exportX509Main/start", func() { exportX509Main([]string{dir}) })
	if len(strings.TrimSpace(out)) == 0 {
		t.Error("exportX509Main (start sort) produced no output")
	}
}

// TestEmitSmoke exercises every emit path with a small non-empty dataset and
// asserts it produces output without panicking.
func TestEmitSmoke(t *testing.T) {
	resetConfig()
	now := time.Now()

	sshCert := makeSSHCert(t, 1, now.Add(-time.Hour), now.Add(time.Hour))
	sshRow := tSshCertificateWithRevocation{
		SshCertificate:              sshCert,
		Validity:                    VALID_STR,
		SshCertificateStringSerials: tCertificateStringSerials{SerialDec: "1", SerialHex: "1"},
	}
	sshCols := getSshColumns()

	x509Cert := makeX509Cert(t, 2, now.Add(-time.Hour), now.Add(time.Hour))
	x509Row := tX509CertificateProvisionerRevocation{
		X509Certificate:              *x509Cert,
		Validity:                     VALID_STR,
		X509CertificateStringSerials: tCertificateStringSerials{SerialDec: "2", SerialHex: "2"},
	}
	x509Cols := getX509Columns()

	for name, fn := range map[string]func() string{
		"table": func() string {
			return emitNoPanic(t, "ssh/table", func() {
				emitTable([]tSshCertificateWithRevocation{sshRow}, sshCols, func(r tSshCertificateWithRevocation) string { return r.SshCertificateStringSerials.SerialDec })
			})
		},
		"json": func() string {
			return emitNoPanic(t, "ssh/json", func() { emitJson([]tSshCertificateWithRevocation{sshRow}) })
		},
		"markdown": func() string {
			return emitNoPanic(t, "ssh/markdown", func() { emitMarkdown([]tSshCertificateWithRevocation{sshRow}, sshCols) })
		},
		"plain": func() string {
			return emitNoPanic(t, "ssh/plain", func() { emitPlain([]tSshCertificateWithRevocation{sshRow}, sshCols) })
		},
	} {
		t.Run("ssh/"+name, func(t *testing.T) {
			if got := fn(); len(strings.TrimSpace(got)) == 0 {
				t.Errorf("%s produced no output", name)
			}
		})
	}

	for name, fn := range map[string]func() string{
		"table": func() string {
			return emitNoPanic(t, "x509/table", func() {
				emitTable([]tX509CertificateProvisionerRevocation{x509Row}, x509Cols, func(r tX509CertificateProvisionerRevocation) string { return r.X509CertificateStringSerials.SerialDec })
			})
		},
		"json": func() string {
			return emitNoPanic(t, "x509/json", func() { emitJson([]tX509CertificateProvisionerRevocation{x509Row}) })
		},
		"markdown": func() string {
			return emitNoPanic(t, "x509/markdown", func() { emitMarkdown([]tX509CertificateProvisionerRevocation{x509Row}, x509Cols) })
		},
		"plain": func() string {
			return emitNoPanic(t, "x509/plain", func() { emitPlain([]tX509CertificateProvisionerRevocation{x509Row}, x509Cols) })
		},
		"openssl": func() string {
			return emitNoPanic(t, "x509/openssl", func() { emitX509OpenSsl([]tX509CertificateProvisionerRevocation{x509Row}) })
		},
	} {
		t.Run("x509/"+name, func(t *testing.T) {
			if got := fn(); len(strings.TrimSpace(got)) == 0 {
				t.Errorf("%s produced no output", name)
			}
		})
	}
}

// mustSSHCert is a non-Test variant of makeSSHCert for fuzz seeds (panics on error).
func mustSSHCert(serial uint64, validAfter, validBefore time.Time) ssh.Certificate {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		panic(err)
	}

	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	userPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		panic(err)
	}

	cert := &ssh.Certificate{
		CertType:        ssh.UserCert,
		Key:             userPub,
		Serial:          serial,
		ValidPrincipals: []string{"alice"},
		ValidAfter:      uint64(validAfter.Unix()),
		ValidBefore:     uint64(validBefore.Unix()),
	}
	if err := cert.SignCert(rand.Reader, signer); err != nil {
		panic(err)
	}
	return *cert
}

// mustX509Cert is a non-Test variant of makeX509Cert for fuzz seeds (panics on error).
func mustX509Cert(serial int64, notBefore, notAfter time.Time) *x509.Certificate {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(serial),
		Subject:               pkix.Name{CommonName: "test"},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		panic(err)
	}
	return cert
}

// Fuzz targets for the parsers. The production code panics on malformed input
// (via logError.Panicf); recover here so fuzzing only surfaces *unexpected*
// crashes (e.g. nil derefs) rather than the intentional panic path.
func FuzzParseValueToSshCertificate(f *testing.F) {
	cert := mustSSHCert(1, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	f.Add(cert.Marshal())
	f.Add([]byte("garbage"))
	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() { _ = recover() }()
		_ = parseValueToSshCertificate(data)
	})
}

func FuzzParseValueToX509Certificate(f *testing.F) {
	cert := mustX509Cert(1, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	f.Add(cert.Raw)
	f.Add([]byte("garbage"))
	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() { _ = recover() }()
		_ = parseValueToX509Certificate(data)
	})
}

func FuzzEscapeMarkdown(f *testing.F) {
	f.Add("hello")
	f.Add("[x](y)")
	f.Add(`a|b#c_d*e`)
	f.Fuzz(func(t *testing.T, s string) {
		_ = escapeMarkdown(s)
	})
}

// FuzzParseValueToCertificateRevocation fuzzes the JSON revocation parser. The
// production code panics (logError.Panic) on malformed JSON; recover so only
// unexpected crashes surface.
func FuzzParseValueToCertificateRevocation(f *testing.F) {
	f.Add([]byte(`{"ProvisionerID":"p1","ReasonCode":1,"Reason":"unspecified"}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`not json`))
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() { _ = recover() }()
		_ = parseValueToCertificateRevocation(data)
	})
}

// FuzzParseValueToX509CertificateData fuzzes the JSON provisioner-data parser.
func FuzzParseValueToX509CertificateData(f *testing.F) {
	f.Add([]byte(`{"Provisioner":{"ID":"i1","Name":"n1","Type":"x509"}}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`not json`))
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() { _ = recover() }()
		_ = parseValueToX509CertificateData(data)
	})
}

// TestDbTableMain runs the dbTable subcommand against a seeded badger DB and
// asserts the bucket records are emitted as JSON.
func TestDbTableMain(t *testing.T) {
	resetConfig()
	dir := t.TempDir()
	db, err := nosql.New("badgerv2", dir, database.WithValueDir(dir))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	for i, v := range []string{"alpha", "beta"} {
		key := []byte(strconv.Itoa(i))
		if err := db.Set([]byte("mybucket"), key, []byte(v)); err != nil {
			t.Fatalf("set: %v", err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	out := emitNoPanic(t, "dbTableMain", func() { dbTableMain([]string{dir, "mybucket"}) })
	// Record Bucket/Key/Value are []byte, so they marshal to base64 in the JSON.
	for _, raw := range []string{"alpha", "beta", "mybucket"} {
		want := base64.StdEncoding.EncodeToString([]byte(raw))
		if !strings.Contains(out, want) {
			t.Errorf("dbTableMain output missing %q (base64 of %q):\n%s", want, raw, out)
		}
	}
}

// TestExportMarkdownMain runs the markdown doc generator and asserts it writes files.
func TestExportMarkdownMain(t *testing.T) {
	resetConfig()
	dir := t.TempDir()
	emitNoPanic(t, "exportMarkdownMain", func() { exportMarkdownMain([]string{dir}) })

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) == 0 {
		t.Error("exportMarkdownMain wrote no files")
	}
}

// findX509Col returns the first column whose title matches, or nil.
func findX509Col(cols []tColumn[tX509CertificateProvisionerRevocation], title string) *tColumn[tX509CertificateProvisionerRevocation] {
	for i := range cols {
		if cols[i].title() == title {
			return &cols[i]
		}
	}
	return nil
}

// TestX509ColumnsContent exercises every x509 column's contentSource under a
// fully-populated config/row, covering the conditional formatting branches.
func TestX509ColumnsContent(t *testing.T) {
	resetConfig()
	config.showDNSNames = true
	config.showEmailAddresses = true
	config.showIPAddresses = true
	config.showURIs = true
	config.showCrl = true
	config.showProvisioner = true
	config.showSignatureAlgorithm = true
	config.showRevoked = true

	now := time.Now()
	cert := makeX509Cert(t, 5, now.Add(-time.Hour), now.Add(time.Hour))
	row := tX509CertificateProvisionerRevocation{
		X509Certificate:              *cert,
		Validity:                     VALID_STR,
		X509Revocation:               tCertificateRevocation{ProvisionerID: "p1", RevokedAt: now.Add(-time.Minute)},
		X509Provisioner:              tX509CertificateProvisioner{ID: "i1", Name: "provisioner-long-name", Type: "x509"},
		X509CertificateStringSerials: tCertificateStringSerials{SerialDec: "5", SerialHex: "5"},
	}

	cols := getX509Columns()
	if len(cols) == 0 {
		t.Fatal("no x509 columns")
	}
	for i, c := range cols {
		title := c.title()
		t.Run(title, func(t *testing.T) {
			got := c.contentSource(row, config) // must not panic
			if title == "Validity" && got != VALID_STR {
				t.Errorf("Validity content = %q, want %q", got, VALID_STR)
			}
			if title == "" {
				t.Errorf("column %d has empty title", i)
			}
		})
	}

	// Provisioner column truncates the name to 6 chars and prefixes the type.
	if p := findX509Col(cols, "Provisioner"); p != nil {
		got := p.contentSource(row, config)
		// Name is truncated to its first 6 chars ("provis") and prefixed with type.
		if got != "x509 provis" {
			t.Errorf("Provisioner content = %q, want %q", got, "x509 provis")
		}
	} else {
		t.Error("no Provisioner column found")
	}

	// Revoked-at renders a timestamp when a revocation exists.
	if r := findX509Col(cols, "Revoked at"); r != nil {
		if got := r.contentSource(row, config); got == "" {
			t.Error("expected non-empty Revoked at content")
		}
	}

	// Serial column honors dec vs hex.
	if s := findX509Col(cols, "Serial number"); s != nil {
		config.serialFormat.Value = SERIAL_DEC
		if got := s.contentSource(row, config); got != "5" {
			t.Errorf("serial dec = %q, want 5", got)
		}
		config.serialFormat.Value = SERIAL_HEX
		if got := s.contentSource(row, config); got != "5" {
			t.Errorf("serial hex = %q, want 5", got)
		}
	}

	// Time format short vs iso differ on the Start column.
	if st := findX509Col(cols, "Start"); st != nil {
		config.timeFormat.Value = TIME_ISO
		iso := st.contentSource(row, config)
		config.timeFormat.Value = TIME_SHORT
		short := st.contentSource(row, config)
		if iso == short {
			t.Errorf("Start iso (%q) should differ from short (%q)", iso, short)
		}
	}
}

// TestSshColumnsContent exercises every ssh column's contentSource.
func TestSshColumnsContent(t *testing.T) {
	resetConfig()
	config.showKeyId = true
	config.showHostType = true
	config.showSignatureAlgorithm = true

	now := time.Now()
	cert := makeSSHCert(t, 7, now.Add(-time.Hour), now.Add(time.Hour))
	row := tSshCertificateWithRevocation{
		SshCertificate:              cert,
		Validity:                    VALID_STR,
		SshCertificateStringSerials: tCertificateStringSerials{SerialDec: "7", SerialHex: "7"},
	}

	cols := getSshColumns()
	if len(cols) == 0 {
		t.Fatal("no ssh columns")
	}
	for i, c := range cols {
		title := c.title()
		t.Run(title, func(t *testing.T) {
			got := c.contentSource(row, config) // must not panic
			if title == "Validity" && got != VALID_STR {
				t.Errorf("Validity content = %q, want %q", got, VALID_STR)
			}
			if title == "" {
				t.Errorf("column %d has empty title", i)
			}
		})
	}
}

func TestGetCertTypeMaps(t *testing.T) {
	if certTypeMap[1] != "User" || certTypeMap[2] != "Host" {
		t.Error("unexpected cert type map")
	}
	if certTypeColors[1] == 0 || certTypeColors[2] == 0 {
		t.Error("expected non-zero cert type colors")
	}
}

func TestTChoiceType(t *testing.T) {
	c := newChoice([]string{"a", "b"}, "a")
	if got, want := c.Type(), "{a|b}"; got != want {
		t.Errorf("Type() = %q, want %q", got, want)
	}
}

// TestEmitAtMaxLoggingLevel runs every shared emitter with loggingLevel set to
// MAX_LOGGING_LEVEL so the info/spacing log branches gated on loggingLevel >= 1/2/3 are
// exercised, including emitTable's alternate Print(&tabby.Config{Spacing, Padding}) path.
func TestEmitAtMaxLoggingLevel(t *testing.T) {
	resetConfig()
	prev := loggingLevel
	loggingLevel = MAX_LOGGING_LEVEL
	defer func() { loggingLevel = prev }()

	now := time.Now()
	sshRow := tSshCertificateWithRevocation{
		SshCertificate:              makeSSHCert(t, 1, now.Add(-time.Hour), now.Add(time.Hour)),
		Validity:                    VALID_STR,
		SshCertificateStringSerials: tCertificateStringSerials{SerialDec: "1", SerialHex: "1"},
	}

	cases := map[string]func() string{
		"table": func() string {
			return emitNoPanic(t, "lvl/table", func() {
				emitTable([]tSshCertificateWithRevocation{sshRow}, getSshColumns(),
					func(r tSshCertificateWithRevocation) string { return r.SshCertificateStringSerials.SerialDec })
			})
		},
		"json": func() string {
			return emitNoPanic(t, "lvl/json", func() { emitJson([]tSshCertificateWithRevocation{sshRow}) })
		},
		"plain": func() string {
			return emitNoPanic(t, "lvl/plain", func() { emitPlain([]tSshCertificateWithRevocation{sshRow}, getSshColumns()) })
		},
		"markdown": func() string {
			return emitNoPanic(t, "lvl/markdown", func() { emitMarkdown([]tSshCertificateWithRevocation{sshRow}, getSshColumns()) })
		},
	}
	for name, fn := range cases {
		t.Run(name, func(t *testing.T) {
			if got := fn(); len(strings.TrimSpace(got)) == 0 {
				t.Errorf("%s produced no output at max logging level", name)
			}
		})
	}
}

// TestSshColumnsFormatBranches covers the serial dec/hex and time short/iso branches in the
// ssh column contentSource closures, which the default-config TestSshColumnsContent does not
// reach (it only exercises the dec + iso defaults).
func TestSshColumnsFormatBranches(t *testing.T) {
	resetConfig()
	now := time.Now()
	cert := makeSSHCert(t, 255, now.Add(-time.Hour), now.Add(time.Hour))
	row := tSshCertificateWithRevocation{
		SshCertificate:              cert,
		Validity:                    VALID_STR,
		SshCertificateStringSerials: tCertificateStringSerials{SerialDec: "255", SerialHex: "ff"},
	}

	cols := getSshColumns()
	findCol := func(title string) *tColumn[tSshCertificateWithRevocation] {
		for i := range cols {
			if cols[i].title() == title {
				return &cols[i]
			}
		}
		return nil
	}

	t.Run("serial hex", func(t *testing.T) {
		s := findCol("Serial number")
		if s == nil {
			t.Fatal("no Serial number column")
		}
		config.serialFormat.Value = SERIAL_HEX
		if got := s.contentSource(row, config); got != "ff" {
			t.Errorf("serial hex = %q, want %q", got, "ff")
		}
	})

	t.Run("start time short", func(t *testing.T) {
		st := findCol("Start")
		if st == nil {
			t.Fatal("no Start column")
		}
		config.timeFormat.Value = TIME_SHORT
		want := time.Unix(int64(cert.ValidAfter), 0).UTC().Format(time.DateOnly)
		if got := st.contentSource(row, config); got != want {
			t.Errorf("Start short = %q, want date-only %q", got, want)
		}
	})
}

// TestEmitJsonEmptySelectionIsArray guards against json emission of "null" when no
// record passes the selection filter: consumers expect an array. Reverting the
// make([]T, 0) initialisation in the handlers fails this test.
func TestEmitJsonEmptySelectionIsArray(t *testing.T) {
	resetConfig()
	config.showValid = false // showExpired/showRevoked default to false -> nothing selected

	dir := seedSSHDB(t, time.Now())

	config.emitSshFormat.Value = FORMAT_JSON
	out := emitNoPanic(t, "ssh/empty-json", func() { exportSshMain([]string{dir}) })
	if got := strings.TrimSpace(out); got != "[]" {
		t.Errorf("ssh empty selection JSON = %q, want []", got)
	}

	config.emitX509Format.Value = FORMAT_JSON
	dirX := seedX509DB(t, time.Now())
	out = emitNoPanic(t, "x509/empty-json", func() { exportX509Main([]string{dirX}) })
	if got := strings.TrimSpace(out); got != "[]" {
		t.Errorf("x509 empty selection JSON = %q, want []", got)
	}
}
