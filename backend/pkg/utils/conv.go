package utils

import (
	"strconv"
)

func StrToUint(str string) (uint, error) {
	res, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
		// switch e := err.(type) {
		// case *strconv.NumError:
		// 	switch e.Err {
		// 	case strconv.ErrRange:
		// 		return 0, errors.New(str + " 超出了 uint 范围")
		// 	case strconv.ErrSyntax:
		// 		return 0, errors.New(str + " 不是有效的整数")
		// 	default:
		// 		return 0, errors.New(str + "  不是有效的整数")
		// 	}
		// default:
		// 	return 0, errors.New(str + "  不是有效的整数")
		// }
	}
	return uint(res), nil
}

func StringToUint(str string) uint {
	val, err := StrToUint(str)
	if err != nil {
		return 0
	}
	return val
}

func StringToInt(str string) int {
	val, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return val
}

func IntToString(num int) string {
	return strconv.Itoa(num)
}
