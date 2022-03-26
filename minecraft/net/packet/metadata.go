package packet

import (
	"bytes"
	"io"
)

type MetadataFieldType int8

func (t MetadataFieldType) WriteTo(w io.Writer) (n int64, err error) {
	return Byte(t).WriteTo(w)
}

func (t *MetadataFieldType) ReadFrom(r io.Reader) (n int64, err error) {
	var fType Byte
	n, err = fType.ReadFrom(r)
	if err != nil {
		return
	}

	*t = MetadataFieldType(fType)
	return
}

const (
	MetadataTypeByte MetadataFieldType = iota
	MetadataTypeShort
	MetadataTypeInt
	MetadataTypeFloat
	MetadataTypeString
	MetadataTypeSlot
	MetadataTypeIntPos
	MetadataTypeFloatPos
)

type (
	MetadataByte     Byte
	MetadataShort    Short
	MetadataInt      Int
	MetadataFloat    Float
	MetadataString   String
	MetadataSlot     Slot
	MetadataIntPos   struct{ X, Y, Z Int }
	MetadataFloatPos struct{ X, Y, Z Float }
)

func (p MetadataIntPos) WriteTo(w io.Writer) (n int64, err error) {
	var m int64
	m, err = p.X.WriteTo(w)
	n += m
	if err != nil {
		return
	}

	m, err = p.Y.WriteTo(w)
	n += m
	if err != nil {
		return
	}

	m, err = p.Z.WriteTo(w)
	n += m
	if err != nil {
		return
	}

	return
}

func (p *MetadataIntPos) ReadFrom(r io.Reader) (n int64, err error) {
	var m int64
	m, err = p.X.ReadFrom(r)
	n += m
	if err != nil {
		return
	}

	m, err = p.Y.ReadFrom(r)
	n += m
	if err != nil {
		return
	}

	m, err = p.Z.ReadFrom(r)
	n += m
	if err != nil {
		return
	}

	return
}

func (p MetadataFloatPos) WriteTo(w io.Writer) (n int64, err error) {
	var m int64
	m, err = p.X.WriteTo(w)
	n += m
	if err != nil {
		return
	}

	m, err = p.Y.WriteTo(w)
	n += m
	if err != nil {
		return
	}

	m, err = p.Z.WriteTo(w)
	n += m
	if err != nil {
		return
	}

	return
}

func (p *MetadataFloatPos) ReadFrom(r io.Reader) (n int64, err error) {
	var m int64
	m, err = p.X.ReadFrom(r)
	n += m
	if err != nil {
		return
	}

	m, err = p.Y.ReadFrom(r)
	n += m
	if err != nil {
		return
	}

	m, err = p.Z.ReadFrom(r)
	n += m
	if err != nil {
		return
	}

	return
}

type EntityMetadata map[byte]interface{}

func (e EntityMetadata) WriteTo(w io.Writer) (n int64, err error) {
	var m int64

	for index, value := range e {
		var fType MetadataFieldType
		b := &bytes.Buffer{}

		switch value.(type) {
		case MetadataByte:
			fType = MetadataTypeByte
			m, err = (Byte)((value).(MetadataByte)).WriteTo(b)
		case MetadataShort:
			fType = MetadataTypeShort
			m, err = (Short)((value).(MetadataShort)).WriteTo(b)
		case MetadataInt:
			fType = MetadataTypeInt
			m, err = (Int)((value).(MetadataInt)).WriteTo(b)
		case MetadataFloat:
			fType = MetadataTypeFloat
			m, err = (Float)((value).(MetadataFloat)).WriteTo(b)
		case MetadataString:
			fType = MetadataTypeString
			m, err = (String)((value).(MetadataString)).WriteTo(b)
		case MetadataSlot:
			fType = MetadataTypeSlot
			m, err = (Slot)((value).(MetadataSlot)).WriteTo(b)
		case MetadataIntPos:
			fType = MetadataTypeIntPos
			m, err = (value).(MetadataIntPos).WriteTo(b)
		case MetadataFloatPos:
			fType = MetadataTypeFloatPos
			m, err = (value).(MetadataFloatPos).WriteTo(b)
		}

		m, err = UnsignedByte((byte(index)&0x1F | byte(fType)<<5) & 0xFF).WriteTo(w)
		n += m
		if err != nil {
			return
		}

		m, err = b.WriteTo(w)
		n += m
		if err != nil {
			return
		}
	}

	m, err = UnsignedByte(0xff).WriteTo(w)
	n += m
	return
}

func (e *EntityMetadata) ReadFrom(r io.Reader) (n int64, err error) {
	var m int64
	for {
		var index UnsignedByte
		n, err = index.ReadFrom(r)
		if err != nil {
			return
		}

		if index == 127 {
			return
		}

		key, fType := byte(index&0x1F), MetadataFieldType(index>>5)
		var v interface{}

		switch fType {
		case MetadataTypeByte:
			var fv MetadataByte
			m, err = (*Byte)(&fv).ReadFrom(r)
			v = fv
		case MetadataTypeShort:
			var fv MetadataShort
			m, err = (*Short)(&fv).ReadFrom(r)
			v = fv
		case MetadataTypeInt:
			var fv MetadataInt
			m, err = (*Int)(&fv).ReadFrom(r)
			v = fv
		case MetadataTypeFloat:
			var fv MetadataFloat
			m, err = (*Float)(&fv).ReadFrom(r)
			v = fv
		case MetadataTypeString:
			var fv MetadataString
			m, err = (*String)(&fv).ReadFrom(r)
			v = fv
		case MetadataTypeSlot:
			var fv MetadataSlot
			m, err = (*Slot)(&fv).ReadFrom(r)
			v = fv
		case MetadataTypeIntPos:
			var fv MetadataIntPos
			m, err = fv.ReadFrom(r)
			v = fv
		case MetadataTypeFloatPos:
			var fv MetadataFloatPos
			m, err = fv.ReadFrom(r)
			v = fv
		}

		n += m
		if err != nil {
			return
		}

		(*e)[key] = v
	}
}
