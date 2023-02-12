package tpl

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"
	"text/template"

	"github.com/labstack/echo/v4"
)

type CommonTemplate struct {
	Templates *template.Template
	AssetsFs  embed.FS
}

func NewCommonTemplate(fs embed.FS, dirs []string, funcMap map[string]interface{}) *CommonTemplate {
	var templates = template.New("GlobalTemplate").Funcs(funcMap)
	var ct = &CommonTemplate{Templates: templates, AssetsFs: fs}
	for _, d := range dirs {
		ct.parseDir(d)
	}
	return ct
}

func (ct *CommonTemplate) parseDir(dir string) {
	fss, _ := ct.AssetsFs.ReadDir(dir)
	for _, item := range fss {
		if item.IsDir() {
			continue
		}
		c, err := ct.AssetsFs.ReadFile(path.Join(dir, item.Name()))
		if err == nil {
			ct.parseItem(item, c, ct.Templates)
		}
	}
}

func (ct *CommonTemplate) parseItem(item fs.DirEntry, c []byte, templates *template.Template) {
	name := strings.TrimSuffix(item.Name(), path.Ext(item.Name()))
	if templates.Lookup(name) != nil {
		return
	}
	tplstr := fmt.Sprintf(`{{define "%s"}}%s{{end}}`, name, c)
	ct.Templates = template.Must(templates.Parse(tplstr))
}

func (t *CommonTemplate) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.Templates.ExecuteTemplate(w, name, data)
}
