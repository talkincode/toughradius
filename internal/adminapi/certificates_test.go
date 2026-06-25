package adminapi

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"gorm.io/gorm"
)

// genSelfSignedPEM generates a self-signed ECDSA certificate and returns the
// certificate and private key as PEM strings, suitable for exercising the
// certificate import/validation paths.
func genSelfSignedPEM(t *testing.T, cn string) (certPEM, keyPEM string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{CommonName: cn},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)
	certPEM = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
	keyPEM = string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}))
	return certPEM, keyPEM
}

// seedCertificate inserts a certificate directly into the database for list/get
// style tests.
func seedCertificate(t *testing.T, db *gorm.DB, name, certType, certPEM, keyPEM string) *domain.SysCert {
	t.Helper()
	cert := &domain.SysCert{
		ID:         common.UUIDint64(),
		Name:       name,
		CertType:   certType,
		Cert:       certPEM,
		PrivateKey: keyPEM,
		Subject:    "CN=" + name,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	require.NoError(t, db.Create(cert).Error)
	return cert
}

func TestCreateCertificate(t *testing.T) {
	db, e, appCtx := CreateTestAppContext(t)
	require.NoError(t, db.AutoMigrate(&domain.SysCert{}))

	serverCert, serverKey := genSelfSignedPEM(t, "server.example.com")
	caCert, _ := genSelfSignedPEM(t, "ca.example.com")
	_, otherKey := genSelfSignedPEM(t, "other.example.com")

	tests := []struct {
		name           string
		body           string
		expectedStatus int
		expectedError  string
		check          func(*testing.T, *domain.SysCert)
	}{
		{
			name: "server cert with matching key",
			body: mustJSON(map[string]string{
				"name": "srv1", "cert_type": "server", "cert": serverCert, "private_key": serverKey,
			}),
			expectedStatus: http.StatusOK,
			check: func(t *testing.T, c *domain.SysCert) {
				assert.NotZero(t, c.ID)
				assert.True(t, c.HasKey)
				assert.Contains(t, c.Subject, "server.example.com")
				assert.NotEmpty(t, c.Fingerprint)
				assert.False(t, c.NotAfter.IsZero())
			},
		},
		{
			name: "ca cert without key",
			body: mustJSON(map[string]string{
				"name": "ca1", "cert_type": "ca", "cert": caCert,
			}),
			expectedStatus: http.StatusOK,
			check: func(t *testing.T, c *domain.SysCert) {
				assert.False(t, c.HasKey)
				assert.Contains(t, c.Subject, "ca.example.com")
			},
		},
		{
			name: "server cert missing key",
			body: mustJSON(map[string]string{
				"name": "srv-nokey", "cert_type": "server", "cert": serverCert,
			}),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "KEY_REQUIRED",
		},
		{
			name: "key does not match cert",
			body: mustJSON(map[string]string{
				"name": "srv-mismatch", "cert_type": "server", "cert": serverCert, "private_key": otherKey,
			}),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "KEY_MISMATCH",
		},
		{
			name: "invalid certificate PEM",
			body: mustJSON(map[string]string{
				"name": "bad", "cert_type": "ca", "cert": "-----BEGIN CERTIFICATE-----\nnotbase64\n-----END CERTIFICATE-----",
			}),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_CERT",
		},
		{
			name:           "invalid json",
			body:           `{invalid}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST",
		},
		{
			name: "missing required cert_type",
			body: mustJSON(map[string]string{
				"name": "no-type", "cert": caCert,
			}),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/system/certificate", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)

			require.NoError(t, CreateCertificate(c))
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// The private key must never be serialized in the response body.
			assert.NotContains(t, rec.Body.String(), "PRIVATE KEY")

			if tt.expectedStatus == http.StatusOK {
				cert := decodeCert(t, rec)
				if tt.check != nil {
					tt.check(t, cert)
				}
			} else if tt.expectedError != "" {
				assert.Contains(t, rec.Body.String(), tt.expectedError)
			}
		})
	}

	t.Run("duplicate name", func(t *testing.T) {
		body := mustJSON(map[string]string{"name": "dup", "cert_type": "ca", "cert": caCert})
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/system/certificate", strings.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)
			require.NoError(t, CreateCertificate(c))
			if i == 0 {
				assert.Equal(t, http.StatusOK, rec.Code)
			} else {
				assert.Equal(t, http.StatusConflict, rec.Code)
				assert.Contains(t, rec.Body.String(), "NAME_EXISTS")
			}
		}
	})
}

func TestListCertificates(t *testing.T) {
	db, e, appCtx := CreateTestAppContext(t)
	require.NoError(t, db.AutoMigrate(&domain.SysCert{}))

	sc, sk := genSelfSignedPEM(t, "srv")
	cc, _ := genSelfSignedPEM(t, "ca")
	seedCertificate(t, db, "server-a", "server", sc, sk)
	seedCertificate(t, db, "ca-a", "ca", cc, "")

	t.Run("list all", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/system/certificate", nil)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)
		require.NoError(t, ListCertificates(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "PRIVATE KEY")

		var resp struct {
			Data  []domain.SysCert `json:"data"`
			Total int64            `json:"total"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, int64(2), resp.Total)
		for _, c := range resp.Data {
			if c.Name == "server-a" {
				assert.True(t, c.HasKey)
			}
			if c.Name == "ca-a" {
				assert.False(t, c.HasKey)
			}
		}
	})

	t.Run("filter by cert_type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/system/certificate?cert_type=ca", nil)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)
		require.NoError(t, ListCertificates(c))
		var resp struct {
			Data  []domain.SysCert `json:"data"`
			Total int64            `json:"total"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, int64(1), resp.Total)
		require.Len(t, resp.Data, 1)
		assert.Equal(t, "ca-a", resp.Data[0].Name)
	})
}

func TestGetCertificate(t *testing.T) {
	db, e, appCtx := CreateTestAppContext(t)
	require.NoError(t, db.AutoMigrate(&domain.SysCert{}))
	sc, sk := genSelfSignedPEM(t, "srv")
	seeded := seedCertificate(t, db, "server-a", "server", sc, sk)

	t.Run("found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/system/certificate/"+strconv.FormatInt(seeded.ID, 10), nil)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)
		c.SetParamNames("id")
		c.SetParamValues(strconv.FormatInt(seeded.ID, 10))
		require.NoError(t, GetCertificate(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "PRIVATE KEY")
		cert := decodeCert(t, rec)
		assert.True(t, cert.HasKey)
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/system/certificate/999", nil)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)
		c.SetParamNames("id")
		c.SetParamValues("999")
		require.NoError(t, GetCertificate(c))
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/system/certificate/abc", nil)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)
		c.SetParamNames("id")
		c.SetParamValues("abc")
		require.NoError(t, GetCertificate(c))
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestUpdateCertificate(t *testing.T) {
	db, e, appCtx := CreateTestAppContext(t)
	require.NoError(t, db.AutoMigrate(&domain.SysCert{}))
	sc, sk := genSelfSignedPEM(t, "srv")
	cc, _ := genSelfSignedPEM(t, "ca")
	seeded := seedCertificate(t, db, "server-a", "server", sc, sk)
	seedCertificate(t, db, "ca-a", "ca", cc, "")

	t.Run("rename and remark", func(t *testing.T) {
		body := mustJSON(map[string]string{"name": "server-renamed", "remark": "prod"})
		req := httptest.NewRequest(http.MethodPut, "/api/v1/system/certificate/"+strconv.FormatInt(seeded.ID, 10), strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)
		c.SetParamNames("id")
		c.SetParamValues(strconv.FormatInt(seeded.ID, 10))
		require.NoError(t, UpdateCertificate(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		cert := decodeCert(t, rec)
		assert.Equal(t, "server-renamed", cert.Name)
		assert.Equal(t, "prod", cert.Remark)
		assert.True(t, cert.HasKey)
	})

	t.Run("rename to existing name conflicts", func(t *testing.T) {
		body := mustJSON(map[string]string{"name": "ca-a"})
		req := httptest.NewRequest(http.MethodPut, "/api/v1/system/certificate/"+strconv.FormatInt(seeded.ID, 10), strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)
		c.SetParamNames("id")
		c.SetParamValues(strconv.FormatInt(seeded.ID, 10))
		require.NoError(t, UpdateCertificate(c))
		assert.Equal(t, http.StatusConflict, rec.Code)
		assert.Contains(t, rec.Body.String(), "NAME_EXISTS")
	})

	t.Run("not found", func(t *testing.T) {
		body := mustJSON(map[string]string{"remark": "x"})
		req := httptest.NewRequest(http.MethodPut, "/api/v1/system/certificate/999", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)
		c.SetParamNames("id")
		c.SetParamValues("999")
		require.NoError(t, UpdateCertificate(c))
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestDeleteCertificate(t *testing.T) {
	db, e, appCtx := CreateTestAppContext(t)
	require.NoError(t, db.AutoMigrate(&domain.SysCert{}))
	cc, _ := genSelfSignedPEM(t, "ca")
	seeded := seedCertificate(t, db, "ca-a", "ca", cc, "")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/system/certificate/"+strconv.FormatInt(seeded.ID, 10), nil)
	rec := httptest.NewRecorder()
	c := CreateTestContext(e, db, req, rec, appCtx)
	c.SetParamNames("id")
	c.SetParamValues(strconv.FormatInt(seeded.ID, 10))
	require.NoError(t, DeleteCertificate(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var count int64
	db.Model(&domain.SysCert{}).Where("id = ?", seeded.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestExportCertificate(t *testing.T) {
	db, e, appCtx := CreateTestAppContext(t)
	require.NoError(t, db.AutoMigrate(&domain.SysCert{}))
	sc, sk := genSelfSignedPEM(t, "srv")
	cc, _ := genSelfSignedPEM(t, "ca")
	server := seedCertificate(t, db, "server-a", "server", sc, sk)
	ca := seedCertificate(t, db, "ca-a", "ca", cc, "")

	t.Run("cert only by default", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/system/certificate/"+strconv.FormatInt(server.ID, 10)+"/export", nil)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)
		c.SetParamNames("id")
		c.SetParamValues(strconv.FormatInt(server.ID, 10))
		require.NoError(t, ExportCertificate(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get(echo.HeaderContentDisposition), "server-a.pem")
		assert.Contains(t, rec.Body.String(), "BEGIN CERTIFICATE")
		assert.NotContains(t, rec.Body.String(), "PRIVATE KEY")
	})

	t.Run("include key", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/system/certificate/"+strconv.FormatInt(server.ID, 10)+"/export?include_key=true", nil)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)
		c.SetParamNames("id")
		c.SetParamValues(strconv.FormatInt(server.ID, 10))
		require.NoError(t, ExportCertificate(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "BEGIN CERTIFICATE")
		assert.Contains(t, rec.Body.String(), "PRIVATE KEY")
	})

	t.Run("include key on cert without key", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/system/certificate/"+strconv.FormatInt(ca.ID, 10)+"/export?include_key=true", nil)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)
		c.SetParamNames("id")
		c.SetParamValues(strconv.FormatInt(ca.ID, 10))
		require.NoError(t, ExportCertificate(c))
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "NO_KEY")
	})
}

func mustJSON(m map[string]string) string {
	b, _ := json.Marshal(m)
	return string(b)
}

func decodeCert(t *testing.T, rec *httptest.ResponseRecorder) *domain.SysCert {
	t.Helper()
	var resp Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	dataBytes, _ := json.Marshal(resp.Data)
	var cert domain.SysCert
	require.NoError(t, json.Unmarshal(dataBytes, &cert))
	return &cert
}
