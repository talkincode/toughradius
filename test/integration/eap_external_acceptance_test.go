//go:build integration && eap_accept

package integration

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/talkincode/toughradius/v9/internal/domain"
	eaphandlers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/handlers"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

const eapAcceptanceSecret = "it-eap-acceptance-secret"

type eapAcceptanceRun struct {
	StartedAt   string                  `json:"started_at"`
	FinishedAt  string                  `json:"finished_at"`
	CommitSHA   string                  `json:"commit_sha"`
	RefName     string                  `json:"ref_name"`
	WorkflowURL string                  `json:"workflow_url"`
	RunnerOS    string                  `json:"runner_os"`
	GoVersion   string                  `json:"go_version"`
	Tool        string                  `json:"tool"`
	ToolVersion string                  `json:"tool_version"`
	Verdict     string                  `json:"verdict"`
	Scenarios   []eapAcceptanceScenario `json:"scenarios"`
}

type eapAcceptanceScenario struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Method     string `json:"method"`
	Expected   string `json:"expected"`
	Status     string `json:"status"`
	Detail     string `json:"detail"`
	DurationMS int64  `json:"duration_ms"`
	Output     string `json:"output,omitempty"`
}

type eapolCase struct {
	id       string
	name     string
	method   string
	expected string
	setup    func(t *testing.T, dir, suffix string) string
}

func TestEAPExternalAcceptance(t *testing.T) {
	resultPath := os.Getenv("EAP_ACCEPTANCE_RESULT_JSON")
	if resultPath == "" {
		resultPath = filepath.Join(t.TempDir(), "eap-acceptance.json")
	}

	run := &eapAcceptanceRun{
		StartedAt:   time.Now().UTC().Format(time.RFC3339),
		CommitSHA:   envOr("GITHUB_SHA", gitValue(t, "rev-parse", "HEAD")),
		RefName:     envOr("GITHUB_REF_NAME", gitValue(t, "branch", "--show-current")),
		WorkflowURL: workflowURL(),
		RunnerOS:    envOr("RUNNER_OS", ""),
		GoVersion:   gitValue(t, "version"),
		Tool:        "eapol_test",
		Verdict:     "failed",
	}
	defer writeEAPAcceptanceResult(t, resultPath, run)

	bin := envOr("EAPOL_TEST_BIN", "eapol_test")
	binPath, err := exec.LookPath(bin)
	if err != nil {
		scenario := eapAcceptanceScenario{
			ID:       "tool-available",
			Name:     "eapol_test binary is available",
			Method:   "tooling",
			Expected: "external tool can be executed",
			Status:   missingToolStatus(),
			Detail:   "eapol_test was not found in PATH; install the eapoltest package or set EAPOL_TEST_BIN",
		}
		run.Scenarios = append(run.Scenarios, scenario)
		run.FinishedAt = time.Now().UTC().Format(time.RFC3339)
		run.Verdict = verdictFromScenarios(run.Scenarios)
		if integrationRequired() || os.Getenv("EAP_ACCEPTANCE_REQUIRED") == "1" {
			t.Fatalf("%s", scenario.Detail)
		}
		t.Skip(scenario.Detail)
	}
	run.ToolVersion = eapolVersion(t, binPath)

	seedLoopbackAcceptanceNAS(t)
	restoreEapMethod(t)

	suffix := uniqueSuffix()
	profileID := seedProfile(t, "it-eap-accept-profile-"+suffix)
	dir := t.TempDir()

	cases := []eapolCase{
		{
			id:       "eap-tls-valid-cert",
			name:     "EAP-TLS valid client certificate",
			method:   "EAP-TLS",
			expected: "Access-Accept",
			setup: func(t *testing.T, dir, suffix string) string {
				ca := newEAPTLSTestCA(t, "IT External EAP-TLS Root CA "+suffix)
				serverCert := ca.issueServer(t, "radius.example.com")
				username := "it-ext-tls-alice-" + suffix + "@example.com"
				clientCert := ca.issueClient(t, "alice", username)
				seedEAPTLSUser(t, profileID, username)
				configureEAPTLS(t, serverCert, ca.certPEM())
				return writeEAPTLSConfig(t, dir, "tls-valid", username, ca, clientCert)
			},
		},
		{
			id:       "eap-tls-untrusted-cert",
			name:     "EAP-TLS untrusted client certificate",
			method:   "EAP-TLS",
			expected: "Access-Reject",
			setup: func(t *testing.T, dir, suffix string) string {
				serverCA := newEAPTLSTestCA(t, "IT External EAP-TLS Server CA "+suffix)
				rogueCA := newEAPTLSTestCA(t, "IT External EAP-TLS Rogue CA "+suffix)
				serverCert := serverCA.issueServer(t, "radius.example.com")
				username := "it-ext-tls-mallory-" + suffix + "@example.com"
				rogueCert := rogueCA.issueClient(t, "mallory", username)
				seedEAPTLSUser(t, profileID, username)
				configureEAPTLS(t, serverCert, serverCA.certPEM())
				return writeEAPTLSConfig(t, dir, "tls-untrusted", username, serverCA, rogueCert)
			},
		},
		{
			id:       "peap-mschapv2-valid",
			name:     "PEAP/MSCHAPv2 valid credentials",
			method:   "PEAP/MSCHAPv2",
			expected: "Access-Accept",
			setup: func(t *testing.T, dir, suffix string) string {
				ca := newEAPTLSTestCA(t, "IT External PEAP Root CA "+suffix)
				serverCert := ca.issueServer(t, "radius.example.com")
				username := "it-ext-peap-alice-" + suffix
				password := "it-ext-peap-pass-" + suffix
				seedPEAPUser(t, profileID, username, password)
				configurePEAPWithCA(t, serverCert, ca.certPEM())
				return writePasswordEAPConfig(t, dir, "peap-valid", "PEAP", `phase1="peapver=0 tls_disable_tlsv1_3=1"`+"\n"+`phase2="auth=MSCHAPV2"`, username, password, ca)
			},
		},
		{
			id:       "peap-mschapv2-wrong-password",
			name:     "PEAP/MSCHAPv2 wrong password",
			method:   "PEAP/MSCHAPv2",
			expected: "Access-Reject",
			setup: func(t *testing.T, dir, suffix string) string {
				ca := newEAPTLSTestCA(t, "IT External PEAP Reject Root CA "+suffix)
				serverCert := ca.issueServer(t, "radius.example.com")
				username := "it-ext-peap-mallory-" + suffix
				password := "it-ext-peap-correct-" + suffix
				seedPEAPUser(t, profileID, username, password)
				configurePEAPWithCA(t, serverCert, ca.certPEM())
				return writePasswordEAPConfig(t, dir, "peap-reject", "PEAP", `phase1="peapver=0 tls_disable_tlsv1_3=1"`+"\n"+`phase2="auth=MSCHAPV2"`, username, password+"-wrong", ca)
			},
		},
		{
			id:       "ttls-pap-valid",
			name:     "EAP-TTLS/PAP valid credentials",
			method:   "EAP-TTLS/PAP",
			expected: "Access-Accept",
			setup: func(t *testing.T, dir, suffix string) string {
				ca := newEAPTLSTestCA(t, "IT External TTLS PAP Root CA "+suffix)
				serverCert := ca.issueServer(t, "radius.example.com")
				username := "it-ext-ttls-pap-alice-" + suffix
				password := "it-ext-ttls-pap-pass-" + suffix
				seedTTLSUser(t, profileID, username, password)
				configureTTLSWithCA(t, serverCert, ca.certPEM())
				return writePasswordEAPConfig(t, dir, "ttls-pap-valid", "TTLS", `phase1="tls_disable_tlsv1_3=1"`+"\n"+`phase2="auth=PAP"`, username, password, ca)
			},
		},
		{
			id:       "ttls-mschapv2-valid",
			name:     "EAP-TTLS/MSCHAPv2 valid credentials",
			method:   "EAP-TTLS/MSCHAPv2",
			expected: "Access-Accept",
			setup: func(t *testing.T, dir, suffix string) string {
				ca := newEAPTLSTestCA(t, "IT External TTLS MSCHAP Root CA "+suffix)
				serverCert := ca.issueServer(t, "radius.example.com")
				username := "it-ext-ttls-mschap-alice-" + suffix
				password := "it-ext-ttls-mschap-pass-" + suffix
				seedTTLSUser(t, profileID, username, password)
				configureTTLSWithCA(t, serverCert, ca.certPEM())
				return writePasswordEAPConfig(t, dir, "ttls-mschap-valid", "TTLS", `phase1="tls_disable_tlsv1_3=1"`+"\n"+`phase2="auth=MSCHAPV2"`, username, password, ca)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			cfg := tc.setup(t, dir, suffix)
			run.Scenarios = append(run.Scenarios, runEAPOLCase(t, binPath, tc, cfg))
		})
	}

	run.Scenarios = append(run.Scenarios, eapAcceptanceScenario{
		ID:       "malformed-config-negative-path",
		Name:     "Malformed external EAP client config",
		Method:   "tooling",
		Expected: "documented skip",
		Status:   "skipped",
		Detail:   "Skipped intentionally: eapol_test parser failures do not exercise ToughRADIUS over RADIUS/EAP. Negative server behavior is covered by untrusted certificate and wrong password scenarios.",
	})
	run.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	run.Verdict = verdictFromScenarios(run.Scenarios)
	if run.Verdict != "accepted" {
		t.Fatalf("EAP external acceptance verdict: %s", run.Verdict)
	}
}

func seedLoopbackAcceptanceNAS(t *testing.T) {
	t.Helper()
	require.NoError(t, h.appCtx.DB().Where("ipaddr = ?", "127.0.0.1").Delete(&domain.NetNas{}).Error)
	nas := &domain.NetNas{
		ID:         common.UUIDint64(),
		Name:       "it-eap-acceptance-nas",
		Identifier: "it-eap-acceptance-nas",
		Ipaddr:     "127.0.0.1",
		Secret:     eapAcceptanceSecret,
		VendorCode: "0",
		Status:     common.ENABLED,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	require.NoError(t, h.appCtx.DB().Create(nas).Error)
}

func configurePEAPWithCA(t *testing.T, serverCert tls.Certificate, caPEM []byte) {
	t.Helper()
	configurePEAP(t, serverCert)
	caPath := filepath.Join(t.TempDir(), "peap-ca.pem")
	require.NoError(t, os.WriteFile(caPath, caPEM, 0o600))
	require.NoError(t, h.appCtx.ConfigMgr().Set("radius", eaphandlers.SettingEapTlsCaFile, caPath))
}

func configureTTLSWithCA(t *testing.T, serverCert tls.Certificate, caPEM []byte) {
	t.Helper()
	configureTTLS(t, serverCert)
	caPath := filepath.Join(t.TempDir(), "ttls-ca.pem")
	require.NoError(t, os.WriteFile(caPath, caPEM, 0o600))
	require.NoError(t, h.appCtx.ConfigMgr().Set("radius", eaphandlers.SettingEapTlsCaFile, caPath))
}

func writeEAPTLSConfig(t *testing.T, dir, name, identity string, ca *eapTLSTestCA, clientCert tls.Certificate) string {
	t.Helper()
	caPath := filepath.Join(dir, name+"-ca.pem")
	certPath := filepath.Join(dir, name+"-client.pem")
	keyPath := filepath.Join(dir, name+"-client-key.pem")
	require.NoError(t, os.WriteFile(caPath, ca.certPEM(), 0o600))
	require.NoError(t, os.WriteFile(certPath, certificatePEM(t, clientCert), 0o600))
	require.NoError(t, os.WriteFile(keyPath, privateKeyPEM(t, clientCert.PrivateKey), 0o600))

	return writeEAPOLConfig(t, dir, name+".conf", fmt.Sprintf(`network={
    key_mgmt=IEEE8021X
    eap=TLS
    identity=%s
    ca_cert=%s
    client_cert=%s
    private_key=%s
    phase1="tls_disable_tlsv1_3=1"
}
`, strconv.Quote(identity), strconv.Quote(caPath), strconv.Quote(certPath), strconv.Quote(keyPath)))
}

func writePasswordEAPConfig(t *testing.T, dir, name, method, phase, identity, password string, ca *eapTLSTestCA) string {
	t.Helper()
	caPath := filepath.Join(dir, name+"-ca.pem")
	require.NoError(t, os.WriteFile(caPath, ca.certPEM(), 0o600))
	return writeEAPOLConfig(t, dir, name+".conf", fmt.Sprintf(`network={
    key_mgmt=IEEE8021X
    eap=%s
    identity=%s
    anonymous_identity=%s
    password=%s
    ca_cert=%s
    %s
}
`, method, strconv.Quote(identity), strconv.Quote("anonymous"), strconv.Quote(password), strconv.Quote(caPath), phase))
}

func writeEAPOLConfig(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	return path
}

func runEAPOLCase(t *testing.T, bin string, tc eapolCase, cfgPath string) eapAcceptanceScenario {
	t.Helper()
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	args := []string{
		"-c" + cfgPath,
		"-a127.0.0.1",
		"-p" + strconv.Itoa(h.cfg.Radiusd.AuthPort),
		"-s" + eapAcceptanceSecret,
		"-r1",
		"-t20",
	}
	cmd := exec.CommandContext(ctx, bin, args...)
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)
	out := sanitizeEAPOLOutput(string(output))

	status, detail := classifyEAPOLResult(tc.expected, out, err, ctx.Err())
	return eapAcceptanceScenario{
		ID:         tc.id,
		Name:       tc.name,
		Method:     tc.method,
		Expected:   tc.expected,
		Status:     status,
		Detail:     detail,
		DurationMS: duration.Milliseconds(),
		Output:     out,
	}
}

func classifyEAPOLResult(expected, output string, runErr, ctxErr error) (string, string) {
	if errors.Is(ctxErr, context.DeadlineExceeded) {
		return "failed", "eapol_test timed out"
	}
	success := strings.Contains(output, "SUCCESS")
	failure := strings.Contains(output, "FAILURE") || strings.Contains(output, "CTRL-EVENT-EAP-FAILURE")
	if expected == "Access-Accept" && runErr == nil && success {
		return "passed", "external supplicant received the expected Access-Accept"
	}
	if expected == "Access-Reject" && runErr != nil && failure {
		return "passed", "external supplicant was rejected as expected"
	}
	if expected == "Access-Reject" && failure && !success {
		return "passed", "external supplicant was rejected as expected"
	}
	if runErr != nil {
		return "failed", fmt.Sprintf("eapol_test returned %v", runErr)
	}
	return "failed", "eapol_test completed but did not emit the expected result marker"
}

func eapolVersion(t *testing.T, bin string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "-v")
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		return "unknown"
	}
	line := strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
	if line == "" {
		return "unknown"
	}
	return line
}

func sanitizeEAPOLOutput(output string) string {
	replacements := []string{
		eapAcceptanceSecret, "[radius-secret]",
	}
	safe := strings.NewReplacer(replacements...).Replace(output)
	lines := strings.Split(safe, "\n")
	filtered := lines[:0]
	for _, line := range lines {
		if strings.Contains(line, "private_key=") || strings.Contains(line, "password=") {
			continue
		}
		filtered = append(filtered, line)
	}
	safe = strings.TrimSpace(strings.Join(filtered, "\n"))
	const maxLen = 6000
	if len(safe) > maxLen {
		safe = safe[:maxLen] + "\n...[truncated]"
	}
	return safe
}

func missingToolStatus() string {
	if integrationRequired() || os.Getenv("EAP_ACCEPTANCE_REQUIRED") == "1" {
		return "failed"
	}
	return "skipped"
}

func verdictFromScenarios(scenarios []eapAcceptanceScenario) string {
	passed := 0
	for _, scenario := range scenarios {
		switch scenario.Status {
		case "failed":
			return "failed"
		case "passed":
			passed++
		}
	}
	if passed == 0 {
		return "incomplete"
	}
	return "accepted"
}

func writeEAPAcceptanceResult(t *testing.T, path string, run *eapAcceptanceRun) {
	t.Helper()
	if run.FinishedAt == "" {
		run.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	}
	if run.Verdict == "" {
		run.Verdict = verdictFromScenarios(run.Scenarios)
	}
	data, err := json.MarshalIndent(run, "", "  ")
	if err != nil {
		t.Logf("marshal EAP acceptance result: %v", err)
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Logf("mkdir EAP acceptance result dir: %v", err)
		return
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		t.Logf("write EAP acceptance result: %v", err)
	}
}

func workflowURL() string {
	server := os.Getenv("GITHUB_SERVER_URL")
	repo := os.Getenv("GITHUB_REPOSITORY")
	runID := os.Getenv("GITHUB_RUN_ID")
	if server == "" || repo == "" || runID == "" {
		return ""
	}
	return strings.TrimRight(server, "/") + "/" + repo + "/actions/runs/" + runID
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func gitValue(t *testing.T, args ...string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", args...)
	if args[0] != "version" {
		cmd = exec.CommandContext(ctx, "git", args...)
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}
