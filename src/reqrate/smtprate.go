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
	smtpRateExp      time.Duration = 12 * time.Hour
	smtpRateMaxCount int           = 3
)

type SMTPSendType string

const (
	SMTPSendTypeToSelf SMTPSendType = "ToSelf"
	SMTPSendTypeError  SMTPSendType = "Error"
	SMTPSendTypeThank  SMTPSendType = "Thank"
)

type Address any

type addressRate struct {
	Type    SMTPSendType
	Address string
	Rate    *rate.Limiter
	Time    time.Time
	RateExp time.Duration
}

var smtpRaterMap sync.Map

func getAddressRate(sendType SMTPSendType, address Address) (rater *addressRate) {
	rater = &addressRate{
		Type:    sendType,
		Address: _getAddress(address),
		Rate:    rate.NewLimiter(rate.Every(smtpRateExp), smtpRateMaxCount),
		Time:    time.Now(),
		RateExp: smtpRateExp,
	}

	raterInterface, ok := smtpRaterMap.LoadOrStore(rater.GetName(), rater)
	rater, ok = raterInterface.(*addressRate)
	if !ok {
		panic("sync.map error")
	}

	return rater
}

func (a *addressRate) GetName() string {
	return _getSMTPAddressName(a.Type, a.Address)
}

func _getSMTPAddressName(sendType SMTPSendType, address Address) string {
	return fmt.Sprintf("%s::%s", sendType, _getAddress(address))
}

func _getAddress(addr Address) string {
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

func CleanAddressRate() {
	for range time.Tick(1 * time.Minute) {
		func() {
			defer func() {
				_ = recover()
			}()

			now := time.Now()
			delList := make([]any, 0, 10)

			smtpRaterMap.Range(func(key, value any) bool {
				rater, ok := value.(*addressRate)
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
				_, _ = smtpRaterMap.LoadAndDelete(v)
			}
		}()
	}
}

func CheckSMTPSendAddressRate(sendType SMTPSendType, address Address) bool {
	rater := getAddressRate(sendType, address)
	rater.Time = time.Now()
	return rater.Rate.Allow()
}
