package cypress

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"io"
)

type typeDecoder func(d *Decoder) (*Message, error)

type Decoder struct {
	r *bufio.Reader

	decoder typeDecoder

	kv *KVParser
	js *json.Decoder
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r: bufio.NewReader(r),
	}
}

func (d *Decoder) probe() error {
	b, err := d.r.ReadByte()
	if err != nil {
		return err
	}

	err = d.r.UnreadByte()
	if err != nil {
		return err
	}

	switch b {
	case '+':
		d.decoder = decodeNative
	case '>':
		d.kv = NewKVParser(d.r)
		d.decoder = decodeKV
	case '{':
		d.js = json.NewDecoder(d.r)
		d.decoder = decodeJSON
	default:
		return ErrUnknownStreamType
	}

	return nil
}

func (d *Decoder) Decode() (*Message, error) {
	if d.decoder == nil {
		err := d.probe()
		if err != nil {
			return nil, err
		}
	}

	return d.decoder(d)
}

func decodeNative(d *Decoder) (*Message, error) {
	b, err := d.r.ReadByte()
	if err != nil {
		return nil, err
	}

	if b != '+' {
		return nil, ErrUnknownStreamType
	}

	dataLen, err := binary.ReadUvarint(d.r)
	if err != nil {
		return nil, err
	}

	buf := pbBufPool.Get().([]byte)

	if uint64(len(buf)) < dataLen {
		buf = make([]byte, dataLen)
	}

	sbuf := buf[:dataLen]

	_, err = io.ReadFull(d.r, sbuf)
	if err != nil {
		pbBufPool.Put(buf)
		return nil, err
	}

	m := &Message{Version: DEFAULT_VERSION}

	err = m.Unmarshal(sbuf)

	pbBufPool.Put(buf)

	if err != nil {
		return nil, err
	}

	return m, nil
}

func decodeKV(d *Decoder) (*Message, error) {
	return d.kv.Generate()
}

func decodeJSON(d *Decoder) (*Message, error) {
	m := &Message{}

	err := d.js.Decode(m)
	if err != nil {
		return nil, err
	}

	m.Version = DEFAULT_VERSION

	return m, nil
}
