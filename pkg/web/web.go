/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package web

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"gorm.io/gorm"
)

type DateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

func (d DateRange) ParseStart() (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", d.Start)
}

func (d DateRange) ParseEnd() (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", d.End)
}

// WEB 参数
type WebForm struct {
	FormItem interface{}
	Posts    url.Values        `json:"-" form:"-" query:"-"`
	Gets     url.Values        `json:"-" form:"-" query:"-"`
	Params   map[string]string `json:"-" form:"-" query:"-"`
}

func EmptyWebForm() *WebForm {
	v := &WebForm{}
	v.Params = make(map[string]string, 0)
	v.Posts = make(url.Values, 0)
	v.Gets = make(url.Values, 0)
	return v
}

func NewWebForm(c echo.Context) *WebForm {
	v := &WebForm{}
	v.Params = make(map[string]string)
	v.Posts, _ = c.FormParams()
	v.Gets = c.QueryParams()
	for _, p := range c.ParamNames() {
		v.Params[p] = c.Param(p)
	}
	return v
}

func (f *WebForm) Set(name string, value string) {
	f.Gets.Set(name, value)
}

func (f *WebForm) Param(name string) string {
	return f.Param(name)
}

func (f *WebForm) Param2(name string, defval string) string {
	if val, ok := f.Params[name]; ok {
		return val
	}
	return defval
}

func (f *WebForm) GetDateRange(name string) (DateRange, error) {
	var dr = DateRange{Start: "", End: ""}
	val := f.GetVal(name)
	if val == "" {
		return dr, nil
	}
	err := json.Unmarshal([]byte(val), &dr)
	if err != nil {
		return dr, err
	}
	return dr, nil
}

func (f *WebForm) GetVal(name string) string {
	val := f.Posts.Get(name)
	if val != "" {
		return val
	}
	val = f.Gets.Get(name)
	if val != "" {
		return val
	}
	return ""
}

func (f *WebForm) GetMustVal(name string) (string, error) {
	val := f.Posts.Get(name)
	if val != "" {
		return val, nil
	}
	val = f.Gets.Get(name)
	if val != "" {
		return val, nil
	}
	return "", errors.New(name + " 不能为空")
}

func (f *WebForm) GetVal2(name string, defval string) string {
	val := f.Posts.Get(name)
	if val != "" {
		return val
	}
	val = f.Gets.Get(name)
	if val != "" {
		return val
	}
	return defval
}

func (f *WebForm) GetIntVal(name string, defval int) int {
	val := f.GetVal(name)
	if val == "" {
		return defval
	}
	v, _ := strconv.Atoi(val)
	return v
}

func (f *WebForm) GetInt64Val(name string, defval int64) int64 {
	val := f.GetVal(name)
	if val == "" {
		return defval
	}
	v, _ := strconv.ParseInt(val, 10, 64)
	return v
}

// ParseTimeDesc 解析时间描述
// now-1hour 最近1小时
// now-1min 最近1分钟
// now-1day 最近1天
func (f *WebForm) ParseTimeDesc(timestr string, defval string) string {
	val := f.GetVal(timestr)
	if val == "" {
		val = defval
	}
	switch {
	case strings.HasPrefix(timestr, "now-") && strings.HasSuffix(timestr, "hour"):
		v := cast.ToInt(timestr[4 : len(timestr)-4])
		return time.Now().Add(time.Hour * time.Duration(v*-1)).Format(time.RFC3339)
	case strings.HasPrefix(timestr, "now-") && strings.HasSuffix(timestr, "min"):
		v := cast.ToInt(timestr[4 : len(timestr)-4])
		return time.Now().Add(time.Minute * time.Duration(v*-1)).Format(time.RFC3339)
	case strings.HasPrefix(timestr, "now-") && strings.HasSuffix(timestr, "sec"):
		v := cast.ToInt(timestr[4 : len(timestr)-4])
		return time.Now().Add(time.Second * time.Duration(v*-1)).Format(time.RFC3339)
	case strings.HasPrefix(timestr, "now-") && strings.HasSuffix(timestr, "day"):
		v := cast.ToInt(timestr[4 : len(timestr)-4])
		return time.Now().Add(time.Hour * 24 * time.Duration(v*-1)).Format(time.RFC3339)
	case timestr == "now":
		return time.Now().Format(time.RFC3339)
	default:
		return time.Now().Format(time.RFC3339)
	}
}

type PageResult struct {
	TotalCount int64       `json:"total_count,omitempty"`
	Pos        int64       `json:"pos"`
	Data       interface{} `json:"data"`
}

type JsonOptions struct {
	Id    string `json:"id"`
	Value string `json:"value"`
}

// 参数快速读取,避免多次错误处理
type ParamReader struct {
	form      *WebForm
	LastError error
}

func NewParamReader(c echo.Context) *ParamReader {
	return NewWebForm(c).newParamReader()
}

func (f *WebForm) newParamReader() *ParamReader {
	return &ParamReader{form: f}
}

func (sr *ParamReader) ReadRequiedString(ref *string, name string) *ParamReader {
	if sr.LastError != nil {
		return sr
	}
	val, err := sr.form.GetMustVal(name)
	if err != nil {
		sr.LastError = err
	} else {
		*ref = val
	}
	return sr
}

func (sr *ParamReader) ReadString(ref *string, name string) *ParamReader {
	val, _ := sr.form.GetMustVal(name)
	*ref = val
	return sr
}

func (sr *ParamReader) ReadStringWithDefault(ref *string, name string, defval string) *ParamReader {
	val, err := sr.form.GetMustVal(name)
	if err != nil {
		val = defval
	}
	*ref = val
	return sr
}

func (sr *ParamReader) ReadInt64(ref *int64, name string, defval int64) *ParamReader {
	*ref = sr.form.GetInt64Val(name, defval)
	return sr
}

func (sr *ParamReader) ReadInt(ref *int, name string, defval int) *ParamReader {
	*ref = sr.form.GetIntVal(name, defval)
	return sr
}

type WebRestResult struct {
	Code    int         `json:"code"`
	Msgtype string      `json:"msgtype"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data"`
}

func RestResult(data interface{}) *WebRestResult {
	return &WebRestResult{
		Code:    0,
		Msgtype: "info",
		Msg:     "success",
		Data:    data,
	}
}
func RestSucc(msg string) *WebRestResult {
	return &WebRestResult{
		Code:    0,
		Msgtype: "info",
		Msg:     msg,
	}
}

func RestError(msg string) *WebRestResult {
	return &WebRestResult{
		Code:    1,
		Msgtype: "error",
		Msg:     msg,
	}
}

func ReadImportExcelData(src io.Reader, sheet string) ([]map[string]interface{}, error) {
	f, err := excelize.OpenReader(src)
	if err != nil {
		return nil, err
	}
	// 获取 Sheet1 上所有单元格
	rows := f.GetRows(sheet)
	head := make(map[int]string)
	var data []map[string]interface{}
	for i, row := range rows {
		item := make(map[string]interface{})
		for k, colCell := range row {
			if i == 0 {
				head[k] = colCell
			} else {
				item[head[k]] = colCell
			}
		}
		if i == 0 {
			continue
		}
		data = append(data, item)
	}

	return data, nil
}

func ReadImportJsonData(src io.Reader) ([]map[string]interface{}, error) {
	buff := bufio.NewReader(src)
	var items []map[string]interface{}
	for {
		data, err := buff.ReadBytes('\n')
		switch {
		case err == io.EOF:
			log.Println("Reached EOF - close this connection.\n  ---")
			break
		case err != nil:
			log.Printf("\nError reading command. %s\n", err)
			break
		}
		item := make(map[string]interface{})
		err2 := common.JsonUnmarshal(data, &item)
		if err2 != nil {
			break
		}
		items = append(items, item)
	}
	return items, nil
}

func ReadImportCsvData(src io.Reader) ([]map[string]interface{}, error) {
	csvReader := csv.NewReader(src)
	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0)
	headers := make(map[int]string, 0)
	for i, row := range rows {
		if i == 0 {
			for k, c := range row {
				headers[k] = string(common.TrimBytes([]byte(strings.TrimSpace(c))))
			}
		} else {
			item := make(map[string]interface{})
			for k, c := range row {
				item[headers[k]] = strings.TrimSpace(c)
			}
			result = append(result, item)
		}
	}
	return result, nil

}

func ParseSortMap(c echo.Context) map[string]string {
	sortmap := make(map[string]string)
	for k, vs := range c.QueryParams() {
		switch {
		case strings.HasPrefix(k, "sort[") && vs[0] != "":
			if common.InSlice(vs[0], []string{"asc", "desc"}) {
				// 排序参数
				sortmap[k[5:len(k)-1]] = vs[0]
			}
		}
	}
	return sortmap
}

func ParseFilterMap(c echo.Context) map[string]string {
	filtermap := make(map[string]string)
	for k, vs := range c.QueryParams() {
		switch {
		case strings.HasPrefix(k, "filter[") && vs[0] != "":
			// 查询参数
			filtermap[k[7:len(k)-1]] = vs[0]
		}
	}
	return filtermap
}

func ParseEqualMap(c echo.Context) map[string]string {
	filtermap := make(map[string]string)
	for k, vs := range c.QueryParams() {
		switch {
		case strings.HasPrefix(k, "equal[") && vs[0] != "":
			// 查询参数
			filtermap[k[6:len(k)-1]] = vs[0]
		}
	}
	return filtermap
}

func CreateToken(secret, uid, level string, exp time.Duration) (string, error) {
	// Create token
	token := jwt.New(jwt.SigningMethodHS256)
	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["usr"] = uid
	claims["uid"] = uid
	claims["lvl"] = level
	claims["exp"] = time.Now().Add(exp).Unix()
	return token.SignedString([]byte(secret))
}

func QueryPageResult[T any](c echo.Context, tx *gorm.DB, prequery *PreQuery) (*PageResult, error) {
	var count, start int
	NewParamReader(c).
		ReadInt(&start, "start", 0).
		ReadInt(&count, "count", 40)
	var data []T
	var total int64
	models := new(T)
	common.Must(prequery.Query(tx.Model(models)).Count(&total).Error)
	query := prequery.Query(tx.Debug().Model(models)).Offset(start).Limit(count)
	if err := query.Find(&data).Error; err != nil {
		return nil, err
	}
	return &PageResult{TotalCount: total, Pos: int64(start), Data: data}, nil
}

func QueryDataResult[T any](c echo.Context, tx *gorm.DB, prequery *PreQuery) ([]T, error) {
	var count int
	NewParamReader(c).ReadInt(&count, "count", 10000)
	var data []T
	if err := prequery.Query(tx.Model(new(T))).Limit(count).Find(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}
