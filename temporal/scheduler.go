package temporal

import (
	"context"
	"errors"
	"github.com/go-modulus/modulus/errors/errlog"
	"github.com/urfave/cli/v2"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/fx"
	"log/slog"
)

type Scheduler struct {
	temporal  client.Client
	schedules []Schedule
	logger    *slog.Logger
}

type SchedulerParams struct {
	fx.In

	Logger    *slog.Logger
	Temporal  client.Client
	Schedules []Schedule `group:"temporal.schedules"`
}

func NewScheduler(params SchedulerParams) *Scheduler {
	return &Scheduler{
		temporal:  params.Temporal,
		schedules: params.Schedules,
		logger:    params.Logger,
	}
}

// SchedulerCommand runs a command to add or updates Temporal schedules.
func SchedulerCommand(scheduler *Scheduler) *cli.Command {
	return &cli.Command{
		Name: "scheduler",
		Usage: "Run it on deploy to update Temporal schedules" +
			"\n" +
			"Example: go run cmd/console/main.go temporal scheduler --queue=default",
		Description: "Adds or updates Temporal schedules. Getting the config from workflow structs that are implementing the Schedule interface.",
		Action: func(ctx *cli.Context) error {
			return scheduler.Invoke(ctx.Context, ctx.String("queue"))
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "queue",
				Aliases:  []string{"q"},
				Usage:    "queue name",
				Required: true,
			},
		},
	}
}

func (w *Scheduler) Invoke(ctx context.Context, queue string) error {
	scheduleClient := w.temporal.ScheduleClient()
	for _, r := range w.schedules {
		opts := r.Schedule(queue)

		handle := scheduleClient.GetHandle(ctx, opts.ID)
		_, err := handle.Describe(ctx)
		if err != nil {
			// if error that means schedule not exists
			_, err := scheduleClient.Create(
				ctx, opts,
			)
			if err != nil {
				if !errors.Is(err, temporal.ErrScheduleAlreadyRunning) {
					return err
				}
			} else {
				w.logger.Info("New task scheduled", slog.String("taskId", opts.ID))
			}
		} else {
			w.logger.Info("Task already scheduled. Try to update", slog.String("taskId", opts.ID))
			err = handle.Update(
				ctx, client.ScheduleUpdateOptions{
					DoUpdate: func(schedule client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
						schedule.Description.Schedule.Spec = &opts.Spec
						schedule.Description.Schedule.Action = opts.Action
						return &client.ScheduleUpdate{
							Schedule: &schedule.Description.Schedule,
						}, nil
					},
				},
			)
			if err != nil {
				w.logger.Error("Cannot update schedule", slog.String("taskId", opts.ID), errlog.Error(err))
				continue
			}
			w.logger.Info("Schedule updated", slog.String("taskId", opts.ID))
		}
	}

	return nil
}
