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
	"bytes"
	"crypto/rand"
	sha256_ "crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	mathrand "math/rand"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/pkg/errors"
)

var (
	EmptyList      []interface{}
	EmptyData      map[string]interface{}
	ShangHaiLOC, _ = time.LoadLocation("Asia/Shanghai")
)

const (
	NA       = "N/A"
	ENABLED  = "enabled"
	DISABLED = "disabled"
)

// defaultSecretSalt is used only for development/testing when env var is not set
const defaultSecretSalt = "toughradius-dev-salt-change-me"

// GetSecretSalt returns the secret salt from environment variable TOUGHRADIUS_SECRET_SALT.
// Falls back to a default value for development only.
// IMPORTANT: Always set TOUGHRADIUS_SECRET_SALT in production!
func GetSecretSalt() string {
	if salt := os.Getenv("TOUGHRADIUS_SECRET_SALT"); salt != "" {
		return salt
	}
	return defaultSecretSalt
}

// FileExists checks if a file exists
func FileExists(file string) bool {
	info, err := os.Stat(file)
	return err == nil && !info.IsDir()
}

// DirExists checks if a directory exists
func DirExists(file string) bool {
	info, err := os.Stat(file)
	return err == nil && info.IsDir()
}

// Must panics if error is not nil
func Must(err error) {
	if err != nil {
		panic(errors.WithStack(err))
	}
}

// Must2 returns value and panics if error is not nil
func Must2(v interface{}, err error) interface{} {
	Must(err)
	return v
}

// UUID generates a UUID string
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

// UUIDint64 generates a unique int64 ID using snowflake algorithm
func UUIDint64() int64 {
	return snowflakeNode.Generate().Int64()
}

// Sha256HashWithSalt returns SHA256 hash with salt
func Sha256HashWithSalt(src string, salt string) string {
	h := sha256_.New()
	h.Write([]byte(src))
	h.Write([]byte(salt))
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}

// InSlice checks if a string is in a slice
func InSlice(v string, sl []string) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

// If returns trueVal if condition is true, otherwise falseVal
func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

// IfEmptyStr returns defval if src is empty string
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

// IsEmptyOrNA checks if value is empty or "N/A"
func IsEmptyOrNA(val string) bool {
	val = strings.TrimSpace(val)
	return val == "" || val == NA
}

// IsNotEmptyAndNA checks if value is not empty and not "N/A"
func IsNotEmptyAndNA(val string) bool {
	val = strings.TrimSpace(val)
	return strings.TrimSpace(val) != "" && val != NA
}

// JsonMarshal marshals value to JSON bytes
func JsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// JsonUnmarshal unmarshals JSON bytes to value
func JsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// ToJson converts value to formatted JSON string
func ToJson(v interface{}) string {
	bs, _ := json.MarshalIndent(v, "", "  ")
	return string(bs)
}

// TrimBytes removes BOM from bytes
func TrimBytes(src []byte) []byte {
	s := bytes.ReplaceAll(src, []byte("\xef\xbb\xbf"), []byte(""))
	return s
}
