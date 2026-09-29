package utils

import (
	"fmt"
	"reflect"
	"strings"

	excelize "github.com/xuri/excelize/v2"
)

func WriteSliceToExcel(filename string, slice interface{}, structTag ...string) error {
	tag := "excel"
	if len(structTag) > 0 {
		tag = structTag[0]
	}

	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("expected slice, got %T", slice)
	}
	if v.Len() == 0 {
		return fmt.Errorf("empty slice")
	}

	// 获取切片的元素类型（可能是结构体或指针）
	elemType := v.Type().Elem()
	// 如果元素是指针，则取其指向的类型（结构体）
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	// 确保是结构体类型
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("slice element must be struct or pointer to struct, got %v", elemType.Kind())
	}

	f := excelize.NewFile()
	sheetName := "Sheet1"

	// 生成表头（基于结构体字段的 excel tag）
	headers := make([]string, elemType.NumField())
	for i := 0; i < elemType.NumField(); i++ {
		tag := elemType.Field(i).Tag.Get(tag)
		// 去除首尾空格 和,逗号前的一部分
		tags := strings.Split(tag, ",")
		tag = strings.TrimSpace(tags[0])

		// 如果没有 tag，则使用字段名
		if tag == "" || tag == "-" {
			tag = elemType.Field(i).Name
		}
		headers[i] = tag
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}

	// 写入数据行
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i)
		// 如果元素是指针，则解引用获取结构体值
		if item.Kind() == reflect.Ptr {
			if item.IsNil() {
				continue // 跳过 nil 指针，或按需处理
			}
			item = item.Elem()
		}
		row := i + 2
		for j := 0; j < elemType.NumField(); j++ {
			fieldVal := item.Field(j)
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue(sheetName, cell, fieldVal.Interface())
		}
	}

	return f.SaveAs(filename)
}
