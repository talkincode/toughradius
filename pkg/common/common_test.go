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
	"fmt"
	mathrand "math/rand"
	"net/url"
	"testing"
)

func TestToCamelCase(t *testing.T) {
	t.Log(ToCamelCase("user_name"))
}

func TestToSnakeCase(t *testing.T) {
	t.Log(ToSnakeCase("UserName"))
}

func TestToCamelCase1(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"user_name", args{"user_name"}, "UserName"},
		{"username", args{"username"}, "Username"},
		{"id", args{"id"}, "ID"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToCamelCase(tt.args.str); got != tt.want {
				t.Errorf("ToCamelCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToSnakeCase1(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"ID", args{"ID"}, "id"},
		{"UserName", args{"UserName"}, "user_name"},
		{"UserNAME", args{"UserNAME"}, "user_name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSnakeCase(tt.args.str); got != tt.want {
				t.Errorf("ToSnakeCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnicode(t *testing.T) {
	s, _ := ToGbkHexString("ok")
	fmt.Println(url.QueryEscape(s))
}

func TestUUID(t *testing.T) {
	t.Log(UUID())
}

func TestFormatTimeDuration(t *testing.T) {
	t.Log(FormatTimeDuration(8640046))
}

func TestGetDataver(t *testing.T) {
	t.Log(NextDataVar())
}

func TestSplit(t *testing.T) {
	for i := 0; i < 100; i++ {
		v := mathrand.Intn(1)
		t.Log(v)
	}
}

func TestGetFieldType(t *testing.T) {
	type tt struct {
		Aaa string
		Bbb int64
	}

	t.Log(GetFieldType(tt{}, "aaa"))
}
