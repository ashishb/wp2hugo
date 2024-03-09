package wpparser

import (
	"bytes"
	"io"
)

type InvalidatorCharacterRemover struct {
	reader io.Reader
}

func (i InvalidatorCharacterRemover) Read(p []byte) (int, error) {
	tmp := make([]byte, len(p))
	n, err := i.reader.Read(tmp)
	if err != nil {
		return n, err
	}
	tmp = bytes.ReplaceAll(tmp, []byte("\u000C"), []byte(""))
	copy(p, tmp)
	return len(tmp), nil
}
