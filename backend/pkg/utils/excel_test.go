package utils

import (
	"os"
	"testing"
)

type Product struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Price    float32 `json:"price"`
	Quantity int     `json:"quantity"`
}

func TestWriteSliceToExcel(t *testing.T) {
	var products = []Product{
		{ID: 1, Name: "Product 1", Price: 10.0, Quantity: 10},
		{ID: 2, Name: "Product 2", Price: 20.0, Quantity: 20},
		{ID: 3, Name: "Product 3", Price: 30.0, Quantity: 30},
	}
	var products2 = []*Product{
		{ID: 1, Name: "Product 1", Price: 10.0, Quantity: 10},
		{ID: 2, Name: "Product 2", Price: 20.0, Quantity: 20},
		{ID: 3, Name: "Product 3", Price: 30.0, Quantity: 30},
	}

	// var products3 = &products

	if err := WriteSliceToExcel("products.xlsx", products); err != nil {
		t.Error(err)
	}
	if err := WriteSliceToExcel("products2.xlsx", products2, "json"); err != nil {
		t.Error(err)
	}
	// if err := WriteSliceToExcel("products3.xlsx", products3, "excel"); err != nil {
	// 	t.Error(err)
	// }

	// 清理测试生成的文件
	if err := os.Remove("products.xlsx"); err != nil {
		t.Error(err)
	}
	if err := os.Remove("products2.xlsx"); err != nil {
		t.Error(err)
	}
}
