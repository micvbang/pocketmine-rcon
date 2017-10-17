package rcon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"net"
)

type Connection struct {
	conn net.Conn
	pass string
	addr string
}

var uniqueId int32 = 0

func NewConnection(addr, pass string) (*Connection, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
// 		log.Fatal(err)
		return nil, err
	}
	c := &Connection{conn: conn, pass: pass, addr: addr}
	if err := c.auth(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Connection) SendCommand(cmd string) string {
	c.sendCommand(2, []byte(cmd))
	pkg := c.readPkg()
	return string(pkg.Body)
}

func (c *Connection) auth() error {
	c.sendCommand(3, []byte(c.pass))
	pkg := c.readPkg()
	if pkg.Type != 2 || pkg.Id != uniqueId {
		return errors.New("Incorrect password.")
	}
	return nil
}

func (c *Connection) sendCommand(typ int32, body []byte) {
	size := int32(4 + 4 + len(body) + 2)
	uniqueId += 1
	id := uniqueId

	wtr := binaryReadWriter{ByteOrder: binary.LittleEndian}
	wtr.Write(size)
	wtr.Write(id)
	wtr.Write(typ)
	wtr.Write(body)
	wtr.Write([]byte{0x0, 0x0})
	if wtr.err != nil {
		log.Fatal(wtr.err)
	}

	c.conn.Write(wtr.buf.Bytes())
}

func (c *Connection) readPkg() Pkg {
	const bufSize = 4096
	b := make([]byte, bufSize)

	// Doesn't handle split messages correctly.
	read, err := c.conn.Read(b)
	if err != nil {
		log.Fatal(err)
	}

	p := Pkg{}
	rdr := binaryReadWriter{ByteOrder: binary.LittleEndian,
		buf: bytes.NewBuffer(b)}
	rdr.Read(&p.Size)
	rdr.Read(&p.Id)
	rdr.Read(&p.Type)
	body := [bufSize - 12]byte{}
	rdr.Read(&body)
	if rdr.err != nil {
		log.Fatal(rdr.err)
	}
	p.Body = body[:read-12]
	return p
}

type Pkg struct {
	Size int32
	Id   int32
	Type int32
	Body []byte
}

type binaryReadWriter struct {
	ByteOrder binary.ByteOrder
	err       error
	buf       *bytes.Buffer
}

func (b *binaryReadWriter) Write(v interface{}) {
	if b.err != nil {
		return
	}
	if b.buf == nil {
		b.buf = new(bytes.Buffer)
	}
	b.err = binary.Write(b.buf, b.ByteOrder, v)
}

func (b *binaryReadWriter) Read(v interface{}) {
	if b.err != nil || b.buf == nil {
		return
	}
	b.err = binary.Read(b.buf, b.ByteOrder, v)
}
