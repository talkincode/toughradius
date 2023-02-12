package translate

import (
	"net/http"
	"path"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/app"
	"github.com/talkincode/toughradius/common"
	"github.com/talkincode/toughradius/common/web"
	"github.com/talkincode/toughradius/webserver"
)

func InitRouter() {

	webserver.GET("/admin/translate", func(c echo.Context) error {
		return c.Render(http.StatusOK, "translate", nil)
	})

	webserver.GET("/admin/translate/query", func(c echo.Context) error {
		module := c.QueryParam("module")
		keyword := c.QueryParam("keyword")
		lang := c.QueryParam("lang")
		if lang == "" {
			lang = app.GApp().GetTranslateLang()
		}
		data := app.GApp().QueryTranslateTable(lang, module, keyword)
		return c.JSON(http.StatusOK, data)
	})

	webserver.POST("/admin/translate/save", func(c echo.Context) error {
		var op, lang, module, source, result string
		web.NewParamReader(c).
			ReadRequiedString(&op, "webix_operation").
			ReadRequiedString(&lang, "lang").
			ReadRequiedString(&module, "module").
			ReadRequiedString(&source, "source").
			ReadRequiedString(&result, "result")
		switch op {
		case "update":
			app.GApp().TranslateUpdate(lang, module, source, result)
			return c.JSON(http.StatusOK, map[string]interface{}{"status": "updated"})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{})
	})

	webserver.GET("/admin/translate/modules", func(c echo.Context) error {
		modules := make(map[string]int)
		lang := app.GApp().GetTranslateLang()
		data := app.GApp().QueryTranslateTable(lang, "", "")
		for _, v := range data {
			modules[v.Module] = 1
		}
		var result []string
		for k, _ := range modules {
			result = append(result, k)
		}
		return c.JSON(http.StatusOK, result)
	})

	webserver.GET("/admin/translate.js", func(c echo.Context) error {
		lang := app.GApp().GetTranslateLang()
		lfile := path.Join(app.GConfig().System.Workdir, "data", "trans_"+lang+".js")
		return c.File(lfile)
	})

	webserver.GET("/admin/translate/switch/:iszh", func(c echo.Context) error {
		if c.Param("iszh") == "1" {
			app.GApp().SetTranslateLang(app.ZhCN)
		} else {
			app.GApp().SetTranslateLang(app.EnUS)
		}
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.GET("/admin/translate/flush", func(c echo.Context) error {
		app.GApp().RenderTranslateFiles()
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.POST("/admin/translate/delete", func(c echo.Context) error {
		var items []app.TransTable
		common.Must(c.Bind(&items))
		app.GApp().RemoveTranslateItems(items)
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.POST("/admin/translate/patch", func(c echo.Context) error {
		type PatchForm struct {
			Module string `json:"module" form:"module"`
			Key    string `json:"key" form:"key"`
			Value  string `json:"value" form:"value"`
		}
		var form PatchForm
		err := c.Bind(&form)
		if err != nil {
			return c.JSON(http.StatusBadRequest, web.RestError(err.Error()))
		}

		app.GApp().Translate(app.ZhCN, form.Module, form.Key, form.Value)
		app.GApp().Translate(app.EnUS, form.Module, form.Key, form.Value)
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

}
