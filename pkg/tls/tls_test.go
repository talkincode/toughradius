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
 */

package tls

import (
	"testing"
)

func TestGenCaCrt(t *testing.T) {
	cacfg := Config{
		Country:            []string{"CN"},
		OrganizationalUnit: []string{"Apache"},
		Organization:       []string{"Apache"},
		Locality:           []string{"Hangzhou"},
		Province:           []string{"Zhejiang"},
		StreetAddress:      []string{"Xihu"},
		PostalCode:         []string{"310000"},
		CommonName:         "Apache",
		Years:              10,
	}
	err := GenerateCaCrt(cacfg, "ca.crt", "ca.key")
	if err != nil {
		t.Fatal(err)
	}
	mycfg := Config{
		Country:            []string{"CN"},
		OrganizationalUnit: []string{"Apache"},
		Organization:       []string{"Apache"},
		Locality:           []string{"Hangzhou"},
		Province:           []string{"Zhejiang"},
		StreetAddress:      []string{"Xihu"},
		PostalCode:         []string{"310000"},
		CommonName:         "Apache",
		DNSNames:           []string{"localhost"},
		Years:              10,
	}
	err = GenerateCrt(mycfg, "ca.crt", "ca.key", "my.crt", "my.key")
	if err != nil {
		t.Fatal(err)
	}
}
