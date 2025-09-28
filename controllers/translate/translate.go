package translate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v8/app"
	"github.com/talkincode/toughradius/v8/assets"
	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/common/web"
	"github.com/talkincode/toughradius/v8/webserver"
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

	webserver.GET("/admin/translate/switch/:lang", func(c echo.Context) error {
		lang := c.Param("lang")
		// Handle legacy binary switch for backward compatibility
		if lang == "1" {
			lang = app.ZhCN
		} else if lang == "0" {
			lang = app.EnUS
		}
		// Validate the language code
		if lang == app.ZhCN || lang == app.EnUS || lang == app.FrFR {
			app.GApp().SetTranslateLang(lang)
		}
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.GET("/admin/translate/flush", func(c echo.Context) error {
		app.GApp().RenderTranslateFiles()
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.GET("/admin/translate/export", func(c echo.Context) error {
		lang := c.QueryParam("lang")
		result := app.GApp().ListTranslateTable(lang)
		bs, _ := json.MarshalIndent(result, "", "  ")
		c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s.json", lang))
		return c.JSONBlob(200, bs)
	})

	webserver.GET("/admin/translate/transall", func(c echo.Context) error {
		lang := c.QueryParam("lang")
		result := app.GApp().ListTranslateTable(lang)
		for _, t := range result {
			if t.Lang == app.EnUS {
				continue
			}
			trs, err := common.Translate(t.Source, app.EnUS, lang)
			if err != nil {
				time.Sleep(time.Millisecond * 200)
			}
			t.Result = trs
			app.GApp().TranslateUpdate(lang, t.Module, t.Source, trs)
		}
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.GET("/admin/translate/transone", func(c echo.Context) error {
		var lang, module, source string
		err := web.NewParamReader(c).
			ReadRequiedString(&lang, "lang").
			ReadRequiedString(&source, "source").LastError
		if err != nil {
			return c.JSON(http.StatusBadRequest, web.RestError(err.Error()))
		}
		trs, err := common.Translate(source, app.EnUS, lang)
		if err != nil {
			return c.JSON(http.StatusBadRequest, web.RestError(err.Error()))
		}
		app.GApp().TranslateUpdate(lang, module, source, trs)
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.GET("/admin/translate/init", func(c echo.Context) error {
		var items []app.TransTable
		err := json.Unmarshal(assets.I18nZhCNResources, &items)
		if err != nil {
			return c.JSON(http.StatusBadRequest, web.RestError(err.Error()))
		}
		for _, t := range items {
			app.GApp().TranslateUpdate(t.Lang, t.Module, t.Source, t.Result)
		}
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
		app.GApp().Translate(app.FrFR, form.Module, form.Key, form.Value)
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

}
