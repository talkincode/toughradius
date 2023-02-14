package index

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/app"
	"github.com/talkincode/toughradius/assets"
	"github.com/talkincode/toughradius/common"
	"github.com/talkincode/toughradius/common/web"
	"github.com/talkincode/toughradius/models"
	"github.com/talkincode/toughradius/webserver"
)

var pushers = []string{
	"/static/webix/webix.min.js",
	"/static/myskin/webix.min.css",
	"/static/views/widgets.min.js",
	"/static/views/wxui.min.js",
	"/static/views/codemirror.js",
	"/static/views/codemirror.css",
	"/static/myskin/materialdesignicons.min.css",
	"static/myskin/fonts/Roboto-Medium-webfont.woff2",
	"/static/myskin/fonts/materialdesignicons-webfont.woff2?v=7.1.96",
	"/static/echarts/echarts.min.js",
}

func InitRouter() {

	// 系统首页
	webserver.GET("/", func(c echo.Context) (err error) {
		// pusher, ok := c.Response().Writer.(http.Pusher)
		// if ok {
		// 	for _, res := range pushers {
		// 		if err = pusher.Push(res, nil); err != nil {
		// 			continue
		// 		}
		// 	}
		// }
		sess, _ := session.Get(webserver.UserSession, c)
		username := sess.Values[webserver.UserSessionName]
		if username == nil || username == "" {
			return c.Redirect(http.StatusTemporaryRedirect, "/login?errmsg=User not logged in or login expired")
		}
		return c.Render(http.StatusOK, "index", map[string]interface{}{})
	})

	type menus struct {
		Id    string `json:"id"`
		Value string `json:"value"`
		Icon  string `json:"icon"`
		Url   string `json:"url,omitempty"`
		Data  []*struct {
			Id    string `json:"id"`
			Value string `json:"value"`
			Icon  string `json:"icon"`
			Url   string `json:"url"`
		} `json:"data,omitempty"`
	}

	// 菜单数据
	webserver.GET("/admin/menu.json", func(c echo.Context) error {
		var menudata []byte
		sess, _ := session.Get(webserver.UserSession, c)
		switch sess.Values[webserver.UserSessionLevel] {
		case "super":
			menudata = assets.AdminMenudata
		case "opr":
			menudata = assets.OprMenudata
		default:
			menudata = []byte("[]")
		}
		var result []*menus
		if err := json.Unmarshal(menudata, &result); err != nil {
			return c.JSONBlob(http.StatusOK, menudata)
		}
		lang := app.GApp().GetTranslateLang()
		for _, m := range result {
			m.Value = app.GApp().Translate(lang, "menus", m.Value, m.Value)
			if m.Data != nil {
				for _, d := range m.Data {
					d.Value = app.GApp().Translate(lang, "menus", d.Value, d.Value)
				}
			}
		}

		return c.JSON(http.StatusOK, result)
	})

	// 登录页面
	webserver.GET("/login", func(c echo.Context) error {
		errmsg := c.QueryParam("errmsg")
		return c.Render(http.StatusOK, "login", map[string]interface{}{
			"errmsg":    errmsg,
			"LoginLogo": "/static/images/login-logo.png",
		})
	})

	webserver.GET("/admin/theme/switch/:isdark", func(c echo.Context) error {
		isdark := c.Param("isdark")
		if isdark == "1" {
			app.GApp().SetSystemTheme("dark")
		} else {
			app.GApp().SetSystemTheme("light")
		}
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	// 登出页面
	webserver.GET("/logout", func(c echo.Context) error {
		sess, _ := session.Get(webserver.UserSession, c)
		sess.Values = make(map[interface{}]interface{})
		_ = sess.Save(c.Request(), c.Response())
		return c.Redirect(http.StatusMovedPermanently, "/login")
	})

	// 登录提交
	webserver.POST("/login", func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")
		if username == "" || password == "" {
			return c.Redirect(http.StatusMovedPermanently, "/login?errmsg=Username and password cannot be empty")
		}
		var user models.SysOpr
		err := app.GDB().Where("username=?", username).First(&user).Error
		if err != nil {
			return c.Redirect(http.StatusMovedPermanently, "/login?errmsg=User does not exist")
		}

		if common.Sha256HashWithSalt(password, common.SecretSalt) != user.Password {
			return c.Redirect(http.StatusMovedPermanently, "/login?errmsg=wrong password")
		}

		sess, _ := session.Get(webserver.UserSession, c)
		sess.Values[webserver.UserSessionName] = username
		sess.Values[webserver.UserSessionLevel] = user.Level
		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			return echo.NewHTTPError(http.StatusMovedPermanently, err.Error())
		}
		return c.Redirect(http.StatusMovedPermanently, "/")
	})

	type AuthForm struct {
		Username string `json:"username" form:"username"`
		Password string `json:"password" form:"password"`
	}

	webserver.POST("/token", func(c echo.Context) error {
		form := new(AuthForm)
		common.Must(c.Bind(form))
		common.MustNotEmpty("username", form.Username)
		common.MustNotEmpty("password", form.Password)
		var user models.SysOpr
		common.Must(app.GDB().Where("username=?", form.Username).First(&user).Error)
		if common.Sha256HashWithSalt(form.Password, common.SecretSalt) != user.Password {
			return echo.NewHTTPError(http.StatusForbidden)
		}

		t, err := web.CreateToken(app.GConfig().Web.Secret, user.Username, user.Level, time.Hour*24*365)
		common.Must(err)
		return c.JSON(http.StatusOK, web.RestResult(map[string]string{
			"token": t,
		}))
	})
}
