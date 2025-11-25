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

package gmail

import (
	"crypto/tls"
	"fmt"
	"mime"
	"os"
	"path"

	"gopkg.in/gomail.v2"
)

type MailSender struct {
	Server   string
	Port     int
	Tls      bool
	Usernam  string
	Alias    string
	Password string
	Mailtos  []string
}

func (s *MailSender) SendMail(mailTo []string, subject string, body string, files []string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(s.Usernam, s.Alias))
	if len(mailTo) == 0 {
		if len(s.Mailtos) == 0 {
			return fmt.Errorf("mail receiver not configured")
		}
		m.SetHeader("To", s.Mailtos...)
	}

	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	if len(files) > 0 {
		m := gomail.NewMessage(
			gomail.SetEncoding(gomail.Base64),
		)
		for _, filename := range files {
			info, err := os.Stat(filename)
			if err != nil {
				return fmt.Errorf("file %s not exists", filename)
			}
			if info.IsDir() {
				return fmt.Errorf("file %s is dir", filename)
			}
			name := path.Base(filename)
			m.Attach(filename,
				gomail.Rename(name),
				gomail.SetHeader(map[string][]string{
					"Content-Disposition": []string{
						fmt.Sprintf(`attachment; filename="%s"`, mime.BEncoding.Encode("UTF-8", name)),
					},
				}),
			)
		}
	}

	d := gomail.NewDialer(s.Server, s.Port, s.Usernam, s.Password)
	if s.Tls {
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return d.DialAndSend(m)
}
