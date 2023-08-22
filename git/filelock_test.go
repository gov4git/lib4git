package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gofrs/flock"
)

func TestFileLock(t *testing.T) {
	lockFile := filepath.Join(os.TempDir(), "filelock")

	flk1 := flock.New(lockFile)
	locked1, err1 := flk1.TryLock()
	if err1 != nil {
		t.Errorf("lock#1 (%v)", err1)
	}
	if !locked1 {
		t.Errorf("expecting lock#1 to succeed")
	}

	flk2 := flock.New(lockFile)
	locked2, err2 := flk2.TryLock()
	if err2 != nil {
		t.Errorf("lock#2 (%v)", err2)
	}
	if locked2 {
		t.Errorf("expecting lock#2 to fail")
	}

	// unlock lock#1 in one second
	go func() {
		time.Sleep(time.Second)
		if err := flk1.Unlock(); err != nil {
			t.Errorf("unlock 1 (%v)", err)
		}
	}()

	flk3 := flock.New(lockFile)
	locked3, err3 := flk3.TryLockContext(context.Background(), time.Second*2)
	if err3 != nil {
		t.Errorf("lock#3 (%v)", err3)
	}
	if !locked3 {
		t.Errorf("expecting lock#3 to succeed")
	}

}
