// Copyright 2015, Cyrill @ Schumacher.fm and the CoreStore contributors
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

package money_test

import (
	"testing"

	"math"

	"github.com/corestoreio/csfw/storage/money"
)

func TestAbs(t *testing.T) {
	tests := []struct {
		have int64
		want int64
	}{
		{13, 13},
		{-13, 13},
		{-45628734653, 45628734653},
	}

	for i, test := range tests {
		c := money.New().Set(test.have)
		have := c.Abs().Raw()
		if have != test.want {
			t.Errorf("\nWant: %d\nHave: %d\nIndex: %d\n", test.want, have, i)
		}
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		have1 int64
		have2 int64
		want  int64
	}{
		{13, 13, 26},
		{-13, -13, -26},
		{-45628734653, -45628734653, -91257469306},
		{math.MaxInt64, 2, 0},
	}

	for i, test := range tests {
		c := money.New().Set(test.have1)
		c = c.Add(money.New().Set(test.have2))
		have := c.Raw()
		if have != test.want {
			t.Errorf("\nWant: %d\nHave: %d\nIndex: %d\n", test.want, have, i)
		}
	}
}

func TestPrecisionAndGet(t *testing.T) {
	tests := []struct {
		prec  int
		have  int64
		wanti int64
		wantf float64
	}{
		{0, 13, 13, 13.0000},
		{10, 13, 1, 1.30000},
		{100, 13, 0, 0.130000},
		{1000, 13, 0, 0.013000},
		{100, -13, 0, -0.130000},
		{0, -45628734653, -45628734653, -45628734653.000000},
		{10, -45628734653, -4562873465, -4562873465.300000},
		{100, -45628734653, -456287346, -456287346.530000},
		{1000, -45628734653, -45628734, -45628734.653000},
		{100, 256, 2, 2.56},
		10: {1234, -45628734653, -4562873, -4562873.46530000}, // fallback to prec 10000
		{100, -45628734655, -456287346, -456287346.550000},
		{100, -45628734611, -456287346, -456287346.110000},
		{100, -45628734699, -456287346, -456287346.990000},
		14: {10000000, 45628734699, 4562, 4562.87346989999969082419},
		15: {10000000, 45628734655, 4562, 4562.87346549999983835733},
	}

	for i, test := range tests {
		c := money.New(money.Precision(test.prec)).Set(test.have)
		haveCI := c.Geti()
		haveCF := c.Getf()

		if haveCI != test.wanti {
			t.Errorf("\nWantI: %d\nHaveI: %d\nIndex: %d\n", test.wanti, haveCI, i)
		}
		if haveCF != test.wantf {
			t.Errorf("\nWantF: %f\nHaveF: %.20f\nIndex: %d\n", test.wantf, haveCF, i)
		}
	}
}

func TestSetf(t *testing.T) {
	tests := []struct {
		prec  int
		want  int64
		havef float64
	}{
		{0, 13, 13.0000},
		{10, 13, 1.30000},
		{100, 13, 0.130000},
		{1000, 13, 0.013000},
		{100, -13, -0.130000},
		{0, -45628734653, -45628734653.000000},
		{10, -45628734653, -4562873465.300000},
		{100, -45628734653, -456287346.530000},
		{1000, -45628734653, -45628734.653000},
		{100, 256, 2.56},
		10: {1234, -45628734653, -4562873.46530000}, // fallback to prec 10000
		{100, -45628734655, -456287346.550000},
		{100, -45628734611, -456287346.110000},
		{100, -45628734699, -456287346.990000},
		14: {10000000, 45628734699, 4562.87346989999969082419},
		15: {10000000, 45628734655, 4562.87346549999983835733},
	}

	for i, test := range tests {
		c := money.New(money.Precision(test.prec)).Setf(test.havef)
		haveR := c.Raw()
		if haveR != test.want {
			t.Errorf("\nWantI: %d\nHaveI: %d\nIndex: %d\n", test.want, haveR, i)
		}
	}
}

func TestSign(t *testing.T) {
	tests := []struct {
		have int64
		want int
	}{{13, 1}, {-13, -1}, {-45628734653, -1}, {45628734699, 1}}
	for i, test := range tests {
		c := money.New().Set(test.have)
		have := c.Sign()
		if have != test.want {
			t.Errorf("\nWant: %d\nHave: %d\nIndex: %d\n", test.want, have, i)
		}
	}
}

func TestSwedish(t *testing.T) {
	tests := []struct {
		prec int
		iv   money.Interval
		have int64
		want string
	}{
		0: {0, money.Interval005, 25689, "25689.00"},
		1: {10, money.Interval005, 25689, "2568.90"},
		2: {100, money.Interval005, 25689, "256.90"},
		3: {1000, money.Interval005, 25689, "25.70"},
		4: {100, money.Interval005, 25642, "256.40"},
		5: {100, money.Interval005, 25644, "256.45"},
	}
	for i, test := range tests {
		c := money.New(money.Precision(test.prec)).Set(test.have)
		have := c.Swedish(money.Swedish(test.iv)).Unformatted()
		if have != test.want {
			t.Errorf("\nWant: %s\nHave: %s\nIndex: %d\n", test.want, have, i)
		}
	}
}
