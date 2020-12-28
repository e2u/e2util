package e2error

import "fmt"

var (
	ErrConfigureError   = func(msg string) error { return fmt.Errorf("configuration error %v", msg) }
	ErrUnknown          = func(msg string) error { return fmt.Errorf("unknown error %v", msg) }
	ErrIllegalParameter = func(msg string) error { return fmt.Errorf("illegal parameter %v", msg) }
	ErrEmptyParameter   = func(msg string) error { return fmt.Errorf("parameter %v can't not empty", msg) }
	ErrEmptyValue       = func(msg string) error { return fmt.Errorf("value %v empty", msg) }
)

// CheckErrors 檢查傳入的多個錯誤,如果 stop ==true 則遇到第一個錯誤就終止操作,否則返回最後一個非空錯誤
func CheckErrors(stop bool, errs ...error) error {
	var lastErr error
	for idx := range errs {
		if lastErr = errs[idx]; lastErr != nil {
			if stop {
				return errs[idx]
			}
		}
	}
	return lastErr
}
