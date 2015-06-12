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

	"github.com/corestoreio/csfw/i18n"
	"github.com/corestoreio/csfw/storage/money"
	"github.com/stretchr/testify/assert"
)

func TestJSON(t *testing.T) {

	tests := []struct {
		prec    int
		haveI   int64
		haveF   i18n.CurrencyFormatter
		haveV   bool
		want    string
		wantErr error
	}{
		{100, 123456, i18n.DefaultCurrency, true, `[1234.56, "$ 1.234,56", "$"]`, nil},
		{100, 123456, i18n.DefaultCurrency, false, `null`, nil},
	}

	for _, test := range tests {
		c := money.New(
			money.Precision(test.prec),
			money.Format(test.haveF),
		).Set(test.haveI)
		c.Valid = test.haveV

		have, err := c.MarshalJSON()
		if test.wantErr != nil {
			assert.Error(t, err, "%v", test)
			assert.Nil(t, have)
		} else {
			haveS := string(have)
			assert.NoError(t, err, "%v", test)
			assert.EqualValues(t, test.want, haveS)
			// @todo test unmarshal ...
		}
	}
}
