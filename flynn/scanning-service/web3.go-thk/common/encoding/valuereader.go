package encoding

import (
	"fmt"
	"io"
)

type ValueReader interface {
	HasMore() bool
	ReadCount() int
	ReadHeader() (TypeHeader, uint32, error)
	io.ByteReader
	io.Reader
	ReadBytes(length int, bytes []byte) ([]byte, error)
	ReadMultiLength(length int) (uint64, error)
	ReadMultiLengthBytes(length int, bytes []byte) ([]byte, error)
}

type noBufValueReader struct {
	reader    io.Reader
	eof       bool
	readCount int
	header    [1]byte
}

func EndOfFile(err error) bool {
	return err == io.EOF || err == io.ErrUnexpectedEOF
}

func (r *noBufValueReader) filerErr(err error) error {
	if EndOfFile(err) {
		r.eof = true
		return io.EOF
	}
	return err
}

func (r noBufValueReader) ReadCount() int {
	return r.readCount
}

func (r *noBufValueReader) HasMore() bool {
	if r.eof {
		return false
	}
	return true
}

func (r *noBufValueReader) ReadHeader() (TypeHeader, uint32, error) {
	if !r.HasMore() {
		return 0, 0, io.EOF
	}
	b, err := r.ReadByte()
	if err != nil {
		return 0, 0, r.filerErr(err)
	}
	return ParseRTLHeader(b)
}

func (r *noBufValueReader) ReadByte() (byte, error) {
	if !r.HasMore() {
		return 0, io.EOF
	}
	n, err := io.ReadFull(r.reader, r.header[:])
	r.readCount += n
	if err != nil {
		return 0, r.filerErr(err)
	}
	if n <= 0 {
		r.eof = true
		return 0, io.EOF
	}
	return r.header[0], nil
}

func (r *noBufValueReader) Read(p []byte) (int, error) {
	if !r.HasMore() {
		return 0, io.EOF
	}

	n, err := io.ReadFull(r.reader, p)
	r.readCount += n
	return n, r.filerErr(err)
}

func (r *noBufValueReader) ReadBytes(length int, bytes []byte) ([]byte, error) {
	return ReadBytesFromReader(r, length, bytes)
}

func (r *noBufValueReader) ReadMultiLength(length int) (uint64, error) {
	return ReadMultiLengthFromReader(r, length)
}

func (r *noBufValueReader) ReadMultiLengthBytes(length int, bytes []byte) ([]byte, error) {
	return ReadMultiLengthBytesFromReader(r, length, bytes)
}

type bufValueReader struct {
	reader    io.Reader // basic reader
	eof       bool      // if the reader EOF
	lastError error     // error of last reading(if exist, except io.EOF)
	buffer    []byte    // buffered bytes
	available uint32    // length of available bytes in buffer
	offset    uint32    // offset for buffer of reading
	readCount int       // counting the read bytes
}

func (r *bufValueReader) ResetCount() {
	r.readCount = 0
}

func (r bufValueReader) ReadCount() int {
	return r.readCount
}

func (r *bufValueReader) HasMore() bool {
	if r.available > r.offset {
		return true
	}
	return r.next()
}

// next read more bytes to buffer when buffer is empty,
// and return whether has more bytes in buffer
func (r *bufValueReader) next() bool {
	if r.available > r.offset {
		return true
	}

	if r.eof || r.lastError != nil {
		return false
	}

	r.offset = 0
	r.available = 0
	for {
		n, err := r.reader.Read(r.buffer)
		if n > 0 {
			r.available = uint32(n)
		}
		if err != nil {
			if err == io.EOF {
				r.eof = true
			} else {
				r.lastError = err
			}
			break
		}
		if n > 0 {
			break
		}
	}
	return r.available > 0
}

func (r *bufValueReader) forward(count int) {
	r.offset += uint32(count)
	r.readCount += count
}

// GetHeader get 1 byte from buffer, parse the byte to (TypeHeader, length)
// if THSingleByte, length will be the byte value.
// if THZeroValue/THTrue, length will be 0
// if TypeHeader is a single byte header, length will be the length of the content
// if TypeHeader is a multi bytes header, length will be the length of the length of the content
// if anything goes wrong, error will not be nil, and (TypeHeader, length) are all meaningless
func (r *bufValueReader) getHeader() (TypeHeader, uint32, error) {
	if !r.HasMore() {
		if r.lastError != nil {
			return 0, 0, r.lastError
		}
		return 0, 0, io.EOF
	}
	b := r.buffer[r.offset]
	return ParseRTLHeader(b)
}

// ReadHeader GetHeader and move 1byte forward if success
func (r *bufValueReader) ReadHeader() (TypeHeader, uint32, error) {
	th, l, err := r.getHeader()
	if err == nil {
		r.offset++
		r.readCount++
	}
	return th, l, err
}

func (r *bufValueReader) ReadByte() (byte, error) {
	if !r.HasMore() {
		if r.lastError != nil {
			return 0, r.lastError
		}
		return 0, io.EOF
	}
	b := r.buffer[r.offset]
	r.offset++
	r.readCount++
	return b, nil
}

func (r *bufValueReader) Read(p []byte) (int, error) {
	if !r.HasMore() {
		if r.lastError != nil {
			return 0, r.lastError
		}
		return 0, io.EOF
	}

	// copy buffer to p
	n := copy(p, r.buffer[r.offset:r.available])
	r.offset += uint32(n)
	r.readCount += n
	if n >= len(p) {
		return n, nil
	}

	// read until fill full p or reader reach EOF
	ret := n
	for {
		if r.next() == false {
			// no more data
			if r.lastError != nil {
				return ret, r.lastError
			}
			return ret, io.EOF
		}
		n = copy(p[ret:], r.buffer[r.offset:r.available])
		ret += n
		r.offset += uint32(n)
		r.readCount += n
		if ret >= len(p) {
			return ret, nil
		}
	}
}

// ReadBytes read length bytes and return a slice, if parameter bytes length not
// sufficient, will create a new slice
func (r *bufValueReader) ReadBytes(length int, bytes []byte) ([]byte, error) {
	//if length <= 0 {
	//	return bytes, ErrLength
	//}
	//if bytes == nil && length > len(bytes) {
	//	bytes = make([]byte, length)
	//}
	//n, err := r.Read(bytes[0:length])
	//if err != nil {
	//	return bytes, err
	//}
	//if n != length {
	//	return bytes, fmt.Errorf("rtl length error: expect %d but %d readed", length, n)
	//}
	//return bytes, nil
	return ReadBytesFromReader(r, length, bytes)
}

// ReadMultiLength read length of multi bytes header value's length
func (r *bufValueReader) ReadMultiLength(length int) (uint64, error) {
	//if length == 1 {
	//	b, err := r.ReadByte()
	//	if err != nil {
	//		return 0, err
	//	}
	//	return uint64(b), nil
	//} else {
	//	lbuf, err := r.ReadBytes(length, nil)
	//	if err != nil {
	//		return 0, err
	//	}
	//	return Numeric.BytesToUint64(lbuf), nil
	//}
	return ReadMultiLengthFromReader(r, length)
}

func (r *bufValueReader) ReadMultiLengthBytes(length int, bytes []byte) ([]byte, error) {
	//l, err := r.ReadMultiLength(length)
	//if err != nil {
	//	return nil, err
	//}
	//
	//buf, err := r.ReadBytes(int(l), bytes)
	//if err != nil {
	//	return bytes, err
	//}
	//
	//return buf, nil
	return ReadMultiLengthBytesFromReader(r, length, bytes)
}

func ParseRTLHeader(b byte) (TypeHeader, uint32, error) {
	for th, thv := range headerTypeMap {
		if thv.Match(b) {
			switch thv.T {
			case THVTByte:
				return th, uint32(b), nil
			case THVTSingleHeader, THVTMultiHeader:
				l := uint32(b & thv.W)
				if l == 0 {
					l = uint32(thv.W + 1)
				}
				return th, l, nil
			default:
				// should not be here
				panic("unknown type")
			}
		}
	}
	return 0, 0, ErrUnsupported
}

func ReadBytesFromReader(r io.Reader, length int, bytes []byte) ([]byte, error) {
	if length <= 0 {
		return bytes, ErrLength
	}
	if bytes == nil && length > len(bytes) {
		bytes = make([]byte, length)
	}
	n, err := r.Read(bytes[0:length])
	if err != nil {
		return bytes, err
	}
	if n != length {
		return bytes, fmt.Errorf("rtl length error: expect %d but %d readed", length, n)
	}
	return bytes, nil
}

func ReadMultiLengthFromReader(vr ValueReader, length int) (uint64, error) {
	if length == 1 {
		b, err := vr.ReadByte()
		if err != nil {
			return 0, err
		}
		return uint64(b), nil
	} else {
		lbuf, err := vr.ReadBytes(length, nil)
		if err != nil {
			return 0, err
		}
		return Numeric.BytesToUint64(lbuf), nil
	}
}

func ReadMultiLengthBytesFromReader(vr ValueReader, length int, bytes []byte) ([]byte, error) {
	l, err := vr.ReadMultiLength(length)
	if err != nil {
		return nil, err
	}

	buf, err := vr.ReadBytes(int(l), bytes)
	if err != nil {
		return bytes, err
	}

	return buf, nil
}

func NewValueReader(r io.Reader, bufferSize int) ValueReader {
	if bufferSize > 0 {
		return &bufValueReader{
			reader:    r,
			eof:       false,
			buffer:    make([]byte, bufferSize),
			available: 0,
			offset:    0,
		}
	} else {
		return &noBufValueReader{reader: r, eof: false, readCount: 0}
	}
}
