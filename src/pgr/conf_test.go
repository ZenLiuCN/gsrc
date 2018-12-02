package pgr

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"testing"
)

func TestConf(t *testing.T) {
	c:=new(Conf)
	d,_:=yaml.Marshal(c)
	_ = ioutil.WriteFile("./test.yml", d, 0666)
}