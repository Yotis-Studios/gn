package gn

import (
	"bytes"
	"encoding/binary"
)

type Packet struct {
	netID uint16
	data  []interface{}
	raw   []byte
}

func NewPacket(netID uint16) *Packet {
	p := new(Packet)
	p.netID = netID
	p.data = make([]interface{}, 0)
	return p
}

func (p *Packet) Add(data interface{}) {
	// type massaging
	switch data.(type) {
	case bool:
		if data.(bool) {
			data = 1
		} else {
			data = 0
		}
	}
	p.data = append(p.data, data)
}

func (p Packet) Get(index int) interface{} {
	return p.data[index]
}

func (p Packet) Build() ([]byte, error) {
	if p.raw != nil {
		return p.raw, nil
	}
	var size = 0
	var buf = new(bytes.Buffer)
	// write packet id
	err := binary.Write(buf, binary.LittleEndian, p.netID)
	if err != nil {
		return nil, err
	}
	size += 2

	// serialize and write data
	for _, data := range p.data {
		var b []byte
		b, err = BytesFromData(data)
		if err != nil {
			return nil, err
		}
		buf.Write(b)
		size += len(b)
	}

	// write packet size and packet to new buffer
	var pBuf = new(bytes.Buffer)
	err = binary.Write(pBuf, binary.LittleEndian, uint16(size))
	if err != nil {
		return nil, err
	}
	pBuf.Write(buf.Bytes())

	p.raw = pBuf.Bytes()

	return p.raw, nil
}

func Load(b []byte) (*Packet, error) {
	p := new(Packet)
	p.raw = b
	r := bytes.NewReader(b)
	var pSize uint16
	err := binary.Read(r, binary.LittleEndian, &pSize)
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, binary.LittleEndian, &p.netID)
	if err != nil {
		return nil, err
	}

	n := len(b)
	if n > 4 {
		j := 4
		for j < n {
			value, size, parseErr := Parse(r)
			if parseErr != nil {
				return nil, parseErr
			}
			p.data = append(p.data, value)
			j += size + 1
		}
	}

	return p, nil
}
