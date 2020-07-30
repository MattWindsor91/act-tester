// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package copier_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"path"
	"testing"

	"github.com/MattWindsor91/act-tester/internal/observing"
	"github.com/stretchr/testify/mock"

	copy2 "github.com/MattWindsor91/act-tester/internal/copier"

	"github.com/MattWindsor91/act-tester/internal/copier/mocks"

	"github.com/stretchr/testify/assert"
)

type closeBuffer struct {
	bytes.Buffer
	// closed holds whether the buffer was closed.
	closed bool
}

func (c *closeBuffer) Close() error {
	c.closed = true
	return nil
}

// TestSendMapping tests SendMapping on a representative mapping.
func TestSendMapping(t *testing.T) {
	t.Parallel()

	// NB: the 'local' files here actually exist in the filesystem relative to this test.
	mapping := map[string]string{
		path.Join("remote", "bin", "a.out"):         path.Join("testdata", "copy_test", "put1.txt"),
		path.Join("remote", "include", "foo.h"):     path.Join("testdata", "copy_test", "put2.txt"),
		path.Join("remote", "src", "blah", "baz.c"): path.Join("testdata", "copy_test", "put3.txt"),
	}

	var m mocks.Copier

	for _, d := range []string{"bin", "include", path.Join("src", "blah")} {
		m.On("MkdirAll", path.Join("remote", d)).Return(nil).Once()
	}

	buffers := make(map[string]*closeBuffer, len(mapping))
	for r := range mapping {
		buffers[r] = new(closeBuffer)
		m.On("Create", r).Return(buffers[r], nil).Once()
	}

	var o mocks.Observer

	onCopy(&o, observing.BatchStart, func(i int, s string, s2 string) bool {
		return i == len(mapping)
	}).Return().Once()
	onCopy(&o, observing.BatchEnd, func(int, string, string) bool {
		return true
	}).Return().Once()
	for r, l := range mapping {
		r := r
		l := l
		onCopy(&o, observing.BatchStep, func(_ int, dst, src string) bool {
			return r == dst && l == src
		}).Return().Once()
	}

	err := copy2.SendMapping(context.Background(), &m, mapping, &o)
	assert.NoError(t, err)

	if m.AssertExpectations(t) {
		for r, l := range mapping {
			bs, err := ioutil.ReadFile(l)
			assert.NoError(t, err, "reading local test file", l)
			assert.Equal(t, bs, buffers[r].Bytes(), "checking copy occurred from", l, "to", r)
			assert.True(t, buffers[r].closed, "buffer not closed for file", r)
		}
	}
	o.AssertExpectations(t)
}

func onCopy(o *mocks.Observer, k observing.BatchKind, f func(int, string, string) bool) *mock.Call {
	return o.On("OnCopy", mock.MatchedBy(func(m copy2.Message) bool {
		return k == m.Kind && f(m.Num, m.Dst, m.Src)
	}))
}
