package internal

import "fmt"

type SendError struct {
	Code    int
	Err     error
	Message string
}

func (s *SendError) Error() string {
	if s.Err == nil {
		return s.Message
	}

	return fmt.Sprintf("%s: %s", s.Message, s.Err.Error())
}
