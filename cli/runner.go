package cli

import (
	"context"
	"github.com/go-modulus/modulus/errlog"
	"github.com/go-modulus/modulus/logger"
	"log/slog"
	"sync"

	"go.uber.org/fx"
)

type Runner struct {
	shutdowner fx.Shutdowner
	done       chan struct{}
	stopOnce   sync.Once
	waitOnce   sync.Once
	wg         sync.WaitGroup
	logger     *slog.Logger
}

func NewRunner(shutdowner fx.Shutdowner, logger *slog.Logger) *Runner {
	return &Runner{
		shutdowner: shutdowner,
		done:       make(chan struct{}),
		logger:     logger,
	}
}

func (p *Runner) start(fn func() error) error {
	go func() {
		defer logger.Recover(p.logger)
		defer func() {
			// Shutdown app when all goroutines are done.
			p.waitOnce.Do(
				func() {
					p.wg.Wait()
					err := p.shutdowner.Shutdown()
					if err != nil {
						p.logger.Error(
							"error occurred while shutting down app",
							errlog.Error(err),
						)
					}
				},
			)
		}()

		err := fn()
		if err != nil {
			p.logger.Error(
				"error occurred while starting app",
				errlog.Error(err),
			)
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
