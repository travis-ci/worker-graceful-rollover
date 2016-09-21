package rollover

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestTimeConsuming(t *testing.T) {
	sem := make(Semaphore, 1)
	conn := ioutil.NopCloser(bytes.NewReader([]byte("")))

	HandleConnection(sem, conn, nil)
}
