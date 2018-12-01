package hide

import (
	`encoding/hex`
	`fmt`
	`hide/asyncio`
	`log`
	`strconv`
	`strings`
	`sync`
	`testing`
	`time`
)

var (
	S_CLOSE_LIGHT      = []byte{0x55, 0xAA, 0x24, 0x01, 0x00, 0x00, 0xDA}
	S_OPEN_LIGHT_WHITE = []byte{0x55, 0xAA, 0x24, 0x01, 0x00, 0x01, 0xDB}
	C_VOICE_0          = []byte{0x55, 0xAA, 0x29, 0x01, 0x00, 0x00, 0xD7}
	C_VOICE_1          = []byte{0x55, 0xAA, 0x29, 0x01, 0x00, 0x01, 0xD6}
	C_VOICE_2          = []byte{0x55, 0xAA, 0x29, 0x01, 0x00, 0x02, 0xD5}
	C_VOICE_3          = []byte{0x55, 0xAA, 0x29, 0x01, 0x00, 0x03, 0xD4}
	C_VOICE_4          = []byte{0x55, 0xAA, 0x29, 0x01, 0x00, 0x04, 0xD3}
	devicePath         = []string{
		`\\?\hid#vid_0525&pid_a4ac&mi_00#7&2033f336&0&0000#{4d1e55b2-f16f-11cf-88cb-001111000030}`,
		`\\?\hid#vid_0525&pid_a4ac&mi_00#7&21d28301&0&0000#{4d1e55b2-f16f-11cf-88cb-001111000030}`,
		`\\?\hid#vid_0525&pid_a4ac&mi_01#7&3a34a8c&0&0000#{4d1e55b2-f16f-11cf-88cb-001111000030}`,
		`\\?\hid#vid_0525&pid_a4ac&mi_01#7&785e128&0&0000#{4d1e55b2-f16f-11cf-88cb-001111000030}`,
	}
	byte1 = []byte{
		0x55, 0xAA, 0x22, 0x03, 0x00, 0x03, 0x02, 0x00, 0xDF, 0x0F, 0x85, 0x75, 0x20, 0x02, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x8C, 0xFE, 0x15, 0x06, 0x75, 0xED, 0x4B, 0x00, 0x20, 0x02, 0x00, 0x00,
		0x4E, 0xDD, 0x4B, 0x00, 0x70, 0x21, 0xC3, 0x22, 0x40, 0x00, 0x00, 0x00, 0x10, 0xF8, 0x3C, 0x00,
		0x7F, 0x15, 0xC8, 0x67, 0xD5, 0x8C, 0x8E, 0x75, 0x36, 0x1E, 0xBA, 0xC7, 0xFE, 0xFF, 0xFF, 0xFF,
	}
)

func TestNames(t *testing.T) {
	d, e := Names()
	log.Printf("error %v\n", e)
	for _, v := range d {
		log.Printf("d: %+v \n", v)
	}
}
func TestOpen(t *testing.T) {
	for _, v := range devicePath {
		if _, e := Open(v); e == nil {
			log.Printf(`open success of %s \n`, v)
		}
	}
}
func TestStat(t *testing.T) {
	for _, v := range devicePath {
		if s, e := Stat(v); e == nil {
			log.Printf(`Stat of %s
Caps ilen %d
Caps outLen %d
Attr SerialNO %s
`, v,
				s.Caps.InputLen,
				s.Caps.OutputLen,
				s.Attr.SerialNo,
			)
		}
	}
}
func TestStat2(t *testing.T) {
	if s, e := Stat(devicePath[2]); e == nil {
		log.Printf(
			`Stat
Caps ilen %d
Caps outLen %d
Attr SerialNO %s
Attr Product %X
Attr VendorId %X
Attr Version %X
Name %s
`,
			s.Caps.InputLen,
			s.Caps.OutputLen,
			s.Attr.SerialNo,
			s.Attr.ProductId,
			s.Attr.VendorId,
			s.Attr.Version,
			s.Name,
		)
	}

}

func TestSerialNo(t *testing.T) {
	d, e := Open(devicePath[3])
	if e != nil {
		t.Logf(`Error %v
`, e)
	}
	t.Logf("%T", d)
	bytes := make(chan []byte)
	go func() {
		b := make([]byte, 0, 1025)
		for {
			time.Sleep(200 * time.Millisecond)
			i, e := d.Read(b)
			if e != nil {
				t.Logf(`read error %T %v`, e, e)
				continue
			}
			if i == 0 {
				continue
			}
			bytes <- b[:i]
		}
	}()
	a := <-bytes
	t.Logf(`%s`, a)
}

func TestVendorDevices(t *testing.T) {
	if v, e := VendorDevices(uint16(0X0525), uint16(0XA4AC)); e != nil {
		t.Logf(`devices %+v`, v)
		for _, d := range v {
			t.Logf(`Info 
Name: %s
Info: %+v
`, d.Name, d,
			)
		}
	} else if e != nil {
		t.Logf(`Error %v`, e)
	}
}

func TestRead(t *testing.T) {
	d, e := Open(devicePath[2])
	if e != nil {
		t.Logf(`Error %v
`, e)
	}
	bytes := make([]byte, 1025)
	i, e := d.Read(bytes)
	if e != nil {
		t.Logf(`	Read Error %v %d
`, e, i)
	}
	t.Logf(`Read data %v
`, hex.Dump(bytes))
}
func TestWriter(t *testing.T) {
	d, e := Open(devicePath[2])
	if e != nil {
		t.Logf(`Error %v
`, e)
	}
	var i int
	i, e = d.Write(GetBytes(&C_VOICE_2))
	if e != nil {
		t.Logf(`Error %v %d
`, e, i)
	}
	b := GetBytes(nil)
	i, e = d.Read(b)
	if e != nil {
		t.Errorf(`Error %v `, e)
	}
	t.Logf(`%+v`, hex.Dump(b))
}

func TestCallDevices(t *testing.T) {
	g := new(sync.WaitGroup)
	if e := callDevices(g); e != nil {
		t.Errorf(`Error send to all device %v`, e)
	}
	g.Wait()
}

func callDevices(group *sync.WaitGroup) error {
	d, e := Names()
	if e != nil {
		return e
	}
	var device []string
	for _, v := range d {
		if strings.Contains(v, `\\?\hid#vid_0525&pid_a4ac&mi_01#`) {
			device = append(device, v)
		}
	}
	log.Printf(`find devices %+v 
`, device)
	if len(device) == 0 {
		return nil
	}
	ds, e := openAll(device)
	if e != nil {
		return e
	}
	sendVoice1(ds, group)
	return nil
}
func openAll(dev []string) (d []*Device, e error) {
	for _, v := range dev {
		if ds, er := Open(v); er != nil {
			e = fmt.Errorf(`Error open %s %v `, v, er)
			return
		} else {
			d = append(d, ds)
		}
	}
	return
}
func sendVoice1(dev []*Device, group *sync.WaitGroup) {
	for i, d := range dev {
		d.SetTimeout(5 * time.Millisecond)
		group.Add(1)
		go func() {
			group.Done()
			d.Read(GetBytes(nil))
		}()
		if i%2 == 0 {
			_, e := d.Write(GetBytes(&C_VOICE_1))
			if e != nil {
				log.Fatalf(`Error send to %s`, d.Name())
			}
		} else {
			_, e := d.Write(GetBytes(&C_VOICE_2))
			if e != nil {
				log.Fatalf(`Error send to %s`, d.Name())
			}
		}
	}
}
func GetBytes(data *[]byte) []byte {
	b := make([]byte, 1025)
	if data == nil {
		return b
	}
	for i, v := range *data {
		b[i+1] = v
	}
	return b
}

func TestFindDevices(t *testing.T) {
	d, e := Names()
	if e != nil {
		t.Fatalf(`error %v`, e)
	}
	var device []string
	for _, v := range d {
		if strings.Contains(v, `\\?\hid#vid_0525&pid_a4ac&mi_01#`) {
			device = append(device, v)
		}
	}
	t.Logf(`find devices %+v 
`, device)
	if len(device) == 0 {
		t.Logf(`Not Device`)
	}
	t.Logf(`Get Device %v`, device)
	ds, e := openAll(device)
	if e != nil {
		t.Fatalf(`Error open device %v`, e)
	}
	if len(ds) == 0 {
		t.Fatalf(`Error no device %v`, e)
	}
	b := []bool{false, false, false, false}
	for {
		for key, value := range ds {
			if !b[key] {
				value.Write(S_OPEN_LIGHT_WHITE)
				b[key] = true
			}
			b := GetBytes(nil)
			c := strconv.Itoa(key)
			time.Sleep(200 * time.Millisecond)
			value.SetTimeout(200 * time.Millisecond)
			i, e := value.Read(b)
			if e != nil && e != asyncio.ErrTimeout {
				if e.Error() == `The device is not connected.` {
					return
				}
				t.Errorf("Error read from %s %v  %d \n", c, e, i)
				continue
			} else if e == asyncio.ErrTimeout {
				t.Logf(`Get Res of %v`, e)
				continue
			}
			switch {
			case i == 0:
				t.Logf(`Get Res of %d`, i)
				continue
			case b[1] == 0x55 && b[2] == 0xAA && b[4] == 0x00:
				l := uint32(b[5]) + uint32(b[6])<<8
				if l > 0 {
					t.Logf(`Get Res of %s`+"\n", c+`♦`+string(b[7:l+7]))
				}
			case i > 6:
				t.Logf("reading from device %v", hex.Dump(b[:7]))
			default:
				t.Logf("reading from device %v", hex.Dump(b))
			}

		}
	}
}

func TestReadString(t *testing.T) {
	d, e := Open(devicePath[2])
	if e != nil {
		t.Error(e)
	}
	d.Write(GetBytes(&S_OPEN_LIGHT_WHITE))
	b := GetBytes(nil)
	_, e = d.Read(b)
	if e != nil {
		t.Error(e)
	}
	t.Logf("%v \n %s", hex.Dump(b[1:7]),GetString(b))
	defer func() {
		time.Sleep(time.Millisecond * 200)
		d.Write(GetBytes(&S_CLOSE_LIGHT))
		b := GetBytes(nil)
		_, e := d.Read(b)
		if e != nil {
			t.Error(e)
		}
		t.Logf("%v \n %s", hex.Dump(b[1:7]),GetString(b))
	}()
	for{
		time.Sleep(time.Millisecond * 200)
		b = GetBytes(nil)
		_, e = d.Read(b)
		if e != nil {
			t.Error(e)
		}
		t.Logf("%v \n %s", hex.Dump(b[1:7]),GetString(b))
	}

}
func GetString(b []byte) string {
	if b[1] == 0x55 && b[2] == 0xAA && b[4] == 0x00 {
		l := uint32(b[5]) + uint32(b[6])<<8
		if l > 0 {
			return string(b[7 : l+7])
		}
	}
	return ""
}
