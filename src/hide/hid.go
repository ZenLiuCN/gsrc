package hide

import (
	`hide/asyncio`
	`fmt`
	`syscall`
)

// Names lists the names of all available HID devices.
func Names() ([]string, error) {
	return FindDevices()
}

// VendorDevices finds accessible devices having the specified vendor and product IDs.
func VendorDevices(vendor uint16, products ...uint16) ([]*DeviceInfo, error) {
	v, err := Names()
	if err != nil {
		return nil, err
	}
	var vv []*DeviceInfo
	for _, n := range v {
		i, err := Stat(n)
		if err != nil {
			if IsAccess(err) {
				continue
			}
			return nil, err
		}
		if i.Attr.VendorId != vendor {
			continue
		}
		for _, iv := range products {
			if iv == i.Attr.ProductId {
				vv = append(vv, i)
				break
			}
		}
	}
	return vv, nil
}

// SerialNo finds accessible devices having the specified serial number.
func SerialNo(sno string) ([]*DeviceInfo, error) {
	v, err := Names()
	if err != nil {
		return nil, err
	}
	var vv []*DeviceInfo
	for _, n := range v {
		i, err := Stat(n)
		if err != nil {
			if IsAccess(err) {
				continue
			}
			return nil, err
		}
		if i.Attr.SerialNo == sno {
			vv = append(vv, i)
		}
	}
	return vv, nil
}

// Stat returns device info from the specified path.
func Stat(name string) (*DeviceInfo, error) {
	d, err := Open(name)
	if err != nil {
		return nil, err
	}
	defer d.Close()
	return d.DeviceInfo()
}

// Open opens the specified device.
func Open(name string) (*Device, error) {
	f, err := asyncio.Open(name)
	if err != nil {
		return nil, newErr("Open", name, err)
	}
	return &Device{f}, nil
}

// Device is a HID device that statisfies io.ReadWriteCloser.
type Device struct {
	*asyncio.File
}

type DeviceInfo struct {
	Name string

	Attr *Attr
	Caps *Caps
}

type Attr struct {
	VendorId  uint16
	ProductId uint16
	Version   uint16
	SerialNo  string
}

type Caps struct {
	Usage     uint16
	UsagePage uint16

	// Report lengths
	InputLen   int
	OutputLen  int
	FeatureLen int

	NumLinkCollectionNodes int
	NumInputButtonCaps     int
	NumInputValueCaps      int
	NumInputDataIndices    int
	NumOutputButtonCaps    int
	NumOutputValueCaps     int
	NumOutputDataIndices   int
	NumFeatureButtonCaps   int
	NumFeatureValueCaps    int
	NumFeatureDataIndices  int
}

type Error struct {
	Func string
	Path string
	Err  error
}

func (e *Error) Error() string {
	return e.Func + " " + e.Path + ": " + e.Err.Error()
}

func newErr(f, p string, err error) error {
	return &Error{f, p, err}
}

func (d *Device) DeviceInfo() (*DeviceInfo, error) {
	i := &DeviceInfo{Name: d.Name()}
	err := statHandle(syscall.Handle(d.Fd()), i)
	if err != nil {
		return nil, newErr("ds4.DeviceInfo", d.Name(), err)
	}
	return i, nil
}

func (d *Device) SetOutputReport(p []byte) error {
	return HidD_SetOutputReport(
		syscall.Handle(d.Fd()),
		&p[0],
		uint32(len(p)))
}

/*// Disconnect device radio, assuming it's using bluetooth.
func (d *Device) DisconnectRadio() error {
	di, err := d.DeviceInfo()
	if err != nil {
		return err
	}
	return DisconnectBluetooth(di.Attr.SerialNo)
}*/

func statHandle(h syscall.Handle, d *DeviceInfo) error {

	var attr HIDD_ATTRIBUTES
	if err := HidD_GetAttributes(h, &attr); err != nil {
		return err
	}

	d.Attr = &Attr{
		VendorId:  attr.VendorID,
		ProductId: attr.ProductID,
		Version:   attr.VersionNumber,
		SerialNo:  GetSerialNo(h),
	}

	var prepd uintptr
	if err := HidD_GetParsedData(h, &prepd); err != nil {
		return err
	}
	defer HidD_FreePreparsedData(prepd)

	var caps HIDP_CAPS
	if errc := HidP_GetCaps(prepd, &caps); errc != HIDP_STATUS_SUCCESS {
		return fmt.Errorf("hid.GetCaps() failed with error code %#x", errc)
	}

	d.Caps = &Caps{
		Usage:     caps.Usage,
		UsagePage: caps.UsagePage,

		InputLen:   int(caps.InputReportByteLength),
		OutputLen:  int(caps.OutputReportByteLength),
		FeatureLen: int(caps.FeatureReportByteLength),

		NumLinkCollectionNodes: int(caps.NumberLinkCollectionNodes),
		NumInputButtonCaps:     int(caps.NumberInputButtonCaps),
		NumInputValueCaps:      int(caps.NumberInputValueCaps),
		NumInputDataIndices:    int(caps.NumberInputDataIndices),
		NumOutputButtonCaps:    int(caps.NumberOutputButtonCaps),
		NumOutputValueCaps:     int(caps.NumberOutputValueCaps),
		NumOutputDataIndices:   int(caps.NumberOutputDataIndices),
		NumFeatureButtonCaps:   int(caps.NumberFeatureButtonCaps),
		NumFeatureValueCaps:    int(caps.NumberFeatureValueCaps),
		NumFeatureDataIndices:  int(caps.NumberFeatureDataIndices),
	}

	return nil
}
