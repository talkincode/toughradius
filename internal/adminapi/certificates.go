package adminapi

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"go.uber.org/zap"
)

// certPayload represents the certificate import request structure. The cert and
// private_key fields carry PEM-encoded material; private_key is optional for CA
// trust anchors but required for server certificates that act as the EAP/TLS
// server identity.
type certPayload struct {
	Name       string `json:"name" validate:"required,min=1,max=128"`
	CertType   string `json:"cert_type" validate:"required,oneof=server ca"`
	Cert       string `json:"cert" validate:"required"`
	PrivateKey string `json:"private_key" validate:"omitempty"`
	Remark     string `json:"remark" validate:"omitempty,max=512"`
}

// certUpdatePayload relaxes validation rules for partial updates. Empty cert and
// private_key fields leave the stored material untouched, so operators can rename
// a certificate or edit its remark without re-uploading the PEM material.
type certUpdatePayload struct {
	Name       string `json:"name" validate:"omitempty,min=1,max=128"`
	CertType   string `json:"cert_type" validate:"omitempty,oneof=server ca"`
	Cert       string `json:"cert" validate:"omitempty"`
	PrivateKey string `json:"private_key" validate:"omitempty"`
	Remark     string `json:"remark" validate:"omitempty,max=512"`
}

// allowedCertSortFields defines the whitelist of sortable columns for the
// certificate list. It prevents SQL injection through the sort parameter, which
// is otherwise interpolated directly into the ORDER BY clause.
var allowedCertSortFields = map[string]bool{
	"id":         true,
	"name":       true,
	"cert_type":  true,
	"subject":    true,
	"not_before": true,
	"not_after":  true,
	"created_at": true,
	"updated_at": true,
}

// parseLeafCertificate decodes the first CERTIFICATE block from a PEM bundle and
// returns the parsed leaf certificate. It returns an error when no CERTIFICATE
// block is present or the DER body cannot be parsed, which lets the import
// handlers reject malformed material before it is persisted.
func parseLeafCertificate(certPEM string) (*x509.Certificate, error) {
	rest := []byte(certPEM)
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			return x509.ParseCertificate(block.Bytes)
		}
	}
	return nil, errors.New("no CERTIFICATE block found in PEM data")
}

// fillCertMetadata parses the leaf certificate of cert.Cert and copies the
// human-readable identity, serial, SHA-256 fingerprint, and validity window onto
// the model so the management UI can display them without re-parsing the PEM.
func fillCertMetadata(cert *domain.SysCert) error {
	leaf, err := parseLeafCertificate(cert.Cert)
	if err != nil {
		return err
	}
	cert.Subject = leaf.Subject.String()
	cert.Issuer = leaf.Issuer.String()
	cert.Serial = fmt.Sprintf("%X", leaf.SerialNumber)
	sum := sha256.Sum256(leaf.Raw)
	cert.Fingerprint = strings.ToUpper(hex.EncodeToString(sum[:]))
	cert.NotBefore = leaf.NotBefore
	cert.NotAfter = leaf.NotAfter
	return nil
}

// validateKeyPair confirms that keyPEM is the private key matching the leaf
// certificate in certPEM. It is used for server certificates, whose key must be
// usable as a TLS server identity.
func validateKeyPair(certPEM, keyPEM string) error {
	_, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	return err
}

// ListCertificates handles GET /api/v1/system/certificate, returning a paginated
// list of locally managed certificates. It accepts the page and perPage query
// parameters (perPage clamped to 1..100, default 10) and optional name (matched
// case-insensitively) and cert_type filters. The sort and order parameters are
// validated against allowedCertSortFields to keep the ORDER BY clause
// injection-safe. The response body is {"data": []domain.SysCert, "total":
// int64}; the private key is never serialized and HasKey reports its presence.
// Any authenticated operator may call it.
//
// @Summary get the certificate list
// @Tags Certificate
// @Param page query int false "Page number"
// @Param perPage query int false "Items per page"
// @Param sort query string false "Sort field"
// @Param order query string false "Sort direction"
// @Param name query string false "Certificate name"
// @Param cert_type query string false "Certificate type (server|ca)"
// @Success 200 {object} ListResponse
// @Router /api/v1/system/certificate [get]
func ListCertificates(c echo.Context) error {
	db := GetDB(c)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	sortField, order := parseSort(c, allowedCertSortFields, "id", "DESC")

	var total int64
	var certs []domain.SysCert

	query := db.Model(&domain.SysCert{})

	if name := strings.TrimSpace(c.QueryParam("name")); name != "" {
		if strings.EqualFold(db.Name(), "postgres") { //nolint:staticcheck
			query = query.Where("name ILIKE ?", "%"+name+"%")
		} else {
			query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(name)+"%")
		}
	}

	if certType := strings.TrimSpace(c.QueryParam("cert_type")); certType != "" {
		query = query.Where("cert_type = ?", certType)
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	query.Order(sortField + " " + order).Limit(perPage).Offset(offset).Find(&certs)

	for i := range certs {
		certs[i].HasKey = certs[i].PrivateKey != ""
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  certs,
		"total": total,
	})
}

// GetCertificate handles GET /api/v1/system/certificate/:id, returning the single
// certificate with the given numeric id. It responds 400 with code INVALID_ID
// when the path parameter is not an integer and 404 with code NOT_FOUND when no
// such certificate exists. The private key is never serialized. Any authenticated
// operator may call it.
//
// @Summary get certificate detail
// @Tags Certificate
// @Param id path int true "Certificate ID"
// @Success 200 {object} domain.SysCert
// @Router /api/v1/system/certificate/{id} [get]
func GetCertificate(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid certificate ID", nil)
	}

	var cert domain.SysCert
	if err := GetDB(c).First(&cert, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Certificate not found", nil)
	}

	cert.HasKey = cert.PrivateKey != ""
	return ok(c, cert)
}

// CreateCertificate handles POST /api/v1/system/certificate, importing a
// certificate from the JSON body bound to certPayload. The local name must be
// unique; a duplicate is rejected 409 with code NAME_EXISTS. The PEM material is
// parsed to derive the subject, issuer, serial, SHA-256 fingerprint, and validity
// window; malformed material is rejected 400 with code INVALID_CERT. Server
// certificates must include a private key matching the certificate (400 codes
// KEY_REQUIRED / KEY_MISMATCH). On success it returns the persisted
// [domain.SysCert] without the private key. This endpoint requires an admin or
// super operator (see requireAdmin).
//
// @Summary import a certificate
// @Tags Certificate
// @Param certificate body certPayload true "Certificate material"
// @Success 200 {object} domain.SysCert
// @Router /api/v1/system/certificate [post]
func CreateCertificate(c echo.Context) error {
	var payload certPayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request parameters", err.Error())
	}

	if err := c.Validate(&payload); err != nil {
		return handleValidationError(c, err)
	}

	var count int64
	GetDB(c).Model(&domain.SysCert{}).Where("name = ?", payload.Name).Count(&count)
	if count > 0 {
		return fail(c, http.StatusConflict, "NAME_EXISTS", "Certificate name already exists", nil)
	}

	cert := domain.SysCert{
		ID:         common.UUIDint64(),
		Name:       payload.Name,
		CertType:   payload.CertType,
		Cert:       strings.TrimSpace(payload.Cert) + "\n",
		PrivateKey: strings.TrimSpace(payload.PrivateKey),
		Remark:     payload.Remark,
	}
	if cert.PrivateKey != "" {
		cert.PrivateKey += "\n"
	}

	if err := fillCertMetadata(&cert); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_CERT", "Invalid certificate PEM", err.Error())
	}

	if cert.CertType == "server" && cert.PrivateKey == "" {
		return fail(c, http.StatusBadRequest, "KEY_REQUIRED", "Server certificate requires a private key", nil)
	}
	if cert.PrivateKey != "" {
		if err := validateKeyPair(cert.Cert, cert.PrivateKey); err != nil {
			return fail(c, http.StatusBadRequest, "KEY_MISMATCH", "Private key does not match certificate", err.Error())
		}
	}

	if err := GetDB(c).Create(&cert).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to import certificate", err.Error())
	}

	cert.HasKey = cert.PrivateKey != ""
	return ok(c, cert)
}

// UpdateCertificate handles PUT /api/v1/system/certificate/:id, applying a partial
// update to an existing certificate from the JSON body bound to
// certUpdatePayload. Empty cert and private_key fields leave the stored material
// untouched so the local name and remark can be edited without re-uploading PEM.
// When new certificate material is supplied it is re-parsed for metadata, and a
// server certificate's key (new or existing) must match the certificate. A
// changed name must remain unique. It responds 404 with code NOT_FOUND when the
// certificate does not exist and returns the updated [domain.SysCert] on success.
// This endpoint requires an admin or super operator (see requireAdmin).
//
// @Summary update a certificate
// @Tags Certificate
// @Param id path int true "Certificate ID"
// @Param certificate body certUpdatePayload true "Certificate fields"
// @Success 200 {object} domain.SysCert
// @Router /api/v1/system/certificate/{id} [put]
func UpdateCertificate(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid certificate ID", nil)
	}

	var cert domain.SysCert
	if err := GetDB(c).First(&cert, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Certificate not found", nil)
	}

	var payload certUpdatePayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse request parameters", err.Error())
	}
	if err := c.Validate(&payload); err != nil {
		return handleValidationError(c, err)
	}

	if payload.Name != "" && payload.Name != cert.Name {
		var count int64
		GetDB(c).Model(&domain.SysCert{}).Where("name = ? AND id <> ?", payload.Name, id).Count(&count)
		if count > 0 {
			return fail(c, http.StatusConflict, "NAME_EXISTS", "Certificate name already exists", nil)
		}
		cert.Name = payload.Name
	}

	if payload.CertType != "" {
		cert.CertType = payload.CertType
	}

	certReplaced := false
	if strings.TrimSpace(payload.Cert) != "" {
		cert.Cert = strings.TrimSpace(payload.Cert) + "\n"
		certReplaced = true
		if err := fillCertMetadata(&cert); err != nil {
			return fail(c, http.StatusBadRequest, "INVALID_CERT", "Invalid certificate PEM", err.Error())
		}
	}

	if strings.TrimSpace(payload.PrivateKey) != "" {
		cert.PrivateKey = strings.TrimSpace(payload.PrivateKey) + "\n"
	}

	// Re-validate the key/cert pairing whenever either side changed and a key is
	// present, so a renamed pair or a replaced certificate cannot drift apart.
	if cert.PrivateKey != "" && (certReplaced || strings.TrimSpace(payload.PrivateKey) != "") {
		if err := validateKeyPair(cert.Cert, cert.PrivateKey); err != nil {
			return fail(c, http.StatusBadRequest, "KEY_MISMATCH", "Private key does not match certificate", err.Error())
		}
	}
	if cert.CertType == "server" && cert.PrivateKey == "" {
		return fail(c, http.StatusBadRequest, "KEY_REQUIRED", "Server certificate requires a private key", nil)
	}

	cert.Remark = payload.Remark

	if err := GetDB(c).Save(&cert).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update certificate", err.Error())
	}

	cert.HasKey = cert.PrivateKey != ""
	return ok(c, cert)
}

// DeleteCertificate handles DELETE /api/v1/system/certificate/:id, removing the
// certificate with the given numeric id. It responds 400 with code INVALID_ID for
// a non-integer path parameter and returns {"data": {"id": id}} on success. This
// endpoint requires an admin or super operator (see requireAdmin).
//
// @Summary delete a certificate
// @Tags Certificate
// @Param id path int true "Certificate ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/system/certificate/{id} [delete]
func DeleteCertificate(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid certificate ID", nil)
	}

	if err := GetDB(c).Delete(&domain.SysCert{}, id).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete certificate", err.Error())
	}

	return ok(c, map[string]interface{}{"id": strconv.FormatInt(id, 10)})
}

// ExportCertificate handles GET /api/v1/system/certificate/:id/export, streaming
// the certificate as a PEM file attachment named "<name>.pem". By default only the
// public certificate is exported; include_key=true additionally appends the
// private key, which is a sensitive operation recorded to the application log
// together with the requesting operator. A request for the key when none is
// stored is rejected 400 with code NO_KEY. This endpoint requires an admin or
// super operator (see requireAdmin).
//
// @Summary export a certificate
// @Tags Certificate
// @Param id path int true "Certificate ID"
// @Param include_key query bool false "Include the private key in the export"
// @Success 200 {string} string "PEM data"
// @Router /api/v1/system/certificate/{id}/export [get]
func ExportCertificate(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid certificate ID", nil)
	}

	var cert domain.SysCert
	if err := GetDB(c).First(&cert, id).Error; err != nil {
		return fail(c, http.StatusNotFound, "NOT_FOUND", "Certificate not found", nil)
	}

	var buf strings.Builder
	buf.WriteString(strings.TrimRight(cert.Cert, "\n"))
	buf.WriteString("\n")

	includeKey := strings.EqualFold(c.QueryParam("include_key"), "true")
	if includeKey {
		if cert.PrivateKey == "" {
			return fail(c, http.StatusBadRequest, "NO_KEY", "Certificate has no stored private key", nil)
		}
		operator := "unknown"
		if op, oerr := resolveOperatorFromContext(c); oerr == nil && op != nil {
			operator = op.Username
		}
		zap.L().Warn("certificate private key exported",
			zap.String("operator", operator),
			zap.String("certificate", cert.Name),
			zap.Int64("certificate_id", cert.ID),
			zap.String("namespace", "adminapi"))
		buf.WriteString(strings.TrimRight(cert.PrivateKey, "\n"))
		buf.WriteString("\n")
	}

	filename := cert.Name + ".pem"
	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename=\""+filename+"\"")
	return c.Blob(http.StatusOK, "application/x-pem-file", []byte(buf.String()))
}

// registerCertificateRoutes registers certificate management routes.
func registerCertificateRoutes() {
	webserver.ApiGET("/system/certificate", ListCertificates)
	webserver.ApiGET("/system/certificate/:id", GetCertificate)
	webserver.ApiGET("/system/certificate/:id/export", ExportCertificate, requireAdmin())
	webserver.ApiPOST("/system/certificate", CreateCertificate, requireAdmin())
	webserver.ApiPUT("/system/certificate/:id", UpdateCertificate, requireAdmin())
	webserver.ApiDELETE("/system/certificate/:id", DeleteCertificate, requireAdmin())
}
