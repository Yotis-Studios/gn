package gn

import (
	"bytes"
	"encoding/binary"
)

type Packet struct {
	netID uint16
	data  []interface{}
}

func NewPacket(netID uint16) *Packet {
	p := new(Packet)
	p.netID = netID
	p.data = make([]interface{}, 0)
	return p
}

func (p Packet) Add(data interface{}) {
	p.data = append(p.data, data)
}

func (p Packet) Get(index int) interface{} {
	return p.data[index]
}

func (p Packet) Build() []byte {
	var size = 0
	var buf = new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p.netID)
	if err != nil {
		panic(err)
	}
	size += 2

	for _, data := range p.data {
		b := BytesFromData(data)
		buf.Write(b)
		size += len(b)
	}

	var pBuf = new(bytes.Buffer)
	err = binary.Write(pBuf, binary.LittleEndian, uint16(size))
	if err != nil {
		panic(err)
	}
	pBuf.Write(buf.Bytes())

	return pBuf.Bytes()
}

func Load(b []byte) *Packet {
	p := new(Packet)
	r := bytes.NewReader(b)
	var pSize uint16
	err := binary.Read(r, binary.LittleEndian, &pSize)
	if err != nil {
		panic(err)
	}
	err = binary.Read(r, binary.LittleEndian, &p.netID)
	if err != nil {
		panic(err)
	}

	n := len(b)
	if n > 4 {
		j := 4
		for j < n {
			value, size := Parse(r, j)
			p.data = append(p.data, value)
			j += size + 1
		}
	}

	return p
}
