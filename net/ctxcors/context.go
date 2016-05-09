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

package ctxcors

import (
	"context"

	"github.com/corestoreio/csfw/util/errors"
)

type keyCtxToken struct{}

type wrapperCtx struct {
	err error
}

// FromContext returns an error not caught by the error handler
func FromContext(ctx context.Context) error {
	wrp, ok := ctx.Value(keyCtxToken{}).(wrapperCtx)
	if !ok {
		return nil
	}
	return errors.Wrap(wrp.err, "[ctxcors] FromContext")
}

// withContextError creates a new context with an error attached.
func withContextError(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, keyCtxToken{}, wrapperCtx{err: err})
}
