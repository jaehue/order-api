package adapters

import (
	"context"
	"time"

	try "github.com/matryer/try"
	"github.com/pangpanglabs/goutils/behaviorlog"
	"github.com/pangpanglabs/goutils/httpreq"
	"github.com/sirupsen/logrus"
)

var client = httpreq.NewClient(httpreq.ClientConfig{
	Timeout: 10 * time.Second,
})

const attempt = 5

func RetryRestApi(ctx context.Context, resp interface{}, method string, url string, param interface{}) error {
	logrus.WithField("Url", url).Info("Call Api")
	err := try.Do(func(attempt int) (bool, error) {
		_, err := httpreq.New(method, url, param).
			WithBehaviorLogContext(behaviorlog.FromCtx(ctx)).
			CallWithClient(&resp, client)

		logrus.WithField("attempt", attempt).Info("attempt")

		if err != nil {
			time.Sleep(2 * time.Second) // wait a Second
		}

		return attempt < 5, err
	})

	return err
}
