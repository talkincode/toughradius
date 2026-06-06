//go:build integration

// Package integration contains container-backed integration tests that exercise
// ToughRADIUS against a real PostgreSQL database (the production default), which
// the unit-test suite never covers because it uses in-memory SQLite.
//
// These tests are gated behind the `integration` build tag and require a running
// PostgreSQL instance described by TEST_DATABASE_* environment variables. Use the
// `make test-integration-pg` target or the provided docker-compose.test.yml to
// provision one. When the environment is absent the suite skips locally, but
// fails hard in CI (or when INTEGRATION_REQUIRED=1) so a misconfigured pipeline
// can never report a false-green.
package integration

import (
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver "pgx" for CREATE/DROP DATABASE

	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/adminapi"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins"
	parsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers/parsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// harness holds the shared state for the whole integration run: one freshly
// created PostgreSQL database, one booted admin HTTP server, and one running
// RADIUS auth/acct server pair. Individual tests use unique identifiers to stay
// isolated on the shared database rather than re-initialising the application.
type harness struct {
	cfg        *config.AppConfig
	appCtx     *app.Application
	radiusSvc  *radiusd.RadiusService
	webBaseURL string
	adminUser  string
	adminPass  string
	dbName     string
	adminDSN   string // DSN to the server's maintenance ("postgres") database
}

var h *harness

// pgEnv describes the PostgreSQL connection provided by the environment.
type pgEnv struct {
	host, user, password, adminDB string
	port                          int
}

// readPGEnv loads TEST_DATABASE_* settings. It returns ok=false when the
// environment is not configured; callers decide whether that is a skip or a
// hard failure.
func readPGEnv() (pgEnv, bool) {
	host := os.Getenv("TEST_DATABASE_HOST")
	portStr := os.Getenv("TEST_DATABASE_PORT")
	user := os.Getenv("TEST_DATABASE_USER")
	if host == "" || portStr == "" || user == "" {
		return pgEnv{}, false
	}
	port, err := net.LookupPort("tcp", portStr)
	if err != nil {
		return pgEnv{}, false
	}
	adminDB := os.Getenv("TEST_DATABASE_NAME")
	if adminDB == "" {
		adminDB = "postgres"
	}
	return pgEnv{
		host:     host,
		port:     port,
		user:     user,
		password: os.Getenv("TEST_DATABASE_PASSWORD"),
		adminDB:  adminDB,
	}, true
}

// integrationRequired reports whether a missing test database should fail the
// run instead of skipping it. This prevents false-green CI.
func integrationRequired() bool {
	return os.Getenv("CI") == "true" || os.Getenv("INTEGRATION_REQUIRED") == "1"
}

func TestMain(m *testing.M) {
	env, ok := readPGEnv()
	if !ok {
		if integrationRequired() {
			fmt.Fprintln(os.Stderr, "integration tests required but TEST_DATABASE_* is not configured")
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "skipping integration tests: TEST_DATABASE_* not configured")
		os.Exit(0)
	}

	code, err := runSuite(env, m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration suite setup failed: %v\n", err)
		os.Exit(1)
	}
	os.Exit(code)
}

// runSuite provisions a unique database, boots the servers, runs the tests, and
// guarantees teardown (database drop + temp dir removal) regardless of outcome.
func runSuite(env pgEnv, m *testing.M) (code int, err error) {
	// Unique database name per run keeps concurrent runs and reruns isolated and
	// makes teardown a single DROP DATABASE. The name pattern is also the only
	// thing the test trusts before issuing destructive DROP statements.
	dbName := fmt.Sprintf("toughradius_it_%d_%d", os.Getpid(), time.Now().UnixNano())
	adminDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		env.host, env.port, env.user, env.password, env.adminDB)

	if err := createDatabase(adminDSN, dbName); err != nil {
		return 1, fmt.Errorf("create database: %w", err)
	}
	defer func() {
		if dropErr := dropDatabase(adminDSN, dbName); dropErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to drop test database %s: %v\n", dbName, dropErr)
		}
	}()

	workdir, err := os.MkdirTemp("", "toughradius_it")
	if err != nil {
		return 1, fmt.Errorf("temp workdir: %w", err)
	}
	defer func() { _ = os.RemoveAll(workdir) }()
	for _, d := range []string{"logs", "public", "private", "backup", "data", filepath.Join("data", "metrics")} {
		if err := os.MkdirAll(filepath.Join(workdir, d), 0o755); err != nil {
			return 1, fmt.Errorf("mkdir %s: %w", d, err)
		}
	}

	webPort, err := freeTCPPort()
	if err != nil {
		return 1, err
	}
	webTLSPort, err := freeTCPPort()
	if err != nil {
		return 1, err
	}
	authPort, err := freeUDPPort()
	if err != nil {
		return 1, err
	}
	acctPort, err := freeUDPPort()
	if err != nil {
		return 1, err
	}

	cfg := &config.AppConfig{
		System: config.SysConfig{
			Appid:    "ToughRADIUS-IT",
			Location: "Asia/Shanghai",
			Workdir:  workdir,
			Debug:    false,
		},
		Database: config.DBConfig{
			Type:   "postgres",
			Host:   env.host,
			Port:   env.port,
			User:   env.user,
			Passwd: env.password,
			Name:   dbName,
		},
		Web: config.WebConfig{
			Host:    "127.0.0.1",
			Port:    webPort,
			TlsPort: webTLSPort,
			Secret:  "integration-test-secret",
		},
		Radiusd: config.RadiusdConfig{
			Enabled:  true,
			Host:     "127.0.0.1",
			AuthPort: authPort,
			AcctPort: acctPort,
			Debug:    false,
		},
		Logger: config.LogConfig{Mode: "development", FileEnable: false},
	}

	appCtx := app.NewApplication(cfg)
	appCtx.Init(cfg) // connects to Postgres and migrates domain.Tables
	defer appCtx.Release()

	adminUser, adminPass := "it-admin", "it-password-123"
	if err := seedAdminOperator(appCtx, adminUser, adminPass); err != nil {
		return 1, fmt.Errorf("seed admin operator: %w", err)
	}

	// Boot the real admin HTTP server (full middleware chain: JWT, validator,
	// appCtx injection, routing) so tests exercise the API as clients do.
	webserver.Init(appCtx)
	adminapi.Init(appCtx)
	go func() { _ = webserver.Listen(appCtx) }()
	webBase := fmt.Sprintf("http://127.0.0.1:%d", webPort)
	if err := waitForHTTP(webBase+"/ready", 10*time.Second); err != nil {
		return 1, fmt.Errorf("admin server did not become ready: %w", err)
	}

	// Boot the RADIUS auth/acct servers backed by the same Postgres database.
	registry.ResetForTest()
	_ = vendors.RegisterParser(&parsers.DefaultParser{})
	_ = vendors.RegisterParser(&parsers.HuaweiParser{})
	_ = vendors.RegisterParser(&parsers.H3CParser{})
	_ = vendors.RegisterParser(&parsers.ZTEParser{})

	radiusSvc := radiusd.NewRadiusService(appCtx)
	defer radiusSvc.Release()
	plugins.InitPlugins(appCtx, radiusSvc.SessionRepo, radiusSvc.AccountingRepo)
	authSvc := radiusd.NewAuthService(radiusSvc)
	acctSvc := radiusd.NewAcctService(radiusSvc)
	go func() { _ = radiusd.ListenRadiusAuthServer(appCtx, authSvc) }()
	go func() { _ = radiusd.ListenRadiusAcctServer(appCtx, acctSvc) }()
	if err := waitForUDP(fmt.Sprintf("127.0.0.1:%d", authPort), 10*time.Second); err != nil {
		return 1, fmt.Errorf("radius auth server did not become ready: %w", err)
	}

	h = &harness{
		cfg:        cfg,
		appCtx:     appCtx,
		radiusSvc:  radiusSvc,
		webBaseURL: webBase,
		adminUser:  adminUser,
		adminPass:  adminPass,
		dbName:     dbName,
		adminDSN:   adminDSN,
	}

	return m.Run(), nil
}

func seedAdminOperator(appCtx *app.Application, username, password string) error {
	hashed, err := common.HashPassword(password)
	if err != nil {
		return err
	}
	var count int64
	if err := appCtx.DB().Model(&domain.SysOpr{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return appCtx.DB().Create(&domain.SysOpr{
		ID:        common.UUIDint64(),
		Username:  username,
		Password:  hashed,
		Level:     "super",
		Status:    common.ENABLED,
		Realname:  "Integration Admin",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}).Error
}

func createDatabase(adminDSN, name string) error {
	db, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	if err := db.Ping(); err != nil {
		return err
	}
	// name is generated internally (toughradius_it_<pid>_<nanos>); not user input.
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %q", name))
	return err
}

func dropDatabase(adminDSN, name string) error {
	db, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	// Terminate lingering connections before dropping.
	_, _ = db.Exec("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()", name)
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %q", name))
	return err
}

func freeTCPPort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func freeUDPPort() (int, error) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenUDP("udp", addr)
	if err != nil {
		return 0, err
	}
	defer func() { _ = l.Close() }()
	return l.LocalAddr().(*net.UDPAddr).Port, nil
}

// waitForHTTP polls a URL until it returns any HTTP response or the timeout
// elapses, replacing fixed sleeps with a bounded readiness check.
func waitForHTTP(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: time.Second}
	var lastErr error
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			return nil
		}
		lastErr = err
		time.Sleep(100 * time.Millisecond)
	}
	return lastErr
}

// waitForUDP confirms the RADIUS server socket is bound by dialing it; UDP dial
// succeeds once the listener exists.
func waitForUDP(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("udp", addr, time.Second)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		lastErr = err
		time.Sleep(100 * time.Millisecond)
	}
	return lastErr
}
