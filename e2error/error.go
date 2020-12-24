package e2error

import "fmt"

var (
	ErrConfigureError   = func(msg string) error { return fmt.Errorf("configuration error %v", msg) }
	ErrUnknown          = func(msg string) error { return fmt.Errorf("unknown error %v", msg) }
	ErrIllegalParameter = func(msg string) error { return fmt.Errorf("illegal parameter %v", msg) }
	ErrEmptyParameter   = func(msg string) error { return fmt.Errorf("parameter %v can't not empty", msg) }
	ErrEmptyValue       = func(msg string) error { return fmt.Errorf("value %v empty", msg) }
)

func CheckErrors(errs ...error) error {
	for idx := range errs {
		if errs[idx] != nil {
			return errs[idx]
		}
	}
	return nil
}
