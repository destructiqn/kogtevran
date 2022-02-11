package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"

	"github.com/google/uuid"
)

func ReadMap(src []byte) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	reader := bytes.NewReader(src)
	for {
		var key String
		_, err := key.ReadFrom(reader)
		if err != nil {
			if err == io.EOF {
				return m, nil
			}

			return nil, fmt.Errorf("reading key: %s", err)
		}

		var vType Byte
		_, err = vType.ReadFrom(reader)
		if err != nil {
			return nil, fmt.Errorf("reading type: %s", err)
		}

		k := string(key)
		switch vType {
		case 1:
			var v Int
			_, err := v.ReadFrom(reader)
			if err != nil {
				return nil, fmt.Errorf("reading int (1): %s", err)
			}

			m[k] = int32(v)
		case 2:
			var v Byte
			_, err := v.ReadFrom(reader)
			if err != nil {
				return nil, fmt.Errorf("reading byte (2): %s", err)
			}

			m[k] = byte(v)
		case 3:
			var v Long
			_, err := v.ReadFrom(reader)
			if err != nil {
				return nil, fmt.Errorf("reading long (3): %s", err)
			}

			m[k] = int64(v)
		case 4:
			var v String
			_, err := v.ReadFrom(reader)
			if err != nil {
				return nil, fmt.Errorf("reading string (4): %s", err)
			}

			m[k] = string(v)
		case 5:
			var v Short
			_, err := v.ReadFrom(reader)
			if err != nil {
				return nil, fmt.Errorf("reading short (5): %s", err)
			}

			m[k] = int16(v)
		case 6:
			var v Float
			_, err := v.ReadFrom(reader)
			if err != nil {
				return nil, fmt.Errorf("reading float (6): %s", err)
			}

			m[k] = float32(v)
		case 7:
			var v Double
			_, err := v.ReadFrom(reader)
			if err != nil {
				return nil, fmt.Errorf("reading double (7): %s", err)
			}

			m[k] = float64(v)
		case 8:
			var v Boolean
			_, err := v.ReadFrom(reader)
			if err != nil {
				return nil, fmt.Errorf("reading boolean (8): %s", err)
			}

			m[k] = bool(v)
		case 9:
			var v ByteArray
			_, err := v.ReadFrom(reader)
			if err != nil {
				return nil, fmt.Errorf("reading byte array (9): %s", err)
			}

			byteMap, err := ReadMap(v)
			if err != nil {
				return nil, fmt.Errorf("reading byte map (9): %s", err)
			}

			m[k] = byteMap
		case 10:
			var v ByteArray
			_, err := v.ReadFrom(reader)
			if err != nil {
				return nil, fmt.Errorf("reading byte array (10): %s", err)
			}

			m[k] = []byte(v)
		case 11:
			var l VarInt
			_, err := l.ReadFrom(reader)
			if err != nil {
				return nil, fmt.Errorf("reading string array length (11): %s", err)
			}

			v := make([]string, l)
			for i := 0; i < int(l); i++ {
				var s String
				_, err := s.ReadFrom(reader)
				if err != nil {
					return nil, fmt.Errorf("reading string array contents (11): %s", err)
				}

				v[i] = string(s)
			}

			m[k] = v
		case 12:
			var l VarInt
			_, err := l.ReadFrom(reader)
			if err != nil {
				return nil, fmt.Errorf("reading byte map array length (12): %s", err)
			}

			v := make([]map[string]interface{}, l)
			for i := 0; i < int(l); i++ {
				var iv ByteArray
				_, err := iv.ReadFrom(reader)
				if err != nil {
					return nil, fmt.Errorf("reading byte map corresponding byte array (12): %s", err)
				}

				byteMap, err := ReadMap(iv)
				if err != nil {
					return nil, fmt.Errorf("reading byte map in byte array (12): %s", err)
				}

				v[i] = byteMap
			}

			m[k] = v
		case 13:
			var v VarInt
			_, err := v.ReadFrom(reader)
			if err != nil {
				return nil, fmt.Errorf("reading var int (13): %s", err)
			}

			m[k] = int32(v)
		case 14:
			var v Long
			_, err := v.ReadFrom(reader)
			if err != nil {
				return nil, fmt.Errorf("reading long (14): %s", err)
			}

			m[k] = int64(v)
		case 15:
			var v UUID
			_, err := v.ReadFrom(reader)
			if err != nil {
				return nil, fmt.Errorf("reading uuid (15): %s", err)
			}

			m[k] = uuid.UUID(v)
		case 16:
			var l VarInt
			_, err := l.ReadFrom(reader)
			if err != nil {
				return nil, fmt.Errorf("reading var int array length (16): %s", err)
			}

			v := make([]int, l)
			for i := 0; i < int(l); i++ {
				var iv VarInt
				_, err := iv.ReadFrom(reader)
				if err != nil {
					return nil, fmt.Errorf("reading var int array contents (16): %s", err)
				}

				v[i] = int(iv)
			}

			m[k] = v
		case 17:
			var v SignedVarInt
			_, err := v.ReadFrom(reader)
			if err != nil {
				return nil, fmt.Errorf("reading signed var int (17): %s", err)
			}

			m[k] = int32(v)
		case 18:
			var l VarInt
			_, err := l.ReadFrom(reader)
			if err != nil {
				return nil, fmt.Errorf("reading signed var int array (18): %s", err)
			}

			v := make([]int, l)
			for i := 0; i < int(l); i++ {
				var iv SignedVarInt
				_, err := iv.ReadFrom(reader)
				if err != nil {
					return nil, fmt.Errorf("reading signed var int array contents (18): %s", err)
				}

				v[i] = int(iv)
			}

			m[k] = v
		case 19:
			var l VarInt
			_, err := l.ReadFrom(reader)
			if err != nil {
				return nil, fmt.Errorf("reading int array length (19): %s", err)
			}

			v := make([]int, l)
			for i := 0; i < int(l); i++ {
				var iv Int
				_, err := iv.ReadFrom(reader)
				if err != nil {
					return nil, fmt.Errorf("reading int array contents (19): %s", err)
				}

				v[i] = int(iv)
			}

			m[k] = v
		case 20:
			var l VarInt
			_, err := l.ReadFrom(reader)
			if err != nil {
				return nil, fmt.Errorf("reading table first level contents length (20): %s", err)
			}

			v := make([][]string, l)
			for i := 0; i < int(l); i++ {
				var il VarInt
				_, err := il.ReadFrom(reader)
				if err != nil {
					return nil, fmt.Errorf("reading table second level contents length (20): %s", err)
				}

				v[i] = make([]string, il)

				for j := 0; j < len(v[i]); j++ {
					var iv String
					_, err := iv.ReadFrom(reader)
					if err != nil {
						return nil, fmt.Errorf("reading table contents (20): %s", err)
					}

					v[i][j] = string(iv)
				}
			}

			m[k] = v
		case 21:
			var l VarInt
			_, err := l.ReadFrom(reader)
			if err != nil {
				return nil, fmt.Errorf("reading long array length (21): %s", err)
			}

			v := make([]int64, l)
			for i := 0; i < int(l); i++ {
				var iv Long
				_, err := iv.ReadFrom(reader)
				if err != nil {
					return nil, fmt.Errorf("reading long array contents (21): %s", err)
				}

				v[i] = int64(iv)
			}

			m[k] = v
		}
	}
}

func Encode(d interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	t, v := reflect.TypeOf(d), reflect.ValueOf(d)

	switch t.Kind() {
	case reflect.Map:
		for _, mk := range v.MapKeys() {
			mv := v.MapIndex(mk)
			_, err := String(mk.String()).WriteTo(buffer)
			if err != nil {
				return nil, err
			}

			err = writeField(mv.Interface(), buffer)
			if err != nil {
				return nil, err
			}
		}
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			ft, fv := t.Field(i), v.Field(i)
			value, ok := ft.Tag.Lookup("bytemap")
			if !ok {
				continue
			}

			_, err := String(value).WriteTo(buffer)
			if err != nil {
				return nil, err
			}

			err = writeField(fv.Interface(), buffer)
			if err != nil {
				return nil, err
			}
		}
	case reflect.Ptr:
		return Encode(v.Elem().Interface())
	default:
		return nil, fmt.Errorf("unsupported type: %s", v.Type())
	}

	return buffer.Bytes(), nil
}

func writeField(f interface{}, writer io.Writer) error {
	v := reflect.ValueOf(f)
	switch v.Kind() {
	case reflect.Int:
		j := v.Int()
		if j >= 0 && j < 2097152 {
			_, err := Byte(13).WriteTo(writer)
			if err != nil {
				return err
			}

			_, err = VarInt(j).WriteTo(writer)
			if err != nil {
				return err
			}
		} else if j < 0 && j > -1048576 {
			_, err := Byte(17).WriteTo(writer)
			if err != nil {
				return err
			}

			_, err = SignedVarInt(j).WriteTo(writer)
			if err != nil {
				return err
			}
		} else {
			_, err := Byte(1).WriteTo(writer)
			if err != nil {
				return err
			}

			_, err = Int(j).WriteTo(writer)
			if err != nil {
				return err
			}
		}
	case reflect.Float32: // Float
		_, err := Byte(6).WriteTo(writer)
		if err != nil {
			return err
		}

		_, err = Float(v.Float()).WriteTo(writer)
		if err != nil {
			return err
		}
	case reflect.Uint8: // Byte
		_, err := Byte(2).WriteTo(writer)
		if err != nil {
			return err
		}

		_, err = Byte(v.Uint()).WriteTo(writer)
		if err != nil {
			return err
		}
	case reflect.Int16: // Short
		_, err := Byte(5).WriteTo(writer)
		if err != nil {
			return err
		}

		_, err = Short(v.Int()).WriteTo(writer)
		if err != nil {
			return err
		}
	case reflect.Int64: // Long
		i := v.Int()
		if i >= 0 && i < 2097152 {
			_, err := Byte(14).WriteTo(writer)
			if err != nil {
				return err
			}

			_, err = VarInt(i).WriteTo(writer)
			if err != nil {
				return err
			}
		} else {
			_, err := Byte(3).WriteTo(writer)
			if err != nil {
				return err
			}

			_, err = Long(i).WriteTo(writer)
			if err != nil {
				return err
			}
		}
	case reflect.String:
		_, err := Byte(4).WriteTo(writer)
		if err != nil {
			return err
		}

		_, err = String(v.String()).WriteTo(writer)
		if err != nil {
			return err
		}
	case reflect.Float64: // Double
		_, err := Byte(7).WriteTo(writer)
		if err != nil {
			return err
		}

		_, err = Double(v.Float()).WriteTo(writer)
		if err != nil {
			return err
		}
	case reflect.Bool:
		_, err := Byte(8).WriteTo(writer)
		if err != nil {
			return err
		}

		_, err = Boolean(v.Bool()).WriteTo(writer)
		if err != nil {
			return err
		}
	case reflect.Struct, reflect.Map: // ByteMap
		_, err := Byte(9).WriteTo(writer)
		if err != nil {
			return err
		}

		byteMap, err := Encode(f)
		if err != nil {
			return err
		}

		_, err = ByteArray(byteMap).WriteTo(writer)
		if err != nil {
			return err
		}
	case reflect.Array, reflect.Slice: // Arrays of different types
		switch f.(type) {
		case []int:
			a := f.([]int)
			if len(a) == 0 {
				_, err := Byte(16).WriteTo(writer)
				if err != nil {
					return err
				}

				_, err = Byte(0).WriteTo(writer)
				if err != nil {
					return err
				}
			} else {
				flag := true
				flag1 := false

				for k := 0; k < 4 && k < len(a); k++ {
					if a[k] < 0 || a[k] > 2097152 {
						flag = false
					}

					if a[k] < 0 && a[k] > -1048576 {
						flag1 = true
					}
				}

				if flag {
					_, err := Byte(16).WriteTo(writer)
					if err != nil {
						return err
					}

					_, err = VarInt(len(a)).WriteTo(writer)
					if err != nil {
						return err
					}

					for _, m := range a {
						_, err := VarInt(m).WriteTo(writer)
						if err != nil {
							return err
						}
					}
				} else if flag1 {
					_, err := Byte(18).WriteTo(writer)
					if err != nil {
						return err
					}

					_, err = VarInt(len(a)).WriteTo(writer)
					if err != nil {
						return err
					}

					for _, m := range a {
						_, err := SignedVarInt(m).WriteTo(writer)
						if err != nil {
							return err
						}
					}
				} else {
					_, err := Byte(19).WriteTo(writer)
					if err != nil {
						return err
					}

					_, err = VarInt(len(a)).WriteTo(writer)
					if err != nil {
						return err
					}

					for _, m := range a {
						_, err := Int(m).WriteTo(writer)
						if err != nil {
							return err
						}
					}
				}
			}
		case []byte:
			_, err := Byte(10).WriteTo(writer)
			if err != nil {
				return err
			}

			_, err = ByteArray(f.([]byte)).WriteTo(writer)
			if err != nil {
				return err
			}
		case []string:
			a := f.([]string)

			_, err := Byte(11).WriteTo(writer)
			if err != nil {
				return err
			}

			_, err = VarInt(len(a)).WriteTo(writer)
			if err != nil {
				return err
			}

			for _, s := range a {
				_, err := String(s).WriteTo(writer)
				if err != nil {
					return err
				}
			}
		case [][]string:
			a := f.([][]string)

			_, err := Byte(20).WriteTo(writer)
			if err != nil {
				return err
			}

			_, err = VarInt(len(a)).WriteTo(writer)
			if err != nil {
				return err
			}

			for _, i := range a {
				_, err = VarInt(len(i)).WriteTo(writer)
				if err != nil {
					return err
				}

				for _, j := range i {
					_, err = String(j).WriteTo(writer)
					if err != nil {
						return err
					}
				}
			}
		case []int64:
			a := f.([]int64)

			_, err := Byte(21).WriteTo(writer)
			if err != nil {
				return err
			}

			_, err = VarInt(len(a)).WriteTo(writer)
			if err != nil {
				return err
			}

			for _, i := range a {
				_, err := Long(i).WriteTo(writer)
				if err != nil {
					return err
				}
			}
		case []interface{}:
			a := f.([]string)

			_, err := Byte(12).WriteTo(writer)
			if err != nil {
				return err
			}

			_, err = VarInt(len(a)).WriteTo(writer)
			if err != nil {
				return err
			}

			for _, byteMap := range a {
				eByteMap, err := Encode(byteMap)
				if err != nil {
					return err
				}

				_, err = ByteArray(eByteMap).WriteTo(writer)
				if err != nil {
					return err
				}
			}
		}
	default: // All others
		switch f.(type) {
		case uuid.UUID:
			_, err := Byte(15).WriteTo(writer)
			if err != nil {
				return err
			}

			i := f.(uuid.UUID)
			_, err = Long(binary.BigEndian.Uint64(i[:8])).WriteTo(writer)
			if err != nil {
				return err
			}

			_, err = Long(binary.BigEndian.Uint64(i[8:])).WriteTo(writer)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported type: %s", v.Type())
		}
	}

	return nil
}
