package defs

func Fprint(w io.Writer, a ...interface{}) (n int, err error)

// Writer is the interface that wraps the basic Write method.
//
// Write writes len(p) bytes from p to the underlying data stream. It
// returns the number of bytes written from p (0 <= n <= len(p)) and any
// error encountered that caused the write to stop early. Write must return
// a non-nil error if it returns n < len(p).
type Writer interface {
	Write(p []byte) (n int, err error)
}

type Conn interface {
    Read(b []byte) (n int, err error)
    Write(b []byte) (n int, err error)
    Close() error
    // ... some additional methods omitted ...
}
