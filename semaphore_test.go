package rollover

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestHandleConnection(t *testing.T) {
	sem := make(Semaphore, 1)
	conn := ioutil.NopCloser(bytes.NewReader([]byte("")))
	notify := new(bytes.Buffer)

	HandleConnection(sem, conn, notify, nil)

	b, err := ioutil.ReadAll(notify)
	if err != nil {
		t.Error(err)
	}

	if string(b) != ".\n" {
		t.Fail()
	}
}
