package temporal

import (
	"context"
	"crypto/tls"
	"fmt"
	cli2 "github.com/go-modulus/modulus/cli"
	"github.com/go-modulus/modulus/module"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
	Host      string `env:"TEMPORAL_HOST, default=localhost:7233"`
	Namespace string `env:"TEMPORAL_NAMESPACE"`
	ApiKey    string `env:"TEMPORAL_API_KEY"`
}

type RegisterOption func() interface{}

type Registerer interface {
	Register(worker.Registry)
}

func Provide[T Registerer](register interface{}, opts ...RegisterOption) fx.Option {
	provider := []any{
		register,
		fx.Annotate(
			func(a T) T { return a },
			fx.As(new(Registerer)),
			fx.ResultTags(`group:"temporal.registerers"`),
		),
	}
	for _, opt := range opts {
		provider = append(provider, opt())
	}
	return fx.Provide(
		provider...,
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

type TestingWorkflowRun struct {
	env *testsuite.TestWorkflowEnvironment
}

func (r *TestingWorkflowRun) GetID() string {
	return ""
}

func (r *TestingWorkflowRun) GetRunID() string {
	return ""
}

func (r *TestingWorkflowRun) Get(ctx context.Context, valuePtr interface{}) error {
	return r.env.GetWorkflowResult(valuePtr)
}

func (r *TestingWorkflowRun) GetWithOptions(
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

	return &TestingWorkflowRun{env: s.env}, nil
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

	return &TestingWorkflowRun{env: s.env}, nil
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

func ShouldContinueAsNew(ctx workflow.Context) bool {
	info := workflow.GetInfo(ctx)
	return info.GetCurrentHistoryLength() > 10_000
}

type Schedule interface {
	Schedule(queue string) client.ScheduleOptions
}

func ScheduleAnnotation[T Schedule]() interface{} {
	return fx.Annotate(
		func(a T) T { return a },
		fx.As(new(Schedule)),
		fx.ResultTags(`group:"temporal.schedules"`),
	)
}

func NewModule() *module.Module {
	config := Config{}
	return module.NewModule("temporal").
		AddDependencies(cli2.NewModule()).
		InitConfig(config).
		AddProviders(
			NewStarter,
			NewWorker,
			NewScheduler,

			func(
				config Config,
				logger *slog.Logger,
			) (client.Client, error) {
				tracingInterceptor, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{})
				if err != nil {
					return nil, fmt.Errorf("unable to create tracing interceptor: %w", err)
				}

				opts := client.Options{
					HostPort:     config.Host,
					Logger:       log.NewStructuredLogger(logger),
					Interceptors: []interceptor.ClientInterceptor{tracingInterceptor},
				}
				if config.Namespace != "" {
					opts.Namespace = config.Namespace
					opts.Credentials = client.NewAPIKeyStaticCredentials(config.ApiKey)

					opts.ConnectionOptions = client.ConnectionOptions{
						TLS: &tls.Config{},
						DialOptions: []grpc.DialOption{
							grpc.WithUnaryInterceptor(
								func(
									ctx context.Context,
									method string,
									req any,
									reply any,
									cc *grpc.ClientConn,
									invoker grpc.UnaryInvoker,
									opts ...grpc.CallOption,
								) error {
									if req == nil {
										return fmt.Errorf("request is nil")
									}

									return invoker(
										metadata.AppendToOutgoingContext(ctx, "temporal-namespace", config.Namespace),
										method,
										req,
										reply,
										cc,
										opts...,
									)
								},
							),
						},
					}
				}

				return client.NewLazyClient(opts)
			},
		).AddCliCommands(
		func(
			worker *Worker,
			scheduler *Scheduler,
		) *cli.Command {
			return &cli.Command{
				Name: "temporal",
				Subcommands: []*cli.Command{
					WorkerCommand(worker),
					SchedulerCommand(scheduler),
				},
			}
		},
	)
}

func NewManifestModule() module.ManifestModule {
	temporalModule := module.NewManifestModule(
		NewModule(),
		"github.com/go-modulus/modulus/temporal",
		"Temporal module for Modulus framework.",
		"1.0.0",
	)
	temporalModule.Install.AppendFiles(
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/temporal/install/temporal.mk",
			DestFile:  "mk/temporal.mk",
		},
	)
	return temporalModule
}
