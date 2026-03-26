package cli

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/fx"
)

// Runner runs a function inside a goroutine running FX shutdowner for graceful shutdown.
type Runner struct {
	shutdowner   fx.Shutdowner
	done         chan struct{}
	stopOnce     sync.Once
	waitOnce     sync.Once
	wg           sync.WaitGroup
	errorHandler ErrorHandler
}

func NewRunner(
	shutdowner fx.Shutdowner,
	errorHandler ErrorHandler,
) *Runner {
	return &Runner{
		shutdowner:   shutdowner,
		done:         make(chan struct{}),
		errorHandler: errorHandler,
	}
}

func (p *Runner) start(fn func() error) error {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				formattedErr := fmt.Errorf("%v", err)
				p.errorHandler.HandleError(formattedErr)
			}
		}()
		defer func() {
			// Shutdown app when all goroutines are done.
			p.waitOnce.Do(
				func() {
					p.wg.Wait()
					err := p.shutdowner.Shutdown()
					if err != nil {
						p.errorHandler.HandleError(err)
					}
				},
			)
		}()

		err := fn()
		if err != nil {
			p.errorHandler.HandleError(err)
		}
	}()

	return nil
}

func (p *Runner) stop() error {
	p.stopOnce.Do(
		func() {
			close(p.done)
			p.wg.Wait()
		},
	)

	return nil
}

func (p *Runner) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		defer cancel()
		select {
		case <-ctx.Done():
		case <-p.done:
		}
	}()

	p.wg.Add(1)
	defer p.wg.Done()

	return fn(ctx)
}
