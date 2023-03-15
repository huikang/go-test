package heartbeat

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func Test_heartbeat_run(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	go func() {
		fmt.Println("start")
		<-time.After(2 * time.Second)
		cancel()
		fmt.Println("cancel")
	}()

	run(ctx)
	// <-ctx.Done()
	<-ctx.Done()
}
