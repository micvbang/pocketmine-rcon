package rcon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
)

type Connection struct {
	conn net.Conn
	pass string
	addr string
}

var uniqueID int32 = 0

func NewConnection(addr, pass string) (*Connection, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	c := &Connection{conn: conn, pass: pass, addr: addr}
	if err := c.auth(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Connection) SendCommand(cmd string) (string, error) {
	err := c.sendCommand(2, []byte(cmd))
	if err != nil {
		return "", err
	}
	pkg, err := c.readPkg()
	if err != nil {
		return "", err
	}
	return string(pkg.Body), err
}

func (c *Connection) auth() error {
	c.sendCommand(3, []byte(c.pass))
	pkg, err := c.readPkg()
	if err != nil {
		return err
	}

	if pkg.Type != 2 || pkg.ID != uniqueID {
		return errors.New("incorrect password")
	}

	return nil
}

func (c *Connection) sendCommand(typ int32, body []byte) error {
	size := int32(4 + 4 + len(body) + 2)
	uniqueID += 1
	id := uniqueID

	wtr := binaryReadWriter{ByteOrder: binary.LittleEndian}
	wtr.Write(size)
	wtr.Write(id)
	wtr.Write(typ)
	wtr.Write(body)
	wtr.Write([]byte{0x0, 0x0})
	if wtr.err != nil {
		return wtr.err
	}

	c.conn.Write(wtr.buf.Bytes())
	return nil
}

func (c *Connection) readPkg() (pkg, error) {
	const bufSize = 4096
	b := make([]byte, bufSize)

	// Doesn't handle split messages correctly.
	read, err := c.conn.Read(b)
	if err != nil {
		return pkg{}, err
	}

	p := pkg{}
	rdr := binaryReadWriter{ByteOrder: binary.LittleEndian,
		buf: bytes.NewBuffer(b)}
	rdr.Read(&p.Size)
	rdr.Read(&p.ID)
	rdr.Read(&p.Type)
	body := [bufSize - 12]byte{}
	rdr.Read(&body)
	if rdr.err != nil {
		return p, rdr.err
	}
	p.Body = body[:read-12]
	return p, nil
}

type pkg struct {
	Size int32
	ID   int32
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
