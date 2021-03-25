package utils

import "fmt"

func Error(info, module string) error {
	return fmt.Errorf("%s; module=%s", info, module)
}
