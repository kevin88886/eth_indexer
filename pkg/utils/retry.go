package utils

import (
	"context"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

func WithRetryCount(count int, duration, maxErrDuration time.Duration, fn func() error) func() error {
	return func() error {

		var (
			err       error
			lastErrAt = time.Now()
		)

		for i := 0; i < count; i++ {
			log.Infof("运行服务. count: %d", i)
			err = fn()
			// 正常关闭, 直接退出
			if err == nil || errors.Is(err, context.Canceled) {
				return err
			}

			errAt := time.Now()
			errDuration := errAt.Sub(lastErrAt)
			// 如果上一次发生错误的时间离当前时间已经超过了错误周期, 则认为这中间是正常运行, 需要重置计数
			if errDuration > maxErrDuration {
				i = 0
			}

			log.Errorf(
				"服务运行异常. count: %d. error: %v, lastErrAt: %s, now: %s, errDuration: %s, maxErrDuration: %s",
				i, err, lastErrAt.Format(time.DateTime), errAt.Format(time.DateTime), errDuration, maxErrDuration,
			)

			lastErrAt = errAt
			time.Sleep(duration)
		}

		return err
	}
}
