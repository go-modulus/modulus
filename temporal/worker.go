package temporal

import (
	"context"
	infraCli "github.com/go-modulus/modulus/cli"
	"github.com/urfave/cli/v2"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
	"sync"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

type localWorker struct {
	worker.Worker
	registeredActivities sync.Map
	registeredWorkflows  sync.Map
}

func (lw *localWorker) RegisterActivity(a interface{}) {
	name := getFunctionName(a)

	if _, loaded := lw.registeredActivities.LoadOrStore(name, struct{}{}); loaded {
		// Activity already registered, skip registration
		return
	}
	lw.Worker.RegisterActivityWithOptions(
		a, activity.RegisterOptions{
			Name: name,
		},
	)
}

func (lw *localWorker) RegisterWorkflow(w interface{}) {
	name := getFunctionName(w)
	if _, loaded := lw.registeredWorkflows.LoadOrStore(name, struct{}{}); loaded {
		// Workflow already registered, skip registration
		return
	}
	lw.Worker.RegisterWorkflowWithOptions(
		w, workflow.RegisterOptions{
			Name: name,
		},
	)
}

type Worker struct {
	runner      *infraCli.Runner
	temporal    client.Client
	registerers []Registerer
}

type WorkersParams struct {
	fx.In

	Runner      *infraCli.Runner
	Temporal    client.Client
	Registerers []Registerer `group:"temporal.registerers"`
}

func NewWorker(params WorkersParams) *Worker {
	return &Worker{
		runner:      params.Runner,
		temporal:    params.Temporal,
		registerers: params.Registerers,
	}
}

func WorkerCommand(w *Worker) *cli.Command {
	return &cli.Command{
		Name: "worker",
		Action: func(ctx *cli.Context) error {
			return w.Invoke(ctx, ctx.String("queue"), ctx.Bool("enable-session-worker"))
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "queue",
				Aliases:  []string{"q"},
				Usage:    "queue name",
				Required: true,
			},
			&cli.BoolFlag{
				Name:    "enable-session-worker",
				Aliases: []string{"s"},
				Usage:   "enable session worker",
				Value:   false,
			},
		},
	}
}

func (w *Worker) Invoke(cliCtx *cli.Context, queue string, enableSessionWorker bool) error {
	return w.runner.Run(
		cliCtx.Context,
		func(ctx context.Context) error {
			tw := worker.New(
				w.temporal, queue, worker.Options{
					EnableSessionWorker: enableSessionWorker,
				},
			)
			lw := &localWorker{
				Worker: tw,
			}

			for _, r := range w.registerers {
				r.Register(lw)
			}

			return lw.Run(w.interruptCh(ctx))
		},
	)
}

func (w *Worker) interruptCh(ctx context.Context) <-chan interface{} {
	interruptCh := make(chan interface{}, 1)
	go func() {
		<-ctx.Done()

		interruptCh <- struct{}{}
	}()

	return interruptCh
}
