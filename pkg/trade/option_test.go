package trade

import (
	"github.com/ghodss/yaml"
	"io/ioutil"
	"testing"
)

func TestWithDefaultSellOption(t *testing.T) {
	var o = new(Option)
	o.BuyOption = DefaultBuyOption
	o.SellOption = DefaultSellOption

	data,_ := yaml.Marshal(o)
	ioutil.WriteFile("config.yaml",data,0777)
}
