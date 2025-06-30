package block_call_with_timeout

import (
	"context"
	"errors"
	"time"
)

type result[RT any] struct {
	result *RT
	err    error
}

func BlockCallWithTimeout[RT any](
	ctx context.Context, timeout time.Duration, f func() (*RT, error),
) (*RT, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resultChan := make(chan result[RT], 1)

	go func() {
		res, err := f()

		select {
		case resultChan <- result[RT]{result: res, err: err}:
		case <-timeoutCtx.Done():
		}
	}()

	select {
	case r := <-resultChan:
		return r.result, r.err
	case <-timeoutCtx.Done():
		if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
			return nil, context.DeadlineExceeded
		}
		return nil, timeoutCtx.Err()
	}
}
