package timehelper

import (
	"fmt"
	"time"
)

// RetryTillNill will retry the given function until it returns nil.
func RetryTillNill(fun func() error) {
	for err := fun(); err != nil; {
		fmt.Printf("Failed to execute function. Err: %s\n", err)
		time.Sleep(time.Second)
	}
}
