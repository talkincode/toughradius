package factoryreset

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
	"github.com/talkincode/toughradius/app"
	"github.com/talkincode/toughradius/common"
	"github.com/talkincode/toughradius/common/web"
	"github.com/talkincode/toughradius/models"
	"github.com/talkincode/toughradius/webserver"
)

func InitRouter() {

	webserver.GET("/admin/cwmp/factoryreset", func(c echo.Context) error {
		return c.Render(http.StatusOK, "cwmp_factoryreset", nil)
	})

	webserver.GET("/admin/cwmp/factoryreset/options", func(c echo.Context) error {
		var data []models.CwmpFactoryReset
		common.Must(app.GDB().Find(&data).Error)
		var opts = make([]web.JsonOptions, 0)
		for _, d := range data {
			opts = append(opts, web.JsonOptions{
				Id:    cast.ToString(d.ID),
				Value: d.Name + "(" + d.SoftwareVersion + ")",
			})
		}
		return c.JSON(http.StatusOK, opts)
	})

	webserver.GET("/admin/cwmp/factoryreset/query", func(c echo.Context) error {
		prequery := web.NewPreQuery(c).
			DefaultOrderBy("updated_at desc").
			KeyFields("oid", "name", "software_version", "product_class", "oui")
		result, err := web.QueryPageResult[models.CwmpFactoryReset](c, app.GDB(), prequery)
		if err != nil {
			return c.JSON(http.StatusOK, common.EmptyList)
		}
		return c.JSON(http.StatusOK, result)
	})

	webserver.POST("/admin/cwmp/factoryreset/add", func(c echo.Context) error {
		form := new(models.CwmpFactoryReset)
		form.ID = common.UUIDint64()
		common.Must(c.Bind(form))
		common.MustNotEmpty("Oid", form.Oid)
		common.Must(app.GDB().Create(form).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Create factory reset information：%v", form))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.POST("/admin/cwmp/factoryreset/update", func(c echo.Context) error {
		form := new(models.CwmpFactoryReset)
		common.Must(c.Bind(form))
		common.MustNotEmpty("Oid", form.Oid)
		common.Must(app.GDB().Save(form).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Modify factory reset information：%v", form))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.GET("/admin/cwmp/factoryreset/delete", func(c echo.Context) error {
		ids := c.QueryParam("ids")
		common.Must(app.GDB().Delete(models.CwmpFactoryReset{}, strings.Split(ids, ",")).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Delete factory reset information：%s", ids))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

}
