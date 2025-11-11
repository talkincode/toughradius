package adminapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// nodePayload defines the network node request structure
type nodePayload struct {
	Name   string `json:"name" validate:"required,min=1,max=100"`
	Tags   string `json:"tags" validate:"omitempty,max=200"`
	Remark string `json:"remark" validate:"omitempty,max=500"`
}

// registerNodesRoutes registers network node routes
func registerNodesRoutes() {
	webserver.ApiGET("/network/nodes", listNodes)
	webserver.ApiGET("/network/nodes/:id", getNode)
	webserver.ApiPOST("/network/nodes", createNode)
	webserver.ApiPUT("/network/nodes/:id", updateNode)
	webserver.ApiDELETE("/network/nodes/:id", deleteNode)
}

// listNodes retrieves the network node list
func listNodes(c echo.Context) error {
	page, pageSize := parsePagination(c)

	base := app.GDB().Model(&domain.NetNode{})
	base = applyNodeFilters(base, c)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query network nodes", err.Error())
	}

	var nodes []domain.NetNode
		if err := base.
			Order("id DESC").
			Offset((page - 1) * pageSize).
			Limit(pageSize).
			Find(&nodes).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query network nodes", err.Error())
	}

	return paged(c, nodes, total, page, pageSize)
}

// getNode retrieves a single network node
func getNode(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid node ID", nil)
	}

	var node domain.NetNode
	if err := app.GDB().Where("id = ?", id).First(&node).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusNotFound, "NODE_NOT_FOUND", "Node not found", nil)
	} else if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query network nodes", err.Error())
	}

	return ok(c, node)
}

// createNode creates a network node
func createNode(c echo.Context) error {
	var payload nodePayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse node parameters", nil)
	}

	// Validate the request payload
	if err := c.Validate(&payload); err != nil {
		return err
	}

	payload.Name = strings.TrimSpace(payload.Name)

	// Check whether another node already uses the same name
	var exists int64
	app.GDB().Model(&domain.NetNode{}).Where("name = ?", payload.Name).Count(&exists)
	if exists > 0 {
		return fail(c, http.StatusConflict, "NODE_EXISTS", "Node name already exists", nil)
	}

	node := domain.NetNode{
		ID:        common.UUIDint64(),
		Name:      payload.Name,
		Tags:      payload.Tags,
		Remark:    payload.Remark,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := app.GDB().Create(&node).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to create network node", err.Error())
	}

	return ok(c, node)
}

// updateNode updates a network node
func updateNode(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid node ID", nil)
	}

	var payload nodePayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse node parameters", nil)
	}

	// Validate the request payload
	if err := c.Validate(&payload); err != nil {
		return err
	}

	var node domain.NetNode
	if err := app.GDB().Where("id = ?", id).First(&node).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusNotFound, "NODE_NOT_FOUND", "Node not found", nil)
	} else if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query network nodes", err.Error())
	}

	// If the name is changed, check whether another node already uses it
	if payload.Name != "" && payload.Name != node.Name {
		payload.Name = strings.TrimSpace(payload.Name)
		var exists int64
		app.GDB().Model(&domain.NetNode{}).Where("name = ? AND id != ?", payload.Name, id).Count(&exists)
		if exists > 0 {
			return fail(c, http.StatusConflict, "NODE_EXISTS", "Node name already exists", nil)
		}
		node.Name = payload.Name
	}

	// Update other fields
	if payload.Tags != "" {
		node.Tags = payload.Tags
	}
	if payload.Remark != "" {
		node.Remark = payload.Remark
	}
	node.UpdatedAt = time.Now()

	if err := app.GDB().Save(&node).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update network node", err.Error())
	}

	return ok(c, node)
}

// deleteNode deletes a network node
func deleteNode(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid node ID", nil)
	}

	// Check whether NAS devices are associated with this node
	var nasCount int64
	app.GDB().Model(&domain.NetNas{}).Where("node_id = ?", id).Count(&nasCount)
	if nasCount > 0 {
		return fail(c, http.StatusConflict, "NODE_IN_USE", "This node still has NAS devices and cannot be deleted", nil)
	}

	if err := app.GDB().Where("id = ?", id).Delete(&domain.NetNode{}).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to delete network node", err.Error())
	}

	return ok(c, map[string]interface{}{
		"id": id,
	})
}

// Filter conditions
func applyNodeFilters(db *gorm.DB, c echo.Context) *gorm.DB {
	if name := strings.TrimSpace(c.QueryParam("name")); name != "" {
		db = db.Where("name ILIKE ?", "%"+name+"%")
	}

	return db
}
