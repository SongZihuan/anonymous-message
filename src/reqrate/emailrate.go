package reqrate

import (
	"fmt"
	"github.com/emersion/go-imap/v2"
	"golang.org/x/time/rate"
	"net/mail"
	"sync"
	"time"
)

// RateExp: 计算周期
// RateMaxCount: 周期内最大发件数量
const (
	userEmailRateExp      time.Duration = 1 * time.Hour
	userEmailRateMaxCount int           = 18
)

type UserEmail any

type userEmailRate struct {
	UserEmail string
	Rate      *rate.Limiter
	Time      time.Time
	RateExp   time.Duration
}

var userEmailRaterMap sync.Map

func getUserEmailRate(uesrMail UserEmail) (rater *userEmailRate) {
	rater = &userEmailRate{
		UserEmail: _getUserEmailName(uesrMail),
		Rate:      rate.NewLimiter(rate.Every(userEmailRateExp), userEmailRateMaxCount),
		Time:      time.Now(),
		RateExp:   userEmailRateExp,
	}

	raterInterface, ok := userEmailRaterMap.LoadOrStore(rater.GetName(), rater)
	rater, ok = raterInterface.(*userEmailRate)
	if !ok {
		panic("sync.map error")
	}

	return rater
}

func (a *userEmailRate) GetName() string {
	return _getUserEmailName(a.UserEmail)
}

func _getUserEmailName(userEmail UserEmail) string {
	return fmt.Sprintf("%s", _getUserEmail(userEmail))
}

func _getUserEmail(addr UserEmail) string {
	switch a := addr.(type) {
	case *mail.Address:
		return a.Address
	case mail.Address:
		return a.Address
	case *imap.Address:
		return a.Addr()
	case imap.Address:
		return a.Addr()
	case string:
		return a
	default:
		panic("not a valid address")
	}
}

func CleanUserEmailRate() {
	for range time.Tick(1 * time.Minute) {
		func() {
			defer func() {
				_ = recover()
			}()

			now := time.Now()
			delList := make([]any, 0, 10)

			userEmailRaterMap.Range(func(key, value any) bool {
				rater, ok := value.(*userEmailRate)
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
				_, _ = userEmailRaterMap.LoadAndDelete(v)
			}
		}()
	}
}

func CheckIMAPRate(envelope *imap.Envelope) bool {
	addressList := make([]imap.Address, 0, len(envelope.Sender)+len(envelope.From)+len(envelope.ReplyTo))
	addressList = append(addressList, envelope.Sender...)
	addressList = append(addressList, envelope.From...)
	addressList = append(addressList, envelope.ReplyTo...)

	return checkMailAddressListRate(addressList)
}

func checkMailAddressListRate(addressList []imap.Address) bool {
	var addressMap = make(map[string]bool, len(addressList))

	for _, address := range addressList {
		if yes, ok := addressMap[address.Addr()]; ok && yes { // 去除重复
			continue
		}

		if !CheckMailAddressRate(address) {
			return false
		}

		addressMap[address.Addr()] = true
	}

	return true
}

func CheckMailAddressListRate(addressList []*imap.Address) bool {
	var addressMap = make(map[string]bool, len(addressList))

	for _, address := range addressList {
		if yes, ok := addressMap[address.Addr()]; ok && yes {
			continue
		}

		if !CheckMailAddressRate(address) {
			return false
		}

		addressMap[address.Addr()] = true
	}

	return true
}

func CheckMailAddressRate(userEmail UserEmail) bool {
	rater := getUserEmailRate(userEmail)
	rater.Time = time.Now()
	return rater.Rate.Allow()
}
