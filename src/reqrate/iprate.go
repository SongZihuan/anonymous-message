package reqrate

import (
	"fmt"
	"golang.org/x/time/rate"
	"net"
	"sync"
	"time"
)

// RateExp: 计算周期
// RateMaxCount: 周期内最大发件数量
const (
	ipRateExp      time.Duration = 1 * time.Hour
	ipRateMaxCount int           = 36
)

type IP any

type ipRate struct {
	IP      string
	Rate    *rate.Limiter
	Time    time.Time
	RateExp time.Duration
}

var ipRaterMap sync.Map

func getIPRate(ip IP) (rater *ipRate) {
	rater = &ipRate{
		IP:      _getIP(ip),
		Rate:    rate.NewLimiter(rate.Every(ipRateExp), ipRateMaxCount),
		Time:    time.Now(),
		RateExp: ipRateExp,
	}

	raterInterface, ok := ipRaterMap.LoadOrStore(rater.GetName(), rater)
	rater, ok = raterInterface.(*ipRate)
	if !ok {
		panic("sync.map error")
	}

	return rater
}

func (a *ipRate) GetName() string {
	return _getIPName(a.IP)
}

func _getIPName(ip IP) string {
	return fmt.Sprintf("%s", _getIP(ip))
}

func _getIP(ip IP) string {
	switch a := ip.(type) {
	case *net.IP:
		return a.String()
	case net.IP:
		return a.String()
	case string:
		return a
	default:
		panic("not a valid ip")
	}
}

func CleanIPRate() {
	for range time.Tick(1 * time.Minute) {
		func() {
			defer func() {
				_ = recover()
			}()

			now := time.Now()
			delList := make([]any, 0, 10)

			ipRaterMap.Range(func(key, value any) bool {
				rater, ok := value.(*ipRate)
				if !ok {
					delList = append(delList, key)
					return true
				}

				if rater.Time.Add(rater.RateExp).Before(now) {
					delList = append(delList, key)
					return true
				}

				return true
			})

			for _, v := range delList {
				_, _ = ipRaterMap.LoadAndDelete(v)
			}
		}()
	}
}

func CheckHttpReqIP(ip IP) bool {
	rater := getIPRate(ip)
	rater.Time = time.Now()
	return rater.Rate.Allow()
}
