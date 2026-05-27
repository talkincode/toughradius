/*
 * Copyright (c) 2024-2025 TalkingCode
 * Licensed under the MIT License. See LICENSE file in the project root for details.
 */

package validutil

import (
	"regexp"
	"strconv"
	"unicode"
)

const (
	// Matching china calls
	cnPhoneStr = `((\d{3,4})-?)?` + // area code
		`\d{7,8}` + // serial number
		`(-\d{1,4})?` // The extension number, extension number connection symbol cannot be omitted.

	// Matching china Mobile
	cnMobileStr = `(0|\+?86)?` + // Match 0,86,+86
		`(13[0-9]|` + // 130-139
		`14[57]|` + // 145,147
		`15[0-35-9]|` + // 150-153,155-159
		`17[0678]|` + // 170,176,177,17u
		`18[0-9])` + // 180-189
		`[0-9]{8}`

	// Match email
	emailStr = `[\w.-]+@[\w_-]+\w{1,}[\.\w-]+`

	// IP4
	ip4Str = `((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)`

	// IP6，Refer to the following web pages.
	// http://blog.csdn.net/jiangfeng08/article/details/7642018
	ip6Str = `(([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|` +
		`(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|` +
		`(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|` +
		`(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|` +
		`(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|` +
		`(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|` +
		`(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|` +
		`(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))`

	// Matches both IP4 and IP6
	ipStr = "(" + ip4Str + ")|(" + ip6Str + ")"

	// match a domain name
	domainStr = `[a-zA-Z0-9][a-zA-Z0-9_-]{0,62}(\.[a-zA-Z0-9][a-zA-Z0-9_-]{0,62})*(\.[a-zA-Z][a-zA-Z0-9]{0,10}){1}`

	// Mach URL
	urlStr = `((https|http|ftp|rtsp|mms)?://)?` + // protocols
		`(([0-9a-zA-Z]+:)?[0-9a-zA-Z_-]+@)?` + // pwd:user@
		"(" + ipStr + "|(" + domainStr + "))" + // IP or domain name
		`(:\d{1,4})?` + // port
		`(/+[a-zA-Z0-9][a-zA-Z0-9_.-]*/*)*` + // path
		`(\?([a-zA-Z0-9_-]+(=[a-zA-Z0-9_-]*)*)*)*` // query
)

func regexpCompile(str string) *regexp.Regexp {
	return regexp.MustCompile("^" + str + "$")
}

var (
	email    = regexpCompile(emailStr)
	ip4      = regexpCompile(ip4Str)
	ip6      = regexpCompile(ip6Str)
	ip       = regexpCompile(ipStr)
	url      = regexpCompile(urlStr)
	cnPhone  = regexpCompile(cnPhoneStr)
	cnMobile = regexpCompile(cnMobileStr)
)

// Determine if val can match the regular expression in exp correctly.
// val can be of type []byte, []rune, string.
func isMatch(exp *regexp.Regexp, val interface{}) bool {
	switch v := val.(type) {
	case []rune:
		return exp.MatchString(string(v))
	case []byte:
		return exp.Match(v)
	case string:
		return exp.MatchString(v)
	default:
		return false
	}
}

// Verify phone numbers in mainland China. The following formats are supported.
// 0578-12345678-1234
// 057812345678-1234
// If an extension number exists, the extension number ligature cannot be omitted.
func IsCnPhone(val interface{}) bool {
	return isMatch(cnPhone, val)
}

// Verify mobile phone numbers in  China
func IsCnMobile(val interface{}) bool {
	return isMatch(cnMobile, val)
}

// Verify that a value is in a standard URL format. Supports formats such as IP and domain name.
func IsURL(val interface{}) bool {
	return isMatch(url, val)
}

// Verify that a value is an IP, verifies IP4 and IP6
func IsIP(val interface{}) bool {
	return isMatch(ip, val)
}

// Verify that a value is IP6
func IsIP6(val interface{}) bool {
	return isMatch(ip6, val)
}

// Verify that a value is IP4.
func IsIP4(val interface{}) bool {
	return isMatch(ip4, val)
}

// Verify that a value matches a mailbox.
func IsEmail(val interface{}) bool {
	return isMatch(email, val)
}

func CheckPassword(s string) bool {
	var hasNumber, hasLetter bool
	for _, c := range s {
		switch {
		case unicode.IsNumber(c):
			hasNumber = true
		case unicode.IsLetter(c):
			hasLetter = true
		}
	}
	return hasNumber && hasLetter
}

func CheckStrongPassword(s string) bool {
	var hasNumber, hasUpperCase, hasLowercase, hasSpecial bool
	for _, c := range s {
		switch {
		case unicode.IsNumber(c):
			hasNumber = true
		case unicode.IsUpper(c):
			hasUpperCase = true
		case unicode.IsLower(c):
			hasLowercase = true
		case c == '#' || c == '|':
			return false
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true
		}
	}
	return hasNumber && hasUpperCase && hasLowercase && hasSpecial
}

func IsInt(val interface{}) bool {
	switch v := val.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	case []byte:
		_, err := strconv.ParseInt(string(v), 10, 64)
		return err == nil
	case string:
		_, err := strconv.ParseInt(v, 10, 64)
		return err == nil
	case []rune:
		_, err := strconv.ParseInt(string(v), 10, 64)
		return err == nil
	default:
		return false
	}
}
