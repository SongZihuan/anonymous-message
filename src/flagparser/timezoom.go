package flagparser

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

var location *time.Location
var locationOnce sync.Once

func TimeZone() *time.Location {
	locationOnce.Do(func() {
		if strings.ToLower(_TimeZone) == "utc" {
			_location := time.UTC
			if _location == nil {
				_location = time.Local
			}

			if _location != nil {
				location = _location
			}
		} else if strings.ToLower(_TimeZone) == "local" || _TimeZone == "" {
			_location := time.Local
			if _location == nil {
				_location = time.UTC
			}

			if _location != nil {
				location = _location
			}
		} else {
			_location, err := time.LoadLocation(_TimeZone)
			if err != nil || _location == nil {
				_location = time.UTC
			}

			if _location != nil {
				location = _location
			}
		}

		if location == nil {
			if _TimeZone == "UTC" || _TimeZone == "Local" || _TimeZone == "" {
				panic(fmt.Errorf("can not get location UTC or Local"))
			}
			panic(fmt.Errorf("can not get location UTC, Local or %s", _TimeZone))
		}
	})

	return location
}
