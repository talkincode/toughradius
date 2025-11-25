package excel

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/talkincode/toughradius/v9/pkg/timeutil"
)

func WriteToFile(sheet string, records []interface{}, filepath string) error {
	xlsx := excelize.NewFile()
	index := xlsx.NewSheet(sheet)

	for i, t := range records {
		WriteRow(t, i, xlsx, sheet)
	}

	xlsx.SetActiveSheet(index)
	return xlsx.SaveAs(filepath)

}

func WriteToTmpFile(sheet string, records []interface{}) (string, error) {
	filename := fmt.Sprintf("%s-%d.xlsx", sheet, time.Now().Unix())
	tmpdir, _ := os.MkdirTemp("", "excel-export")
	filepath := path.Join(tmpdir, filename)
	xlsx := excelize.NewFile()
	index := xlsx.NewSheet(sheet)

	for i, t := range records {
		WriteRow(t, i, xlsx, sheet)
	}

	xlsx.SetActiveSheet(index)
	return filepath, xlsx.SaveAs(filepath)

}

var COLNAMES = map[int]string{0: "A", 1: "B", 2: "C", 3: "D", 4: "E", 5: "F", 6: "G", 7: "H", 8: "I", 9: "J", 10: "K", 11: "L", 12: "M",
	13: "N", 14: "O", 15: "P", 16: "Q", 17: "R", 18: "S", 19: "T", 20: "U", 21: "V", 22: "W", 23: "X", 24: "Y",
	25: "Z", 26: "AA", 27: "AB", 28: "AC", 29: "AD", 30: "AE", 31: "AF", 32: "AG", 33: "AH", 34: "AI", 35: "AJ",
	36: "AK", 37: "AL", 38: "AM", 39: "AN", 40: "AO", 41: "AP", 42: "AQ", 43: "AR", 44: "AS", 45: "AT", 46: "AU", 47: "AV", 48: "AW",
}

func WriteRow(t interface{}, i int, xlsx *excelize.File, sheet string) {
	d := reflect.TypeOf(t).Elem()
	count := 0
	for j := 0; j < d.NumField(); j++ {
		// Set the table header
		column := COLNAMES[count]
		count++
		if i == 0 {
			xtag := d.Field(j).Tag.Get("db")
			if xtag == "" || xtag == "-" {
				xtag = d.Field(j).Tag.Get("json")
				xtag = strings.TrimSuffix(xtag, ",string")
			}
			if xtag == "" || xtag == "-" {
				continue
			}
			xlsx.SetCellValue(sheet, fmt.Sprintf("%s%d", column, i+1), xtag)
		}
		// Set the content
		// column := strings.Split(d.Field(j).Tag.Get("xlsx"), "-")[0]
		ctype := d.Field(j).Type.String()
		switch ctype {
		case "string":
			xlsx.SetCellValue(sheet, fmt.Sprintf("%s%d", column, i+2), reflect.ValueOf(t).Elem().Field(j).String())
		case "int":
			xlsx.SetCellValue(sheet, fmt.Sprintf("%s%d", column, i+2), fmt.Sprintf("%d", reflect.ValueOf(t).Elem().Field(j).Int()))
		case "int32":
			xlsx.SetCellValue(sheet, fmt.Sprintf("%s%d", column, i+2), fmt.Sprintf("%d", reflect.ValueOf(t).Elem().Field(j).Int()))
		case "int64":
			xlsx.SetCellValue(sheet, fmt.Sprintf("%s%d", column, i+2), fmt.Sprintf("%d", reflect.ValueOf(t).Elem().Field(j).Int()))
		case "bool":
			xlsx.SetCellValue(sheet, fmt.Sprintf("%s%d", column, i+2), fmt.Sprintf("%v", reflect.ValueOf(t).Elem().Field(j).Bool()))
		case "float32":
			xlsx.SetCellValue(sheet, fmt.Sprintf("%s%d", column, i+2), fmt.Sprintf("%f", reflect.ValueOf(t).Elem().Field(j).Float()))
		case "float64":
			xlsx.SetCellValue(sheet, fmt.Sprintf("%s%d", column, i+2), fmt.Sprintf("%f", reflect.ValueOf(t).Elem().Field(j).Float()))
		case "time.Time":
			xlsx.SetCellValue(sheet, fmt.Sprintf("%s%d", column, i+2), reflect.ValueOf(t).Elem().Field(j).Interface().(time.Time).Format("2006-01-02 15:04:05"))
		case "timeutil.LocalTime":
			xlsx.SetCellValue(sheet, fmt.Sprintf("%s%d", column, i+2), time.Time(reflect.ValueOf(t).Elem().Field(j).Interface().(timeutil.LocalTime)).Format("2006-01-02 15:04:05"))
		default:
			xlsx.SetCellValue(sheet, fmt.Sprintf("%s%d", column, i+2), reflect.ValueOf(t).Elem().Field(j).String())
		}
	}
}
