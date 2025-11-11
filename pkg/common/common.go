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

package common

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/rand"
	sha1_ "crypto/sha1"
	sha256_ "crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	mathrand "math/rand"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"github.com/bwmarrin/snowflake"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var (
	EmptyList      []interface{}
	EmptyData      map[string]interface{}
	ShangHaiLOC, _ = time.LoadLocation("Asia/Shanghai")
)

const (
	NA         = "N/A"
	ENABLED    = "enabled"
	DISABLED   = "disabled"
	SecretSalt = "Teamsacsca172021"
)

func init() {
}

// Usage print usage
func Usage(str string) {
	fmt.Fprintf(os.Stderr, str)
	flag.PrintDefaults()
}

// MakeDir Create directory
func MakeDir(path string) {
	f, err := os.Stat(path)
	if err != nil || f.IsDir() == false {
		if err := os.Mkdir(path, os.ModePerm); err != nil {
			fmt.Println("create dir fail！", err)
			return
		}
	}
}

func FileExists(file string) bool {
	info, err := os.Stat(file)
	return err == nil && !info.IsDir()
}

func Copy(from, to string) error {
	f, e := os.Stat(from)
	if e != nil {
		return e
	}
	if f.IsDir() {
		// fromis directory，then definetois also directory
		if list, e := ioutil.ReadDir(from); e == nil {
			for _, item := range list {
				if e = Copy(filepath.Join(from, item.Name()), filepath.Join(to, item.Name())); e != nil {
					return e
				}
			}
		}
	} else {
		// fromis file，then createtodirectory
		p := filepath.Dir(to)
		if _, e = os.Stat(p); e != nil {
			if e = os.MkdirAll(p, 0777); e != nil {
				return e
			}
		}
		// Read source file
		file, e := os.Open(from)
		if e != nil {
			return e
		}
		defer file.Close()
		bufReader := bufio.NewReader(file)
		// Create a file to save
		out, e := os.Create(to)
		if e != nil {
			return e
		}
		defer out.Close()
		// Then connect file streams
		_, e = io.Copy(out, bufReader)
	}
	return e
}

func CopyFile(from io.Reader, to string, mode os.FileMode) error {
	// fromis file，then createtodirectory
	p := filepath.Dir(to)
	if _, e := os.Stat(p); e != nil {
		if e = os.MkdirAll(p, mode); e != nil {
			return e
		}
	}
	// Create a file to save
	out, e := os.Create(to)
	if e != nil {
		return e
	}
	defer out.Close()
	// Then connect file streams
	_, e = io.Copy(out, from)
	return e
}

func DirExists(file string) bool {
	info, err := os.Stat(file)
	return err == nil && info.IsDir()
}

// create a file to test the upload and download.
func CreateTmpFile(data []byte) (*os.File, error) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "toughradius-")
	if err != nil {
		return nil, fmt.Errorf("cannot create temporary file, %s", err.Error())
	}
	_, err = tmpFile.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to write to temporary file, %s", err.Error())
	}
	return tmpFile, nil
}

// panic error
func Must(err error) {
	if err != nil {
		panic(errors.WithStack(err))
	}
}

func MustNotEmpty(name string, v interface{}) {
	if IsEmpty(v) {
		panic(errors.New(name + " is empty"))
	}
}

// panic error
func MustStringValue(val string, err error) string {
	if err != nil {
		panic(err)
	}
	if val == "" || val == "N/A" {
		panic(fmt.Errorf("value cannot be null or N/A"))
	}
	return val
}

// panic error
func MustDebug(err error, debug bool) {
	if err != nil {
		if debug {
			panic(errors.WithStack(err))
		} else {
			panic(err)
		}
	}
}

func MustCallBefore(err error, callbefore func()) {
	if err != nil {
		callbefore()
		panic(errors.WithStack(err))
	}
}

func Must2(v interface{}, err error) interface{} {
	Must(err)
	return v
}

func IgnoreError(v interface{}, err error) interface{} {
	return v
}

func UUID() string {
	unix32bits := uint32(time.Now().UTC().Unix())
	buff := make([]byte, 12)
	numRead, err := rand.Read(buff)
	if numRead != len(buff) || err != nil {
		Must(err)
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x-%x", unix32bits, buff[0:2], buff[2:4], buff[4:6], buff[6:8], buff[8:])
}

var snowflakeNode, _ = snowflake.NewNode(int64(mathrand.Intn(1000)))

// Generate int64
func UUIDint64() int64 {
	return snowflakeNode.Generate().Int64()
}

func UUIDBase32() (string, error) {
	id := snowflakeNode.Generate()
	// Print out the ID in a few different ways.
	return id.Base32(), nil
}

// Convert to Big Hump format
func ToCamelCase(str string) string {
	temp := strings.Split(str, "_")
	for i, r := range temp {
		temp[i] = strings.Title(r)
	}
	return strings.Join(temp, "")
}

var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

// Convert to underlined format
func ToSnakeCase(str string) string {
	snake := matchAllCap.ReplaceAllString(str, "${1}_${2}")
	return strings.ToLower(snake)
}

func Sha1Hash(src string) string {
	h := sha1_.New()
	h.Write([]byte(src))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func Sha256Hash(src string) string {
	h := sha256_.New()
	h.Write([]byte(src))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func Sha256HashWithSalt(src string, salt string) string {
	h := sha256_.New()
	h.Write([]byte(src))
	h.Write([]byte(salt))
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}

// Determine if the string is in the list.
func InSlice(v string, sl []string) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

func IfEmpty(src interface{}, defval interface{}) interface{} {
	if IsEmpty(src) {
		return defval
	}
	return src
}

func IfNA(src string, defval string) string {
	if src == "N/A" || src == "" {
		return defval
	}
	return src
}

func EmptyToNA(src string) string {
	if strings.TrimSpace(src) == "" {
		return "N/A"
	}
	return src
}

func IfEmptyStr(src string, defval string) string {
	if src == "" {
		return defval
	}
	return src
}

// IsEmpty checks if a value is empty or not.
// A value is considered empty if
// - integer, float: zero
// - bool: false
// - string, array: len() == 0
// - slice, map: nil or len() == 0
// - interface, pointer: nil or the referenced value is empty
func IsEmpty(value interface{}) bool {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String, reflect.Array, reflect.Map, reflect.Slice:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Invalid:
		return true
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return true
		}
		return IsEmpty(v.Elem().Interface())
	case reflect.Struct:
		v, ok := value.(time.Time)
		if ok && v.IsZero() {
			return true
		}
	}

	return false
}

func CheckEmpty(name string, v interface{}) {
	if IsEmpty(v) {
		panic(errors.New(name + " is empty"))
	}
}

func IsNotEmpty(value interface{}) bool {
	return !IsEmpty(value)
}

func split(s string, size int) []string {
	ss := make([]string, 0, len(s)/size+1)
	for len(s) > 0 {
		if len(s) < size {
			size = len(s)
		}
		ss, s = append(ss, s[:size]), s[size:]

	}
	return ss
}

func File2Base64(file string) string {
	data := Must2(ioutil.ReadFile(file))
	return base64.StdEncoding.EncodeToString(data.([]byte))
}

func Base642file(b64str string, file string) error {
	data, err := base64.StdEncoding.DecodeString(b64str)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, data, 777)
}

func parseWithLocation(name string, timeStr string) (time.Time, error) {
	locationName := name
	if l, err := time.LoadLocation(locationName); err != nil {
		println(err.Error())
		return time.Time{}, err
	} else {
		lt, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr, l)
		fmt.Println(locationName, lt)
		return lt, nil
	}
}

var mobileRe, _ = regexp.Compile("(?i:Mobile|iPod|iPhone|Android|Opera Mini|BlackBerry|webOS|UCWEB|Blazer|PSP)")

func MobileAgent(userAgent string) string {
	return mobileRe.FindString(userAgent)
}

// Generate checksum
func GenValidateCode(width int) string {
	numeric := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := len(numeric)
	mathrand.Seed(time.Now().UnixNano())

	var sb strings.Builder
	for i := 0; i < width; i++ {
		fmt.Fprintf(&sb, "%d", numeric[mathrand.Intn(r)])
	}
	return sb.String()
}

func SetEmptyStrToNA(t interface{}) {
	d := reflect.TypeOf(t).Elem()
	for j := 0; j < d.NumField(); j++ {
		ctype := d.Field(j).Type.String()
		if ctype == "string" {
			val := reflect.ValueOf(t).Elem().Field(j)
			if val.String() == "" {
				val.SetString(NA)
			}
		}
	}
}

func IsEmptyOrNA(val string) bool {
	val = strings.TrimSpace(val)
	return val == "" || val == NA
}

func IsNotEmptyAndNA(val string) bool {
	val = strings.TrimSpace(val)
	return strings.TrimSpace(val) != "" && val != NA
}

func Md5HashFile(filePath string) (string, error) {
	var returnMD5String string
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}
	hashInBytes := hash.Sum(nil)[:16]
	returnMD5String = hex.EncodeToString(hashInBytes)
	return returnMD5String, nil
}

func Md5Hash(src string) string {
	hash := md5.Sum([]byte(src))
	return hex.EncodeToString(hash[:])
}

func UrlJoin(hurl string, elm ...string) string {
	u, err := url.Parse(hurl)
	Must(err)
	u.Path = path.Join(u.Path, path.Join(elm...))
	return u.String()
}

func UrlJoin2(hurl string, elm ...string) string {
	u, err := url.Parse(hurl)
	Must(err)
	u.Path = path.Join(u.Path, path.Join(elm...))
	sb := strings.Builder{}
	sb.WriteString(u.Scheme)
	sb.WriteString("://")
	sb.WriteString(u.Host)
	sb.WriteString(u.Path)
	return sb.String()
}

var notfloat = errors.New("not float value")

func ParseFloat64(v interface{}) (float64, error) {
	switch v.(type) {
	case float64:
		return v.(float64), nil
	case int64:
		return float64(v.(int64)), nil
	case int:
		return float64(v.(int)), nil
	case string:
		fv, err := strconv.ParseFloat(v.(string), 64)
		if err != nil {
			return 0, err
		}
		return fv, nil
	}
	return 0, notfloat
}

var notint = errors.New("not int value")

func ParseInt64(v interface{}) (int64, error) {
	switch v.(type) {
	case float64:
		return int64(v.(float64)), nil
	case int64:
		return v.(int64), nil
	case int:
		return int64(v.(int)), nil
	case string:
		ival, err := strconv.ParseInt(v.(string), 10, 64)
		if err != nil {
			return 0, err
		}
		return ival, nil
	}
	return 0, notint
}

func ParseString(v interface{}) (string, error) {
	switch v.(type) {
	case float64:
		return strconv.FormatFloat(v.(float64), 'f', 2, 64), nil
	case int64:
		return strconv.FormatInt(v.(int64), 10), nil
	case int:
		return strconv.Itoa(v.(int)), nil
	case string:
		return v.(string), nil
	case nil:
		return "", nil
	case time.Time:
		return v.(time.Time).Format("2006-01-02 15:04:05"), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func ToGbkHexString(src string) (string, error) {
	var buf strings.Builder
	reader := transform.NewReader(bytes.NewReader([]byte(src)), simplifiedchinese.GBK.NewEncoder())
	data, e := ioutil.ReadAll(reader)
	if e != nil {
		return "", e
	}
	for _, b := range data {
		buf.WriteByte('\\')
		buf.WriteString(strings.ToUpper(hex.EncodeToString([]byte{b})))
	}
	return buf.String(), nil
}

func ToGbkString(src string) (string, error) {
	reader := transform.NewReader(bytes.NewReader([]byte(src)), simplifiedchinese.GBK.NewEncoder())
	data, e := ioutil.ReadAll(reader)
	if e != nil {
		return "", e
	}
	return string(data), nil
}

func GetPointString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func GetPointInt64(s *int64) int64 {
	if s != nil {
		return *s
	}
	return 0
}

func GetPointBool(s *bool) bool {
	if s != nil {
		return *s
	}
	return false
}

func GetPointTime(s *time.Time) time.Time {
	if s != nil {
		return *s
	}
	return time.Time{}
}

func GenerateRangeNum(min, max int) int {
	mathrand.Seed(time.Now().Unix())
	randNum := mathrand.Intn(max-min) + min
	return randNum
}

func GenerateDataVer() string {
	mathrand.Seed(time.Now().Unix())
	r1 := mathrand.Intn(600-100) + 100
	r2 := mathrand.Intn(900-500) + 500
	return fmt.Sprintf("%d-%d", r1, r2)
}

// DeepCopy deep copy value
func DeepCopy(value interface{}) interface{} {
	if valueMap, ok := value.(map[string]interface{}); ok {
		newMap := make(map[string]interface{})
		for k, v := range valueMap {
			newMap[k] = DeepCopy(v)
		}
		return newMap
	} else if valueSlice, ok := value.([]interface{}); ok {
		newSlice := make([]interface{}, len(valueSlice))
		for k, v := range valueSlice {
			newSlice[k] = DeepCopy(v)
		}

		return newSlice
	}
	return value
}

func JsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func JsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func FormatTimeDuration(seconds int64) string {
	var day = seconds / (24 * 3600)
	hour := (seconds - day*3600*24) / 3600
	minute := (seconds - day*24*3600 - hour*3600) / 60
	second := seconds - day*24*3600 - hour*3600 - minute*60
	return fmt.Sprintf("%d day %d:%d:%d", day, hour, minute, second)
}

func ReplaceNaN(v float64, r float64) float64 {
	if math.IsNaN(v) {
		return r
	}
	return v
}

func ToJson(v interface{}) string {
	bs, _ := json.MarshalIndent(v, "", "  ")
	return string(bs)
}

func NextDataVar() string {
	var ctime = time.Now()
	return strconv.FormatInt(int64(ctime.YearDay()), 10) + "-" +
		strconv.FormatInt(int64(ctime.Hour()*60+ctime.Minute()), 10)
}

func StructToMap(obj interface{}) (newMap map[string]interface{}, err error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &newMap)
	return
}

func TrimBytes(src []byte) []byte {
	s := bytes.ReplaceAll(src, []byte("\xef\xbb\xbf"), []byte(""))
	return s
}

func GetFieldType(mod interface{}, name string) string {
	f, ok := reflect.TypeOf(mod).FieldByNameFunc(func(s string) bool {
		return strings.ToLower(s) == strings.ToLower(name)
	})
	if ok {
		return f.Type.String()
	}
	return ""
}

// GbkToUtf8 GBK to UTF-8
func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

// Utf8ToGbk UTF-8 to GBK
func Utf8ToGbk(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}
