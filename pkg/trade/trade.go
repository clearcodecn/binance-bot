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
	"context"
	"encoding/json"
	"errors"
	"github.com/adshao/go-binance/v2"
	lru "github.com/hashicorp/golang-lru"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

type Trade struct {
	mu     sync.RWMutex
	option Option

	client *binance.Client

	closers []io.Closer

	logger logrus.FieldLogger

	boughtMutex sync.Mutex
	boughtInfo  map[string]*BoughtInfo

	sellChan chan *SellBill

	buyChan chan []*symbolPriceChange

	cacheMutex  sync.Mutex
	boughtCache *lru.Cache
}

// NewTrade returns Trade object.
func NewTrade(opts ...Options) *Trade {
	var option Option
	for _, o := range opts {
		o(&option)
	}

	cache, _ := lru.New(1024)

	t := &Trade{}
	t.boughtCache = cache
	t.option = option
	t.boughtInfo = make(map[string]*BoughtInfo)
	t.sellChan = make(chan *SellBill, 60)
	t.SetSystemOption(option.SystemOption)
	t.init()

	return t
}

// SetAkSk goroutine unsafe
func (t *Trade) SetSystemOption(option SystemOption) {
	t.option.SystemOption = option

	var proxyURL *url.URL
	var err error
	if t.option.SystemOption.ProxyURL != "" {
		proxyURL, err = url.Parse(t.option.SystemOption.ProxyURL)
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

	bclient := binance.NewClient(t.option.SystemOption.AccessKey, t.option.SystemOption.SecretKey)
	bclient.HTTPClient = client
	bclient.Debug = t.option.SystemOption.Debug

	t.client = bclient
}

func (t *Trade) init() {
	// Logger
	l := logrus.New()
	l.ReportCaller = true
	t.logger = l

	if t.option.SystemOption.LogFile != "" {
		fi, err := os.OpenFile(t.option.SystemOption.LogFile, os.O_CREATE|os.O_APPEND, 0777)
		if err != nil {
			panic(err)
			return
		}
		t.closers = append(t.closers, fi)
		l.SetOutput(fi)
	}
	if t.option.SystemOption.Debug {
		l.SetLevel(logrus.DebugLevel)
	}

	// Reset Bought Info
	if t.option.BuyOption.BoughtFile != "" {
		data, err := ioutil.ReadFile(t.option.BuyOption.BoughtFile)
		if err != nil {
			if !os.IsNotExist(err) {
				panic(err)
			}
		}
		json.Unmarshal(data, &t.boughtInfo)
	}
}

func (t *Trade) Run(stopChan chan struct{}) error {
	if t.client == nil {
		return errors.New("client not init")
	}

	ctx, cancel := context.WithCancel(context.Background())

	go t.runBuy(ctx)
	go t.runSell(ctx)

	go func() {
		<-stopChan
		cancel()

	}()

	return nil
}

func (t *Trade) Close() error {
	for _, c := range t.closers {
		if err := c.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (t *Trade) runSell(ctx context.Context) {

}

func (t *Trade) getBoughtInfo() map[string]*BoughtInfo {
	var info = make(map[string]*BoughtInfo)
	for k, v := range t.boughtInfo {
		i := BoughtInfo{}
		i = *v
		info[k] = &i
	}
	return info
}
