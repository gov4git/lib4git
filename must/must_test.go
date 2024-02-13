package must

import (
	"context"
	"fmt"
	"testing"
)

func _TestPanic(t *testing.T) {
	Panic(context.Background(), fmt.Errorf("xxx"))
}
