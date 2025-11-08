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

package aes

import (
	"testing"
)

const key = "12345678123456781234567812345678"

type Item struct {
	Foo string
}

func TestAes(t *testing.T) {
	orig := "hello world"
	t.Log("src：", orig)

	encryptCode, _ := EncryptToB64(orig, key)
	t.Log("encyrpt：", encryptCode)
	t.Log(len(encryptCode))

	decryptCode, _ := DecryptFromB64(encryptCode, key)
	t.Log("result：", decryptCode)
}

func TestAes2(t *testing.T) {
	src := "hello"
	dest, _ := Encrypt([]byte(src), key)
	destb, _ := EncryptToB64(src, key)
	t.Log(dest)
	t.Log(destb)
	res, _ := Decrypt(dest, key)
	t.Log(res, string(res))
}
