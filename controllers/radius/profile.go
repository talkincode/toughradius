package radius

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
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// InitProfileRouter RADIUS 资费策略的增删改查
func InitProfileRouter() {

	// 页面展示 assets/templates/radius_profile.html
	webserver.GET("/admin/radius/profile", func(c echo.Context) error {
		return c.Render(http.StatusOK, "radius_profile", map[string]interface{}{
			"oprlevel": webserver.GetCurrUserlevel(c),
		})
	})

	webserver.GET("/admin/radius/profile/options", func(c echo.Context) error {
		var data []models.RadiusProfile
		common.Must(app.GDB().Order("id").Find(&data).Error)
		var options = make([]web.JsonOptions, 0)
		for _, d := range data {
			options = append(options, web.JsonOptions{
				Id:    cast.ToString(d.ID),
				Value: d.Name,
			})
		}
		return c.JSON(http.StatusOK, options)
	})

	webserver.GET("/admin/radius/profile/query", func(c echo.Context) error {
		var count, start int
		web.NewParamReader(c).
			ReadInt(&start, "start", 0).
			ReadInt(&count, "count", 40)
		var data []models.RadiusProfile
		getQuery := func() *gorm.DB {
			query := app.GDB().Model(&models.RadiusProfile{})
			if len(web.ParseSortMap(c)) == 0 {
				query = query.Order("updated_at desc")
			} else {
				for name, stype := range web.ParseSortMap(c) {
					query = query.Order(fmt.Sprintf("%s %s", name, stype))
				}
			}

			for name, value := range web.ParseFilterMap(c) {
				if common.InSlice(name, []string{"pnode_id"}) {
					query = query.Where(fmt.Sprintf("%s = ?", name), value)
				} else {
					query = query.Where(fmt.Sprintf("%s like ?", name), "%"+value+"%")
				}
			}
			keyword := c.QueryParam("keyword")
			if keyword != "" {
				query = query.Where("username like ?", "%"+keyword+"%").
					Or("remark like ?", "%"+keyword+"%").
					Or("addr_pool like ?", "%"+keyword+"%").
					Or("tags like ?", "%"+keyword+"%")
			}
			return query
		}
		var total int64
		common.Must(getQuery().Count(&total).Error)

		query := getQuery().Offset(start).Limit(count)
		if query.Find(&data).Error != nil {
			return c.JSON(http.StatusOK, common.EmptyList)
		}
		return c.JSON(http.StatusOK, &web.PageResult{TotalCount: total, Pos: int64(start), Data: data})
	})

	webserver.GET("/admin/radius/profile/get", func(c echo.Context) error {
		var id string
		web.NewParamReader(c).
			ReadRequiedString(&id, "id")
		var data models.RadiusProfile
		common.Must(app.GDB().Where("id=?", id).First(&data).Error)
		return c.JSON(http.StatusOK, data)
	})

	webserver.POST("/admin/radius/profile/add", func(c echo.Context) error {
		form := new(models.RadiusProfile)
		form.ID = common.UUIDint64()
		common.Must(c.Bind(form))
		common.CheckEmpty("name", form.Name)
		common.Must(app.GDB().Create(form).Error)
		webserver.PubOpLog(c, fmt.Sprintf("创建RADIUS策略：%v", form))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.POST("/admin/radius/profile/update", func(c echo.Context) error {
		form := new(models.RadiusProfile)
		common.Must(c.Bind(form))
		common.CheckEmpty("name", form.Name)
		common.Must(app.GDB().Save(form).Error)
		common.Must(app.GDB().Model(&models.RadiusUser{}).Where("profile_id=?", form.ID).Updates(map[string]interface{}{
			"addr_pool":  form.AddrPool,
			"active_num": form.ActiveNum,
			"up_rate":    form.UpRate,
			"down_rate":  form.DownRate,
		}).Error)
		webserver.PubOpLog(c, fmt.Sprintf("更新RADIUS策略：%v", form))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.GET("/admin/radius/profile/delete", func(c echo.Context) error {
		ids := c.QueryParam("ids")
		common.Must(app.GDB().Delete(models.RadiusProfile{}, strings.Split(ids, ",")).Error)
		webserver.PubOpLog(c, fmt.Sprintf("删除RADIUS策略：%s", ids))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.POST("/admin/radius/profile/import", func(c echo.Context) error {
		datas, err := webserver.ImportData(c, "radius_profile")
		common.Must(err)
		common.Must(app.GDB().Model(models.RadiusProfile{}).Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(datas).Error)
		return c.JSON(http.StatusOK, web.RestSucc("Success"))
	})

	webserver.GET("/admin/radius/profile/export", func(c echo.Context) error {
		var data []models.RadiusProfile
		common.Must(app.GDB().Find(&data).Error)
		return webserver.ExportCsv(c, data, "radius_profile")
	})
}
