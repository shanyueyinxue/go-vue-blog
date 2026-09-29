package utils

func Any[T any](f func(T) bool, arr ...T) bool {
	for _, v := range arr {
		if f(v) {
			return true
		}
	}
	return false
}
func All[T any](f func(T) bool, arr ...T) bool {
	for _, v := range arr {
		if !f(v) {
			return false
		}
	}
	return true
}
