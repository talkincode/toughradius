package apitoken

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v8/app"
	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/common/web"
	"github.com/talkincode/toughradius/v8/models"
	"github.com/talkincode/toughradius/v8/webserver"
)

func InitRouter() {
	webserver.GET("/admin/apitoken", func(c echo.Context) error {
		return c.Render(http.StatusOK, "apitoken", nil)
	})

	webserver.GET("/admin/apitoken/query", func(c echo.Context) error {
		var count, start int
		web.NewParamReader(c).
			ReadInt(&start, "start", 0).
			ReadInt(&count, "count", 40)
		var data []models.SysApiToken
		prequery := web.NewPreQuery(c).
			DefaultOrderBy("created_at desc").
			KeyFields("uid", "name", "remark")

		var total int64
		common.Must(prequery.Query(app.GDB().Model(&models.SysApiToken{})).Count(&total).Error)

		query := prequery.Query(app.GDB().Model(&models.SysApiToken{})).Offset(start).Limit(count)
		if query.Find(&data).Error != nil {
			return c.JSON(http.StatusOK, common.EmptyList)
		}
		return c.JSON(http.StatusOK, &web.PageResult{TotalCount: total, Pos: int64(start), Data: data})
	})

	webserver.GET("/admin/apitoken/get", func(c echo.Context) error {
		var id string
		web.NewParamReader(c).
			ReadRequiedString(&id, "id")
		var data models.SysApiToken
		common.Must(app.GDB().Where("id=?", id).First(&data).Error)
		return c.JSON(http.StatusOK, data)
	})

	type apitokenReq struct {
		Name   string `json:"name" form:"name"`
		Remark string `json:"remark" form:"remark"`
		Expire string `json:"expire" form:"expire"`
	}

	webserver.POST("/admin/apitoken/add", func(c echo.Context) error {
		form := new(apitokenReq)
		common.Must(c.Bind(form))
		common.CheckEmpty("name", form.Name)
		common.CheckEmpty("expire", form.Expire)
		token := new(models.SysApiToken)
		token.ID = common.UUID()
		token.Uid = common.UUID()
		token.Name = form.Name
		token.Level = "api"
		var err error
		token.ExpireTime, err = time.Parse("2006-01-02 15:04:05", form.Expire[0:10]+" 23:59:59")
		if err != nil {
			return c.JSON(200, web.RestError(fmt.Sprintf("expire %s format error", form.Expire)))
		}
		token.Token, err = web.CreateToken(app.GConfig().Web.Secret, token.Uid, token.Level, token.ExpireTime.Sub(time.Now()))
		if err != nil {
			return c.JSON(200, web.RestError(fmt.Sprintf("create token error %s ", err.Error())))
		}
		token.Remark = form.Remark
		token.CreatedAt = time.Now()

		common.Must(app.GDB().Create(token).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Create API Token information：%v", form))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.GET("/admin/apitoken/delete", func(c echo.Context) error {
		ids := c.QueryParam("ids")
		common.Must(app.GDB().Delete(models.SysApiToken{}, strings.Split(ids, ",")).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Delete API Token information：%s", ids))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

}
