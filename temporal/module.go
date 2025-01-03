package temporal

import (
	"context"
	"fmt"
	cli2 "github.com/go-modulus/modulus/cli"
	"github.com/go-modulus/modulus/module"
	"github.com/urfave/cli/v2"
	"log/slog"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/fx"
)

type Config struct {
	Address string `env:"TEMPORAL_ADDRESS, default=localhost:7233"`
}

type Registerer interface {
	Register(worker.Registry)
}

func Provide[T Registerer](register interface{}) fx.Option {
	return fx.Provide(
		register,
		fx.Annotate(
			func(a T) T { return a },
			fx.As(new(Registerer)),
			fx.ResultTags(`group:"temporal.registerers"`),
		),
	)
}

// Look client.Client interface
type Starter interface {
	ExecuteWorkflow(
		ctx context.Context,
		options client.StartWorkflowOptions,
		workflow interface{},
		args ...interface{},
	) (client.WorkflowRun, error)

	SignalWithStartWorkflow(
		ctx context.Context,
		workflowID string,
		signalName string,
		signalArg interface{},
		options client.StartWorkflowOptions,
		workflow interface{},
		workflowArgs ...interface{},
	) (client.WorkflowRun, error)

	SignalWorkflow(ctx context.Context, workflowID string, runID string, signalName string, arg interface{}) error
}

func NewStarter(client client.Client) Starter {
	return client
}

type TestringWorkflowRun struct {
	env *testsuite.TestWorkflowEnvironment
}

func (r *TestringWorkflowRun) GetID() string {
	return ""
}

func (r *TestringWorkflowRun) GetRunID() string {
	return ""
}

func (r *TestringWorkflowRun) Get(ctx context.Context, valuePtr interface{}) error {
	return r.env.GetWorkflowResult(valuePtr)
}

func (r *TestringWorkflowRun) GetWithOptions(
	ctx context.Context,
	valuePtr interface{},
	options client.WorkflowRunGetOptions,
) error {
	return r.env.GetWorkflowResult(valuePtr)
}

type TestingStarter struct {
	env *testsuite.TestWorkflowEnvironment
}

func (s TestingStarter) ExecuteWorkflow(
	ctx context.Context,
	options client.StartWorkflowOptions,
	workflow interface{},
	args ...interface{},
) (client.WorkflowRun, error) {
	s.env.SetStartWorkflowOptions(options)
	s.env.ExecuteWorkflow(workflow, args...)

	return &TestringWorkflowRun{env: s.env}, nil
}

func (s TestingStarter) SignalWithStartWorkflow(
	ctx context.Context,
	workflowID string,
	signalName string,
	signalArg interface{},
	options client.StartWorkflowOptions,
	workflow interface{},
	workflowArgs ...interface{},
) (client.WorkflowRun, error) {
	if options.ID == "" {
		options.ID = workflowID
	}
	s.env.SetStartWorkflowOptions(options)
	s.env.RegisterDelayedCallback(
		func() {
			s.env.SignalWorkflow(signalName, signalArg)
		},
		0,
	)
	s.env.ExecuteWorkflow(workflow, workflowArgs...)

	return &TestringWorkflowRun{env: s.env}, nil
}

func (s TestingStarter) SignalWorkflow(
	ctx context.Context,
	workflowID string,
	runID string,
	signalName string,
	arg interface{},
) error {
	return s.env.SignalWorkflowByID(workflowID, signalName, arg)
}

func NewTestingStarter(env *testsuite.TestWorkflowEnvironment) *TestingStarter {
	return &TestingStarter{env: env}
}

func ShouldContunueAsNew(ctx workflow.Context) bool {
	info := workflow.GetInfo(ctx)
	return info.GetCurrentHistoryLength() > 10_000
}

func NewModule() *module.Module {
	config := Config{}
	return module.NewModule("github.com/go-modulus/modulus/temporal").
		AddDependencies(cli2.NewModule()).
		InitConfig(config).
		AddProviders(
			NewStarter,
			NewWorker,

			func(
				config *Config,
				logger *slog.Logger,
			) (client.Client, error) {
				tracingInterceptor, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{})
				if err != nil {
					return nil, fmt.Errorf("unable to create tracing interceptor: %w", err)
				}

				opts := client.Options{
					HostPort:     config.Address,
					Logger:       log.NewStructuredLogger(logger),
					Interceptors: []interceptor.ClientInterceptor{tracingInterceptor},
				}

				return client.NewLazyClient(opts)
			},
		).AddCliCommands(
		func(worker *Worker) *cli.Command {
			return &cli.Command{
				Name: "temporal",
				Subcommands: []*cli.Command{
					worker.Command(),
				},
			}
		},
	)
}
