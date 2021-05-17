// Copyright By git@clearcode.cn
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package trade

import (
	"github.com/adshao/go-binance/v2"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"time"
)

type Trade struct {
	option Option

	client *binance.Client
}

func NewTrade(opts ...Options) *Trade {
	var option Option
	for _, o := range opts {
		o(&option)
	}

	if option.SystemOption.AccessKey == "" || option.SystemOption.SecretKey == "" {
		panic("accessKey or secretKey can't be empty")
	}

	t := &Trade{}
	t.option = option

	{
		var proxyURL *url.URL
		var err error
		if option.SystemOption.ProxyURL != "" {
			proxyURL, err = url.Parse(option.SystemOption.ProxyURL)
			if err != nil {
				logrus.WithError(err).Error("failed to parse proxy url, will not use proxy")
			}
		}

		var client = &http.Client{
			Transport: http.DefaultTransport,
			Timeout:   30 * time.Second,
		}

		if proxyURL != nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}

		bclient := binance.NewClient(option.SystemOption.AccessKey, option.SystemOption.SecretKey)
		bclient.HTTPClient = client
		bclient.Debug = option.SystemOption.Debug

		t.client = bclient
	}

	t.init()

	return t
}

func (t *Trade) init() {

}
