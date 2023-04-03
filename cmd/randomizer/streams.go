package main

import (
	"fmt"
	"io"
)

// A reader that swaps two random bytes in the input
type swapRandomBytesReader struct {
	io.Reader
	Length string
}

func (r *swapRandomBytesReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)

	if err != nil && err != io.EOF {
		return n, err
	}

	fmt.Println("before:", string(p[:n]))
	shuffleBytes(p[:n])
	fmt.Println("after:", string(p[:n]))
	// swapRandomBytes(p[:n])
	return n, err
}

func (r *swapRandomBytesReader) Close() error {
	if closer, ok := r.Reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// LimitReader returns a Reader that reads from r
// but stops with EOF after n bytes.
// The underlying implementation is a *LimitedReader.
func MyLimitReader(r io.Reader, n int64) *MyLimitedReader { return &MyLimitedReader{r, n} }

// A LimitedReader reads from R but limits the amount of
// data returned to just N bytes. Each call to Read
// updates N to reflect the new amount remaining.
// Read returns EOF when N <= 0 or when the underlying R returns EOF.
type MyLimitedReader struct {
	R io.Reader // underlying reader
	N int64     // max bytes remaining
}

func (l *MyLimitedReader) Read(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.R.Read(p)
	l.N -= int64(n)
	return
}

func (r *MyLimitedReader) Close() error {
	if closer, ok := r.R.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
