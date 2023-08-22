package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gofrs/flock"
)

func TestFileLock(t *testing.T) {
	lockFile := filepath.Join(os.TempDir(), "filelock")

	flk1 := flock.New(lockFile)
	locked1, err1 := flk1.TryLock()
	if err1 != nil {
		t.Errorf("lock 1 (%v)", err1)
	}
	if !locked1 {
		t.Errorf("expecting lock to succeed")
	}

	flk2 := flock.New(lockFile)
	locked2, err2 := flk2.TryLock()
	if err2 != nil {
		t.Errorf("lock 2 (%v)", err2)
	}
	if locked2 {
		t.Errorf("expecting lock to fail")
	}
}
