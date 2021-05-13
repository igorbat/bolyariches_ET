// This file is part of Go WML.
//
// Go WML is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Go WML is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Go WML.  If not, see <https://www.gnu.org/licenses/>.

package wml

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type Data map[string]interface{}
type RawData string
type Multiple []interface{}
type Tr string
type Raw string
type Domain struct {
	V interface{}
	D string
}

// Merges multiple Datas, priority to the last parameter
func MergeData(first Data, others ...Data) Data {
	data := make(Data)
	for k, v := range first {
		data[k] = v
	}
	for _, d := range others {
		for k, v := range d {
			data[k] = v
		}
	}
	return data
}

func (d Data) Bytes() []byte {
	return []byte(d.String())
}

func (d Data) String() string {
	return d.Indent(0)
}

func (d *Data) Contains(key string) bool {
	_, ok := (*d)[key]
	return ok
}

func (d *Data) Single() (string, interface{}, error) {
	if len(*d) == 1 {
		for k, v := range *d {
			return k, v, nil
		}
	}
	return "", nil, errors.New("The data isn't single-valued.")
}

func (d *Data) ToTag() (Tag, error) {
	key, value, err := d.Single()
	if err == nil {
		if data, ok := value.(Data); ok {
			return Tag{key, data}, nil
		}
	}
	return Tag{}, errors.New("The data isn't a tag.")
}

func (d *Data) ReadData(path string) (Data, error) {
	value, err := d.Read(path)
	if err != nil {
		return nil, err
	}
	if data, ok := value.(Data); ok {
		return data, nil
	} else {
		return nil, errors.New("Incorrect type of the Data attribute.")
	}
}

func (d *Data) ReadString(path string) (string, error) {
	value, err := d.Read(path)
	if err != nil {
		return "", err
	}
	if data, ok := value.(string); ok {
		return data, nil
	} else {
		return "", errors.New("Incorrect type of the Data attribute.")
	}
}

// Read nested data using dot notation.
func (d *Data) Read(path string) (interface{}, error) {
	pathSlice := strings.Split(path, ".")

	var data Data = *d
	for _, v := range pathSlice[:len(pathSlice)-1] {
		if data.Contains(v) {
			switch data[v].(type) {
			case Data:
				data = data[v].(Data)
			default:
				return nil, errors.New("The path doesn't exist")
			}
		}
	}

	key := pathSlice[len(pathSlice)-1]
	if data.Contains(key) {
		return data[key], nil
	} else {
		return nil, errors.New("The path doesn't exist")
	}
}

func (d *Data) Indent(nesting uint) string {
	tabulation := strings.Repeat("\t", int(nesting))
	var keys []string
	for k := range *d {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	attributes := ""
	subTags := ""
	for _, key := range keys {
		var prepend string
		var value interface{}
		// Check whether the data is Domain or not. If it's Domain, add a textdomain line.
		switch (*d)[key].(type) {
		case Domain:
			prepend = tabulation + "#textdomain " + (*d)[key].(Domain).D + "\n"
			value = (*d)[key].(Domain).V
		default:
			prepend = ""
			value = (*d)[key]
		}
		switch value.(type) {
		case bool:
			var v string
			if value.(bool) {
				v = "yes"
			} else {
				v = "no"
			}
			attributes += prepend
			attributes += tabulation + key + "=" + v + "\n"
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			float32, float64:
			attributes += prepend
			attributes += tabulation + key + "=" + fmt.Sprintf("%v", value) + "\n"
		case string:
			attributes += prepend
			attributes += tabulation + key + "=\"" + value.(string) + "\"\n"
		case Tr:
			attributes += prepend
			attributes += tabulation + key + "=_\"" + string(value.(Tr)) + "\"\n"
		case Raw:
			attributes += prepend
			attributes += tabulation + key + "=" + string(value.(Raw)) + "\n"
		case Data:
			v := value.(Data)
			subTags += prepend
			subTags += (&Tag{key, v}).Indent(nesting)
		case RawData:
			subTags += prepend
			subTags += tabulation + "[" + key + "]\n" +
				string(value.(RawData)) +
				tabulation + "[/" + key + "]\n"
		case Multiple:
			for _, v := range value.(Multiple) {
				switch v.(type) {
				case Domain:
					switch v.(Domain).V.(type) {
					case Data, RawData:
						subTags += prepend
						subTags += (&Data{key: v}).Indent(nesting)
					}
				case Data, RawData:
					subTags += prepend
					subTags += (&Data{key: v}).Indent(nesting)
				}
			}
			/*case []Data:
				for _, v := range value.([]Data) {
					subTags += prepend
					subTags += (&Tag{key, v}).Indent(nesting)
				}
			case []RawData:
				for _, v := range value.([]RawData) {
					subTags += prepend
					subTags += tabulation + "[" + key + "]\n" +
						IndentString(string(v), nesting+1) +
						tabulation + "[/" + key + "]\n"
				}
			case []Domain:
				for _, v := range value.([]Domain) {
					switch v.V.(type) {
					case Data, RawData:
						subTags += prepend
						subTags += (&Data{key: v}).Indent(nesting)
					}
				}*/
		}
	}
	return attributes + subTags
}
