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

package mfa

import (
	"fmt"
	"testing"
)

func initAuth(user string) (secret, code string) {
	ng := NewGoogleAuth()
	secret = ng.GetSecret()
	fmt.Println("Secret:", secret)
	// Dynamic code (a 6-digit number is dynamically generated every 30s)
	code, err := ng.GetCode(secret)
	fmt.Println("Code:", code, err)
	// Username
	qrCode := ng.GetQrcode(user, code, "TeamsAcsDemo")
	fmt.Println("Qrcode", qrCode)
	return
}

func TestGoogleAuth_VerifyCode(t *testing.T) {

	// fmt.Println("-----------------Enable two-factor authentication----------------------")
	user := "testxxx@google.com"
	secret, code := initAuth(user)
	t.Log(secret, code)
	t.Log("Information Validation")
	// Authentication, dynamic code (from Google Authenticator or freeotp)
	bool, err := NewGoogleAuth().VerifyCode(secret, code)
	if bool {
		t.Log("√")
	} else {
		t.Fatal("X", err)
	}
}
