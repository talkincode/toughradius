package iploc

import (
	"bytes"
	"io"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var provincePrefix = map[int32]map[int32]int{'北': {'京': 0}, '天': {'津': 0}, '河': {'北': 0, '南': 0}, '山': {'东': 0, '西': 0}, '内': {'蒙': 1}, '辽': {'宁': 0}, '吉': {'宁': 0}, '黑': {'龙': 1}, '上': {'海': 0}, '江': {'苏': 0, '西': 0}, '浙': {'江': 0}, '安': {'徽': 0}, '福': {'福': 0}, '湖': {'南': 0, '北': 0}, '广': {'东': 0, '西': 0}, '海': {'南': 0}, '重': {'庆': 0}, '四': {'川': 0}, '贵': {'州': 0}, '云': {'南': 0}, '西': {'藏': 0}, '陕': {'西': 0}, '甘': {'肃': 0}, '青': {'海': 0}, '宁': {'夏': 0}, '新': {'疆': 0}, '香': {'港': 0}, '澳': {'门': 0}, '台': {'湾': 0}}
var adEndSuffix = [2]map[int32]int{{'市': 1, '州': 1, '区': 1, '盟': 1}, {'县': 2, '市': 2, '旗': 2}}

type Detail struct {
	IP    IP
	Start IP
	End   IP
	Location
}

func (detail Detail) String() string {
	return detail.Location.String()
}

func (detail Detail) Bytes() []byte {
	return detail.Location.Bytes()
}

func (detail Detail) InIP(ip IP) bool {
	return detail.Start.Compare(ip) < 1 && detail.End.Compare(ip) > -1
}

func (detail Detail) In(rawIP string) bool {
	ip, err := ParseIP(rawIP)
	if err != nil {
		return false
	}
	return detail.InIP(ip)
}

func (detail Detail) InUint(uintIP uint32) bool {
	return detail.InIP(ParseUintIP(uintIP))
}

func (detail *Detail) fill() *Detail {
	if detail.Region == "N/A" {
		return detail
	}

	var (
		rs   = []rune(detail.Country)
		s    [2][]rune
		p    map[int32]int
		ok   bool
		i, n int
	)

	if p, ok = provincePrefix[rs[0]]; ok {
		i, ok = p[rs[1]]
		i += 2
	}
	if !ok {
		return detail
	}

	detail.Country = "中国"
	detail.Province = string(rs[:i])

	if i >= len(rs) {
		if rs[0] == '北' || rs[0] == '天' || rs[0] == '上' || rs[0] == '重' {
			detail.City = string(rs[:i-1])
		}
		return detail
	}

	if rs[i] == '市' {
		i++
		detail.City = string(rs[:i])
	} else if rs[i] == '省' {
		i++
	}

	for ; i < len(rs); i++ {
		s[n] = append(s[n], rs[i])
		if _, ok = adEndSuffix[n][rs[i]]; ok {
			if rs[i] != '市' && i+1 < len(rs) && rs[i+1] == '市' {
				continue
			}
			n++
		}
		if n > 1 {
			break
		}
	}

	if detail.City != "" {
		detail.County = string(s[0])
	} else {
		detail.City = string(s[0])
		detail.County = string(s[1])
	}

	return detail
}

type Location struct {
	Country  string
	Region   string
	Province string
	City     string
	County   string
	raw      string
}

func (location Location) String() string {
	return gbkToUtf8(location.Bytes())
}

func (location Location) GetCity() string {
	return gbkToUtf8([]byte(location.City))
}

func (location Location) GetRegion() string {
	return gbkToUtf8([]byte(location.Region))
}

func (location Location) GetProvince() string {
	return gbkToUtf8([]byte(location.Province))
}

func (location Location) GetCountry() string {
	return gbkToUtf8([]byte(location.Country))
}

func (location Location) GetCounty() string {
	return gbkToUtf8([]byte(location.County))
}

func (location Location) Bytes() []byte {
	return []byte(location.raw)
}

func parseLocation(country, region []byte) Location {
	location := Location{
		Country: string(country),
		Region:  string(region),
	}
	location.raw = location.Country
	if region != nil {
		location.raw += " " + location.Region
	}
	return location
}

func gbkToUtf8(s []byte) string {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := io.ReadAll(reader)
	if e != nil {
		return ""
	}
	return string(d)
}
