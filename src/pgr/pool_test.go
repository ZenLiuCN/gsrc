package pgr

import (
	"logger"
	"testing"
)

const TestYaml  =`
host: "192.168.99.100"
port: 65432
database: "static"
user: "zen"
password: "zen"
usefallbacktls: false
loglevel: 0
runtimeparams: {}
prefersimpleprotocol: false
maxconnections: 50
acquiretimeout: 5s
`

func TestInitFromYaml(t *testing.T) {
	logger.Init(logger.DEBUG)
	SetLogger(logger.GetLogger())
	p,e:=InitFromYaml([]byte(TestYaml))
	if e!=nil{
		t.Fatal(e)
	}
	t.Logf("%+v",p.Stat())
	t.Log(p.Exec(`create table if not exists abc(a integer,b text)`))
	t.Log(p.Exec(`insert into abc values(1,'a',$1)`, map[string]interface{}{
		`a`:"123",`b`:1,
	}))
	var m =make(map[string]interface{})
	t.Log(p.QueryRow(`SELECT c from abc where a=$1`,1).Scan(&m))
	t.Log(m)
}