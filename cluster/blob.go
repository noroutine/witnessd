package cluster

import "io"

type Blob struct {
    client *Client
    key []byte
    pos int64
}

func (client *Client) Open(key []byte) *Blob {
    // load blob data

    return &Blob{
        client: client,
        key: key,
        pos: 0,
    }
}

func (blob *Blob) Seek(offset int64, whence int) (n int64, err error) {
    return 0, nil
}

func (blob *Blob) Write(p []byte) (n int, err error) {
    return 0, nil
}

func (blob *Blob) WriteByte(c byte) error {
    return nil
}

func (blob *Blob) WriteAt(p []byte, off int64) (n int, err error) {
    return 0, nil
}

func (blob *Blob) WriteTo(w io.Writer) (n int64, err error) {
    return 0, nil
}

func (blob *Blob) Read(p []byte) (n int, err error) {
    return 0, nil
}

func (blob *Blob) ReadByte() (c byte, err error) {
    return 0, nil
}

func (blob *Blob) UnreadByte() error {
    return nil
}

func (blob *Blob) ReadAt(p []byte, off int64) (n int, err error) {
    return 0, nil
}

func (blob *Blob) ReadFrom(r io.Reader) (n int64, err error) {
    return 0, nil
}

func (blob *Blob) Flush() {

}

func (blob *Blob) Size() int64 {
    return 0
}

func (blob *Blob) Close() error {
    return nil
}
