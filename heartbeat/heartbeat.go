package heartbeat

import (
	"context"
	"fmt"
	"time"
)

func run(ctx context.Context) {
	// incomingHeartbeatCtx will complete if incoming heartbeats time out.
	incomingHeartbeatCtx, incomingHeartbeatCtxCancel :=
		context.WithTimeout(context.Background(), 2*time.Second)

	fmt.Println("init cancelFn", incomingHeartbeatCtxCancel)
	x := 5
	defer func(cancelFn context.CancelFunc) {
		fmt.Println("defer x", x)
		fmt.Println("Defer function", cancelFn)
		cancelFn()
	}(incomingHeartbeatCtxCancel)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("run cancelFn", incomingHeartbeatCtxCancel)
			incomingHeartbeatCtxCancel()
			return
		case <-incomingHeartbeatCtx.Done():
			fmt.Println("heartbeat timeout")
			incomingHeartbeatCtxCancel()
			return
		case <-time.After(1 * time.Second):
			incomingHeartbeatCtxCancel()
			fmt.Println("new incomingHeartbeatCtx")
			x = 2
			incomingHeartbeatCtx, incomingHeartbeatCtxCancel =
				context.WithTimeout(context.Background(), 2*time.Second)
		}
	}
}
