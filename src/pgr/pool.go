package pgr

import (
	"github.com/jackc/pgx"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func InitFromFile(file string) (p *pgx.ConnPool, e error) {
	c := new(Conf)
	var b []byte
	b, e = ioutil.ReadFile(file)
	if e != nil {
		return nil, e
	}
	e = yaml.Unmarshal(b, c)
	if e != nil {
		return nil, e
	}
	return c.CreatePool()
}
func InitFromYaml(yamldata []byte) (p *pgx.ConnPool, e error) {
	c := new(Conf)
	e = yaml.Unmarshal(yamldata, c)
	if e != nil {
		return nil, e
	}
	return c.CreatePool()
}