// Licensed to ClickHouse, Inc. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. ClickHouse, Inc. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package column

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2/lib/binary"
)

type Type string

func (t Type) params() string {
	switch start, end := strings.Index(string(t), "("), strings.LastIndex(string(t), ")"); {
	case len(t) == 0, start <= 0, end <= 0, end < start:
		return ""
	default:
		return string(t[start+1 : end])
	}
}

type Error struct {
	ColumnType string
	Err        error
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.ColumnType, e.Err)
}

type ColumnConverterError struct {
	Op       string
	Hint     string
	From, To string
}

func (e *ColumnConverterError) Error() string {
	var hint string
	if len(e.Hint) != 0 {
		hint += ". " + e.Hint
	}
	return fmt.Sprintf("clickhouse [%s]: converting %s to %s is unsupported%s", e.Op, e.From, e.To, hint)
}

type UnsupportedColumnTypeError struct {
	t Type
}

func (e *UnsupportedColumnTypeError) Error() string {
	return fmt.Sprintf("clickhouse: unsupported column type %q", e.t)
}

type Interface interface {
	Name() string
	Type() Type
	Rows() int
	Row(i int, ptr bool) interface{}
	ScanRow(dest interface{}, row int) error
	Append(v interface{}) (nulls []uint8, err error)
	AppendRow(v interface{}) error
	AppendStringRow(v string) error
	Decode(decoder *binary.Decoder, rows int) error
	Encode(*binary.Encoder) error
	ScanType() reflect.Type
	Reset()
}

type CustomSerialization interface {
	ReadStatePrefix(*binary.Decoder) error
	WriteStatePrefix(*binary.Encoder) error
}

func (col *String) AppendStringRow(v string) error {
	col.data = append(col.data, v)
	return nil
}

func (col *DateTime64) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *UUID) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *Ring) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *Map) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *DateTime) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *Decimal) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *Point) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *Array) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *LowCardinality) AppendStringRow(v string) error {
	return col.AppendRow(v)
}
func (col *Tuple) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *IPv4) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *IPv6) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *BigInt) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *Polygon) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *FixedString) AppendStringRow(v string) (err error) {
	return fmt.Errorf("unsupport")
}
func (col *Bool) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *Nullable) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *Date) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (jCol *JSONObject) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *SimpleAggregateFunction) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *Date32) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *MultiPolygon) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *Float32) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *Float64) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *Int8) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *Int16) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *Int32) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *Int64) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *UInt8) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *UInt16) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *UInt32) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *UInt64) AppendStringRow(v string) error {
	return fmt.Errorf("unsupport")
}
func (col *Enum16) AppendStringRow(elem string) error {
	return fmt.Errorf("unsupport")
}
func (col *Enum8) AppendStringRow(elem string) error {
       return col.AppendRow(elem)
}
func (Nothing) AppendStringRow(string) error {
	return fmt.Errorf("unsupport")
}
func (Interval) AppendStringRow(string) error {
	return fmt.Errorf("unsupport")
}

func (col *String) Reset() {
	col.data = col.data[:0]
}

func (col *DateTime64) Reset() {
	col.values.data = col.values.data[:0]
}
func (col *UUID) Reset() {
	col.data = col.data[:0]
}
func (col *Ring) Reset() {
	col.set.Reset()
}
func (col *Map) Reset() {
	col.keys.Reset()
	col.values.Reset()
}
func (col *DateTime) Reset() {
	col.values.data = col.values.data[:0]
}
func (col *Decimal) Reset() {
	col.values = col.values[:0]
}
func (col *Point) Reset() {
	col.lon.data = col.lon.data[:0]
	col.lat.data = col.lat.data[:0]
}
func (col *Array) Reset() {
	col.values.Reset()
	for i := range col.offsets {
		col.offsets[i].values.Reset()
	}
}
func (col *LowCardinality) Reset() {
	col.index.Reset()
	col.rows = 0
	col.keys8.data = col.keys8.data[:0]
	col.keys16.data = col.keys16.data[:0]
	col.keys32.data = col.keys32.data[:0]
	col.keys64.data = col.keys64.data[:0]
	col.append.keys = col.append.keys[:0]
	col.append.index = make(map[interface{}]int)
}
func (col *Tuple) Reset() {
	col.columns = col.columns[:0]
}
func (col *IPv4) Reset() {
	col.data = col.data[:0]
}
func (col *IPv6) Reset() {
	col.data = col.data[:0]
}
func (col *BigInt) Reset() {
	col.data = col.data[:0]
}
func (col *Polygon) Reset() {
	col.set.Reset()
}
func (col *FixedString) Reset() {
	col.data = col.data[:0]
}
func (col *Bool) Reset() {
	col.values.data = col.values.data[:0]
}
func (col *Nullable) Reset() {
	col.base.Reset()
	col.nulls.data = col.nulls.data[:0]
}
func (col *Date) Reset() {
	col.values.data = col.values.data[:0]
}
func (jCol *JSONObject) Reset() {
	jCol.columns = jCol.columns[:0]
}
func (col *SimpleAggregateFunction) Reset() {
	col.base.Reset()
}
func (col *Date32) Reset() {
	col.values.data = col.values.data[:0]
}
func (col *MultiPolygon) Reset() {
	col.set.Reset()
}
func (col *Float32) Reset() {
	col.data = col.data[:0]
}
func (col *Float64) Reset() {
	col.data = col.data[:0]
}
func (col *Int8) Reset() {
	col.data = col.data[:0]
}
func (col *Int16) Reset() {
	col.data = col.data[:0]
}
func (col *Int32) Reset() {
	col.data = col.data[:0]
}
func (col *Int64) Reset() {
	col.data = col.data[:0]
}
func (col *UInt8) Reset() {
	col.data = col.data[:0]
}
func (col *UInt16) Reset() {
	col.data = col.data[:0]
}
func (col *UInt32) Reset() {
	col.data = col.data[:0]
}
func (col *UInt64) Reset() {
	col.data = col.data[:0]
}
func (col *Enum16) Reset() {
	col.values.data = col.values.data[:0]
	// FIXME: Since the line numbers of the patch files are aligned, we will not delete them yet, just comment them out. col.iv = make(map[string]uint16)
	// FIXME: Since the line numbers of the patch files are aligned, we will not delete them yet, just comment them out. col.vi = make(map[uint16]string)
}
func (col *Enum8) Reset() {
	col.values.data = col.values.data[:0]
	// FIXME: Since the line numbers of the patch files are aligned, we will not delete them yet, just comment them out. col.iv = make(map[string]uint8)
	// FIXME: Since the line numbers of the patch files are aligned, we will not delete them yet, just comment them out. col.vi = make(map[uint8]string)
}
func (Nothing) Reset() {
}
func (col *Interval) Reset() {
	col.values.data = col.values.data[:0]
}
