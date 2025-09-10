package parser

import (
	"io"
)

func (t *GetInfoReq) Encoder(w io.Writer) error {
	var err error
	err = hbuf.WriterInt64(w, 1, int64(t.UserId))
	if err != nil {
		return err
	}
	err = hbuf.WriterBytes(w, 0, []byte(t.Name))
	if err != nil {
		return err
	}
	err = hbuf.WriterInt64(w, 2, int64(t.Age))
	if err != nil {
		return err
	}
	return nil
}

func (t *GetInfoReq) Decoder(r io.Reader) error {
	return nil
}

func (t *InfoReq) Encoder(w io.Writer) error {
	var err error
	err = hbuf.WriterInt64(w, 1, int64(t.UserId))
	if err != nil {
		return err
	}
	return nil
}

func (t *InfoReq) Decoder(r io.Reader) error {
	return nil
}

func (t *InfoSet) Encoder(w io.Writer) error {
	var err error
	err = hbuf.WriterInt64(w, 1, int64(t.UserId))
	if err != nil {
		return err
	}
	err = hbuf.WriterBytes(w, 0, []byte(t.Name))
	if err != nil {
		return err
	}
	err = hbuf.WriterInt64(w, 2, int64(t.Age))
	if err != nil {
		return err
	}
	return nil
}

func (t *InfoSet) Decoder(r io.Reader) error {
	return nil
}

func (t *GetInfoResp) Encoder(w io.Writer) error {
	var err error
	err = hbuf.WriterInt64(w, 0, int64(t.V1))
	if err != nil {
		return err
	}
	err = hbuf.WriterInt64(w, 50, int64(t.B1))
	if err != nil {
		return err
	}
	err = hbuf.WriterInt64(w, 1, int64(t.V2))
	if err != nil {
		return err
	}
	err = hbuf.WriterInt64(w, 51, int64(t.B2))
	if err != nil {
		return err
	}
	err = hbuf.WriterInt64(w, 2, int64(t.V3))
	if err != nil {
		return err
	}
	err = hbuf.WriterInt64(w, 52, int64(t.B3))
	if err != nil {
		return err
	}
	err = hbuf.WriterInt64(w, 3, int64(t.V4))
	if err != nil {
		return err
	}
	err = hbuf.WriterInt64(w, 53, int64(t.B4))
	if err != nil {
		return err
	}
	err = hbuf.WriterUint64(w, 4, uint64(t.V5))
	if err != nil {
		return err
	}
	err = hbuf.WriterUint64(w, 54, uint64(t.B5))
	if err != nil {
		return err
	}
	err = hbuf.WriterUint64(w, 5, uint64(t.V6))
	if err != nil {
		return err
	}
	err = hbuf.WriterUint64(w, 55, uint64(t.B6))
	if err != nil {
		return err
	}
	err = hbuf.WriterUint64(w, 6, uint64(t.V7))
	if err != nil {
		return err
	}
	err = hbuf.WriterUint64(w, 56, uint64(t.B7))
	if err != nil {
		return err
	}
	err = hbuf.WriterUint64(w, 7, uint64(t.V8))
	if err != nil {
		return err
	}
	err = hbuf.WriterUint64(w, 57, uint64(t.B8))
	if err != nil {
		return err
	}
	err = hbuf.WriterBool(w, 8, t.V9)
	if err != nil {
		return err
	}
	err = hbuf.WriterBool(w, 58, t.B9)
	if err != nil {
		return err
	}
	err = hbuf.WriterFloat(w, 9, t.V10)
	if err != nil {
		return err
	}
	err = hbuf.WriterFloat(w, 59, t.B10)
	if err != nil {
		return err
	}
	err = hbuf.WriterDouble(w, 10, t.V11)
	if err != nil {
		return err
	}
	err = hbuf.WriterDouble(w, 60, t.B11)
	if err != nil {
		return err
	}
	err = hbuf.WriterBytes(w, 11, []byte(t.V12))
	if err != nil {
		return err
	}
	err = hbuf.WriterBytes(w, 61, []byte(t.B12))
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	return nil
}

func (t *GetInfoResp) Decoder(r io.Reader) error {
	return nil
}
