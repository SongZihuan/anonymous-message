package utils

func IsChanClose[T any](ch chan T) bool {
	if ch == nil {
		return true
	}

	select {
	case _, ok := <-ch:
		if !ok {
			return true
		} else {
			return false
		}
	default:
		return false
	}
}

func IsChanOpen[T any](ch chan T) bool {
	return !IsChanClose(ch)
}
