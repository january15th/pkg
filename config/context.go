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

package config

import "golang.org/x/net/context"

// ctxKey type is unexported to prevent collisions with context keys defined in
// other packages.
type ctxKey uint

// Key* defines the keys to access a value in a context.Context
const (
	ctxKeyReader ctxKey = iota + 1
	ctxKeyReaderPubSuber
	ctxKeyScopedReader
	ctxKeyWriter
)

// FromContextReader returns a config.Reader from a context. If not found returns
// the config.DefaultManager
func FromContextReader(ctx context.Context) Reader {
	if r, ok := ctx.Value(ctxKeyReader).(Reader); ok {
		return r
	}
	return DefaultManager
}

// NewContextReader adds a config.Reader to a context
func NewContextReader(ctx context.Context, r Reader) context.Context {
	return context.WithValue(ctx, ctxKeyReader, r)
}

// FromContextScopedReader returns a config.ScopedReader from a context.
func FromContextScopedReader(ctx context.Context) (r ScopedReader, ok bool) {
	r, ok = ctx.Value(ctxKeyScopedReader).(ScopedReader)
	return
}

// NewContextScopedReader adds a config.ScopedReader to a context
func NewContextScopedReader(ctx context.Context, r ScopedReader) context.Context {
	return context.WithValue(ctx, ctxKeyScopedReader, r)
}

// FromContextReaderPubSuber returns a config.ReaderPubSuber from a context.
func FromContextReaderPubSuber(ctx context.Context) (r ReaderPubSuber, ok bool) {
	r, ok = ctx.Value(ctxKeyReaderPubSuber).(ReaderPubSuber)
	return
}

// NewContextReaderPubSuber adds a ReaderPubSuber to a context.
func NewContextReaderPubSuber(ctx context.Context, r ReaderPubSuber) context.Context {
	return context.WithValue(ctx, ctxKeyReaderPubSuber, r)
}

// FromContextWriter returns a config.Writer from a context.
func FromContextWriter(ctx context.Context) (w Writer, ok bool) {
	w, ok = ctx.Value(ctxKeyWriter).(Writer)
	return
}

// NewContextWriter adds a writer to a context
func NewContextWriter(ctx context.Context, w Writer) context.Context {
	return context.WithValue(ctx, ctxKeyWriter, w)
}
