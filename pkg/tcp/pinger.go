package tcp

import (
	"context"
	"fmt"
	"io"
	"time"
)

const defaultPingInterval = 5 * time.Second

func Pinger(ctx context.Context, w io.Writer, reset <-chan time.Duration) {
	var interval time.Duration

	select {
	case <-ctx.Done():
		return
	case interval = <-reset: // pulled initial interval off reset channel (1)
	default:
	}

	if interval <= 0 {
		interval = defaultPingInterval
	}

	timer := time.NewTimer(interval) // (2)
	defer func() {
		if !timer.Stop() {
			<-timer.C
		}
	}()

	for {
		fmt.Println("Running with interval ", interval)
		select {
		case <-ctx.Done():
			return // (3)
		case newInterval := <-reset:
			if !timer.Stop() {
				<-timer.C
			}
			if newInterval > 0 {
				interval = newInterval
			}
		case <-timer.C: // 5
			if _, err := w.Write([]byte("ping")); err != nil {
				fmt.Println("TIMED OUT ", err)
				// act on consec timetouts
				return
			}

		}
		_ = timer.Reset(interval) // 6
	}

}
