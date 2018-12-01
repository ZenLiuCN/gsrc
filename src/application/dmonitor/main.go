package main

import (
	`crypto/tls`
	`dockerstats`
	`flag`
	`fmt`
	`github.com/mackerelio/go-osstat/cpu`
	`github.com/mackerelio/go-osstat/memory`
	`gopkg.in/gomail.v2`
	`gopkg.in/yaml.v2`
	`io/ioutil`
	`logger`
	`os`
	`regexp`
	`strconv`
	`strings`
	`time`
)

var config string
var conf *Config
var last = time.Time{}
var send = false
var log *logger.Logger

func main() {
	conf = new(Config)
	last = time.Now()
	flag.StringVar(&config, "conf", `dmonitor_conf.yml`, "config file path")
	flag.Parse()
	if len(config) == 0 {
		panic(fmt.Errorf(`config file not set`))
	}
	if e := conf.ParseFile(config); e != nil {
		panic(e)
	}

	logfile, cancel := logger.NewRotateWriter(conf.LogFile, logger.MustToBytes("1MB"), time.Second*5)
	defer cancel()
	if conf.Debug {
		logger.Init(logger.TRACE, os.Stdout, logfile)
	} else {
		logger.Init(logger.INFO, os.Stdout, logfile)
	}
	log = logger.GetLogger()
	if conf.Sleep.String() != "0s" {
		for {
			m := NewMonitor()
			for res := range m.Stream {
				if process(res) {
					continue
				}
				break
			}
			m.Stop()
			time.Sleep(conf.Sleep)
		}
	} else {
		m := NewMonitor()
		for res := range m.Stream {
			if process(res) {
				continue
			}
		}
	}
}
func process(res *dockerstats.StatsResult) bool {
	var content string
	if res.Error != nil {
		panic(res.Error)
	}
	if len(res.Stats) == 0 {
		log.Errorf("No Docker container is running, complete.")
		if last.Add(conf.Smtp.MaxFreq).Before(time.Now()) {
			msg := fmt.Sprintf(`No Docker container is running`)
			e := conf.SendMail(msg)
			if e != nil {
				log.Errorf("send email failed %v \n content:%s", e, msg)
				return true
			}
			last = time.Now()
		}
		//m.Stop()
		return true
	}
	send = false
	for _, s := range res.Stats {
		if v := conf.HasContainer(s.Container,s.Name); v != "" {
			msg := fmt.Sprintf("\ncontainer:%s<%s>", v, s.Name)
			if p, e := strconv.ParseFloat(strings.Replace(s.Memory.Percent, "%", "", -1), 0); e != nil {
				panic(e)
			} else if p >= conf.MaxMem {
				log.Warnf("container %s over mem use %f of %s", v, conf.MaxMem, s.Memory.Percent)
				msg += fmt.Sprintf(` mem over limit %f to %s`, conf.MaxMem, s.Memory.Percent, )
				log.Tracef("container %s over mem use %f of %s \n %s", v, conf.MaxMem, s.Memory.Percent, msg)
				send = true
			}
			if p, e := strconv.ParseFloat(strings.Replace(s.CPU, "%", "", -1), 0); e != nil {
				panic(e)

			} else if p > conf.MaxCPU {
				log.Warnf("container %s over cpu use %f of %s", v, conf.MaxCPU, s.CPU)
				msg += fmt.Sprintf(` cpu over limit %f to %s.`, conf.MaxCPU, s.CPU)
				log.Tracef("container %s over cpu use %f of %s \n %s", v, conf.MaxCPU, s.CPU, msg)
				send = true
			}
			if send && len(strings.Split(msg, ":")) > 1 && strings.Split(msg, ":")[1] != "" {
				content += msg
			}
			if conf.Log{
				log.Infoln(s)
			}
		}
		if conf.Debug {
			log.Infoln(s)
		}
	}
	if cpuStat := HostCpu(); len(cpuStat) != 0 {
		send = true
		content += cpuStat
		log.Warnln(cpuStat)
	}
	if memStat := HostMem(); len(memStat) != 0 {
		send = true
		content += memStat
		log.Warnln(memStat)
	}

	log.Tracef("send mail check %s %t \n %s", last.Add(conf.Smtp.MaxFreq), send, content)
	if send && last.Add(conf.Smtp.MaxFreq).Before(time.Now()) && len(content) != 0 {
		e := conf.SendMail(content)
		if e != nil {
			log.Errorf("send email failed %v \n content:%s", e, content)
			return true
		}
		log.Tracef("send email  content:%s", content)
		last = time.Now()
		send = false
	}
	return false
}

type Config struct {
	Container  []string
	Name       []string
	DockerPath string
	Smtp       Smtp
	MaxHostCPU float64
	MaxHostMem float64
	MaxCPU     float64
	MaxMem     float64
	Sleep      time.Duration
	LogFile    string
	Debug      bool
	Log      bool
}
type Smtp struct {
	Host    string
	Port    int
	User    string
	Pwd     string
	Emails  []string
	Header  string
	From    string
	SSL     bool
	MaxFreq time.Duration
}

func (c *Config) ParseFile(f string) error {
	d, e := ioutil.ReadFile(f)
	if e != nil {
		return e
	}
	e = yaml.UnmarshalStrict(d, c)
	if e != nil {
		return e
	}
	if len(c.Container) == 0 && len(c.Name) == 0 {
		return fmt.Errorf(`not container configuration %v`, c)
	}
	if len(c.LogFile) == 0 {
		c.LogFile = `/var/log/docker-monitor.log`
	}
	if c.MaxCPU <= 0 || c.MaxCPU > 100 {
		return fmt.Errorf(` max cpu percent not valid %v`, c)
	}
	if c.MaxMem <= 0 || c.MaxMem > 100 {
		return fmt.Errorf(` max mem percent not valid %v`, c)
	}
	if c.MaxHostCPU == 0 {
		c.MaxHostCPU = 98.0
	}
	if c.MaxHostMem == 0 {
		c.MaxHostMem = 98.0
	}
	if c.MaxHostCPU < 0 || c.MaxHostCPU > 100 {
		return fmt.Errorf(` max host cpu percent not valid %v`, c)
	}
	if c.MaxHostMem < 0 || c.MaxHostMem > 100 {
		return fmt.Errorf(` max host mem percent not valid %v`, c)
	}
	if c.Smtp.MaxFreq.String() == "0s" {
		c.Smtp.MaxFreq = time.Minute * 15
	}
	if len(c.Smtp.Emails) == 0 {
		return fmt.Errorf(` did not config warning emails %v`, c)
	}
	if len(c.Smtp.Header) == 0 {
		c.Smtp.Header = `Docker Warning`
	}
	if len(c.Smtp.Host) == 0 {
		return fmt.Errorf(` did not config stmp host %v`, c)
	}
	if c.Smtp.Port == 0 {
		return fmt.Errorf(` did not config stmp port %v`, c)
	}
	if c.DockerPath != "" {
		cf, ok := dockerstats.DefaultCommunicator.(dockerstats.CliCommunicator)
		//println(fmt.Sprintf(`%+v %t`, cf, ok))
		if ok {
			cf.DockerPath = c.DockerPath
		}
		//println(fmt.Sprintf(`%+v %t`, cf, ok))
	}
	if h, e := os.Hostname(); e == nil {
		c.Smtp.Header = fmt.Sprintf(`Host %s %s`, h, c.Smtp.Header)
	}
	return nil
}
func (c *Config) HasContainer(f string,n string) string {
	for _, v := range c.Container {
		if v == f {
			return v
		}
	}
	for _, v := range c.Name {
		ok, _ := regexp.MatchString(v, n)
		if ok {
			return f
		}
	}
	return ""
}
func (c *Config) SendMail(content string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", c.Smtp.From)
	m.SetHeader("To", c.Smtp.Emails...)
	m.SetHeader("Subject", c.Smtp.Header)
	m.SetBody("text/plain", content)
	d := gomail.Dialer{Host: c.Smtp.Host, Port: c.Smtp.Port, Username: c.Smtp.User, Password: c.Smtp.Pwd, SSL: c.Smtp.SSL}
	if c.Smtp.SSL {
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	if err := d.DialAndSend(m); err != nil {
		log.Errorf("Send Mail failed %v", err)
		return err
	}
	log.Infof("send mail done %s", content)
	return nil
}

func HostCpu() string {
	before, err := cpu.Get()
	if err != nil {
		log.Errorf("%s\n", err)
		return ""
	}
	time.Sleep(time.Duration(1) * time.Second)
	after, err := cpu.Get()
	if err != nil {
		log.Errorf("%s\n", err)
		return ""
	}
	total := float64(after.Total - before.Total)
	cpuRate := float64(after.User-before.User) / total * 100
	if cpuRate >= conf.MaxHostCPU {
		return fmt.Sprintf("\n host cpu usage %f%% over %f %%", cpuRate, conf.MaxHostCPU)
	}
	log.Infof("host cpu used %f%%", cpuRate)
	return ""
}
func HostMem() string {
	before, err := memory.Get()
	if err != nil {
		log.Errorf("%s\n", err)
		return ""
	}
	cpuRate := float64(before.Used) / float64(before.Total) * 100
	if cpuRate >= conf.MaxHostMem {
		return fmt.Sprintf("\n host mem usage %f%% over %f %%", cpuRate, conf.MaxHostMem)
	}
	log.Infof("host mem used %f%% total: %s,used:%s,free:%s", cpuRate, logger.BytesToString(before.Total), logger.BytesToString(before.Used), logger.BytesToString(before.Free))
	return ""
}

func NewMonitor() *dockerstats.Monitor {
	var com = dockerstats.DefaultCommunicator
	if conf.DockerPath != "" {
		cf, ok := dockerstats.DefaultCommunicator.(dockerstats.CliCommunicator)
		//println(fmt.Sprintf(`%+v %t`, cf, ok))
		if ok {
			cf.DockerPath = conf.DockerPath
			com = cf
		}
		//println(fmt.Sprintf(`%+v %t`, cf, ok))
	}
	m := dockerstats.Monitor{
		Stream: make(chan *dockerstats.StatsResult),
		Comm:   com,
	}
	m.Start()

	return &m
}
