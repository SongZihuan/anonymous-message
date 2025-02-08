package flagparser

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

var location *time.Location
var locationOnce sync.Once

func TimeZoom() *time.Location {
	locationOnce.Do(func() {
		if strings.ToLower(_TimeZoom) == "utc" {
			_location := time.UTC
			if _location == nil {
				_location = time.Local
			}

			if _location != nil {
				location = _location
			}
		} else if strings.ToLower(_TimeZoom) == "local" || _TimeZoom == "" {
			_location := time.Local
			if _location == nil {
				_location = time.UTC
			}

			if _location != nil {
				location = _location
			}
		} else {
			_location, err := time.LoadLocation(_TimeZoom)
			if err != nil || _location == nil {
				_location = time.UTC
			}

			if _location != nil {
				location = _location
			}
		}

		if location == nil {
			if _TimeZoom == "UTC" || _TimeZoom == "Local" || _TimeZoom == "" {
				panic(fmt.Errorf("can not get location UTC or Local"))
			}
			panic(fmt.Errorf("can not get location UTC, Local or %s", _TimeZoom))
		}
	})

	return location
}
