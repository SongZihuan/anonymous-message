package flagparser

import "fmt"

func InitFlagParser() error {
	err := initFlag()
	if err != nil {
		return fmt.Errorf("init flag error: %v", err)
	}

	return nil
}
