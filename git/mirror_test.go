package git

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
)

func TestMirror(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	dir1, dir2, dir3 := filepath.Join(dir, "1"), filepath.Join(dir, "2"), filepath.Join(dir, "3")
	fmt.Println("r1=", dir1)
	fmt.Println("r2=", dir2)
	fmt.Println("r3=", dir3)

	r1 := InitPlain(ctx, dir1, false)

}
