// Copyright 2015-2016, Cyrill @ Schumacher.fm and the CoreStore contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cfgmodel

import (
	"strconv"
	"strings"

	"github.com/corestoreio/csfw/config"
	"github.com/corestoreio/csfw/store/scope"
	"github.com/corestoreio/csfw/util"
	"github.com/corestoreio/csfw/util/bufferpool"
	"github.com/juju/errors"
)

// CSVSeparator separates CSV values. Default value.
const CSVSeparator = ','

// WithCSVSeparator applies a custom CSV separator to the types
// StringCSV or IntCSV
func WithCSVSeparator(sep rune) Option {
	return func(b *optionBox) Option {
		prev := CSVSeparator
		switch {
		case b.StringCSV != nil:
			prev = b.StringCSV.Separator
			b.StringCSV.Separator = sep
		case b.IntCSV != nil:
			prev = b.IntCSV.Separator
			b.IntCSV.Separator = sep
		}
		return WithCSVSeparator(prev)
	}
}

// StringCSV represents a path in config.Getter which will be saved as a
// CSV string and returned as a string slice. Separator is a comma.
type StringCSV struct {
	Str
	// Separator is your custom separator, default is constant CSVSeparator
	Separator rune
}

// NewStringCSV creates a new CSV string type. Acts as a multiselect.
// Default separator: constant CSVSeparator
func NewStringCSV(path string, opts ...Option) StringCSV {
	ret := StringCSV{
		Separator: CSVSeparator,
		Str:       NewStr(path),
	}
	(&ret).Option(opts...)
	return ret
}

// Option sets the options and returns the last set previous option
func (str *StringCSV) Option(opts ...Option) (previous Option) {
	ob := &optionBox{
		baseValue: &str.baseValue,
		StringCSV: str,
	}
	for _, o := range opts {
		previous = o(ob)
	}
	str = ob.StringCSV
	str.baseValue = *ob.baseValue
	return
}

// Get returns a string slice. Splits the stored string by comma.
// Can return nil,nil. Empty values will be discarded. Returns a slice
// containing unique entries. No validation will be made.
func (str StringCSV) Get(sg config.ScopedGetter) ([]string, error) {
	s, err := str.Str.Get(sg)
	if err != nil {
		return nil, err
	}
	if s == "" {
		return nil, nil
	}
	var ret util.StringSlice = strings.Split(s, string(str.Separator))
	return ret.Unique(), nil
}

// Write writes a slice with its scope and ID to the writer.
// Validates the input string slice for correct values if set in source.Slice.
func (str StringCSV) Write(w config.Writer, sl []string, s scope.Scope, scopeID int64) error {
	for _, v := range sl {
		if err := str.ValidateString(v); err != nil {
			return err
		}
	}
	return str.baseValue.Write(w, strings.Join(sl, string(str.Separator)), s, scopeID)
}

// IntCSV represents a path in config.Getter which will be saved as a
// CSV string and returned as an int64 slice. Separator is a comma.
type IntCSV struct {
	Str
	// Lenient ignores errors in parsing integers
	Lenient bool
	// Separator is your custom separator, default is constant CSVSeparator
	Separator rune
}

// NewIntCSV creates a new int CSV type. Acts as a multiselect.
func NewIntCSV(path string, opts ...Option) IntCSV {
	ret := IntCSV{
		Str:       NewStr(path),
		Separator: CSVSeparator,
	}
	(&ret).Option(opts...)
	return ret
}

// Option sets the options and returns the last set previous option
func (ic *IntCSV) Option(opts ...Option) (previous Option) {
	ob := &optionBox{
		baseValue: &ic.baseValue,
		IntCSV:    ic,
	}
	for _, o := range opts {
		previous = o(ob)
	}
	ic = ob.IntCSV
	ic.baseValue = *ob.baseValue
	return
}

// Get returns an int slice. Int string gets splited by comma.
// Can return nil,nil. If multiple values cannot be casted to int then the
// last known error gets returned.
func (ic IntCSV) Get(sg config.ScopedGetter) ([]int, error) {
	s, err := ic.Str.Get(sg)
	if err != nil {
		return nil, err
	}
	if s == "" {
		return nil, nil
	}

	csv := strings.Split(s, string(ic.Separator))

	ret := make([]int, 0, len(csv))

	for _, line := range csv {
		line = strings.TrimSpace(line)
		if line != "" {
			v, err := strconv.Atoi(line)
			if err != nil && false == ic.Lenient {
				return ret, err
			}
			if err == nil {
				ret = append(ret, v)
			}
		}
	}
	return ret, nil
}

// Write writes int values as a CSV string
func (ic IntCSV) Write(w config.Writer, sl []int, s scope.Scope, scopeID int64) error {

	val := bufferpool.Get()
	defer bufferpool.Put(val)

	for i, v := range sl {

		if err := ic.ValidateInt(v); err != nil {
			return err
		}

		if _, err := val.WriteString(strconv.Itoa(v)); err != nil {
			return errors.Mask(err)
		}
		if i < len(sl)-1 {
			if _, err := val.WriteRune(ic.Separator); err != nil {
				return errors.Mask(err)
			}
		}
	}
	return ic.baseValue.Write(w, val.String(), s, scopeID)
}