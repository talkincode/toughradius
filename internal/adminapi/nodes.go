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

// 网络节点请求结构
type nodePayload struct {
	Name   string `json:"name"`
	Tags   string `json:"tags"`
	Remark string `json:"remark"`
}

// 注册网络节点路由
func registerNodesRoutes() {
	webserver.ApiGET("/network/nodes", listNodes)
	webserver.ApiGET("/network/nodes/:id", getNode)
	webserver.ApiPOST("/network/nodes", createNode)
	webserver.ApiPUT("/network/nodes/:id", updateNode)
	webserver.ApiDELETE("/network/nodes/:id", deleteNode)
}

// 获取网络节点列表
func listNodes(c echo.Context) error {
	page, pageSize := parsePagination(c)

	base := app.GDB().Model(&domain.NetNode{})
	base = applyNodeFilters(base, c)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "查询网络节点失败", err.Error())
	}

	var nodes []domain.NetNode
	if err := base.
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&nodes).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "查询网络节点失败", err.Error())
	}

	return paged(c, nodes, total, page, pageSize)
}

// 获取单个网络节点
func getNode(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "无效的节点 ID", nil)
	}

	var node domain.NetNode
	if err := app.GDB().Where("id = ?", id).First(&node).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusNotFound, "NODE_NOT_FOUND", "节点不存在", nil)
	} else if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "查询网络节点失败", err.Error())
	}

	return ok(c, node)
}

// 创建网络节点
func createNode(c echo.Context) error {
	var payload nodePayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解析节点参数", nil)
	}

	payload.Name = strings.TrimSpace(payload.Name)
	if payload.Name == "" {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "name 不能为空", nil)
	}

	// 检查节点名称是否已存在
	var exists int64
	app.GDB().Model(&domain.NetNode{}).Where("name = ?", payload.Name).Count(&exists)
	if exists > 0 {
		return fail(c, http.StatusConflict, "NODE_EXISTS", "节点名称已存在", nil)
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
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "创建网络节点失败", err.Error())
	}

	return ok(c, node)
}

// 更新网络节点
func updateNode(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "无效的节点 ID", nil)
	}

	var payload nodePayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解析节点参数", nil)
	}

	var node domain.NetNode
	if err := app.GDB().Where("id = ?", id).First(&node).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusNotFound, "NODE_NOT_FOUND", "节点不存在", nil)
	} else if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "查询网络节点失败", err.Error())
	}

	// 更新字段
	if payload.Name != "" {
		node.Name = strings.TrimSpace(payload.Name)
	}
	if payload.Tags != "" {
		node.Tags = payload.Tags
	}
	if payload.Remark != "" {
		node.Remark = payload.Remark
	}
	node.UpdatedAt = time.Now()

	if err := app.GDB().Save(&node).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "更新网络节点失败", err.Error())
	}

	return ok(c, node)
}

// 删除网络节点
func deleteNode(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "无效的节点 ID", nil)
	}

	// 检查是否有 NAS 设备关联此节点
	var nasCount int64
	app.GDB().Model(&domain.NetNas{}).Where("node_id = ?", id).Count(&nasCount)
	if nasCount > 0 {
		return fail(c, http.StatusConflict, "NODE_IN_USE", "该节点下还有 NAS 设备，无法删除", nil)
	}

	if err := app.GDB().Where("id = ?", id).Delete(&domain.NetNode{}).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "删除网络节点失败", err.Error())
	}

	return ok(c, map[string]interface{}{
		"id": id,
	})
}

// 筛选条件
func applyNodeFilters(db *gorm.DB, c echo.Context) *gorm.DB {
	if name := strings.TrimSpace(c.QueryParam("name")); name != "" {
		db = db.Where("name ILIKE ?", "%"+name+"%")
	}

	return db
}
