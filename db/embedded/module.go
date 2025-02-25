package embedded

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/go-modulus/modulus/db/migrator"
	"github.com/go-modulus/modulus/db/pgx"
	"github.com/go-modulus/modulus/module"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"time"
)

type ModuleConfig struct {
	RuntimePath string                `env:"EMBEDDED_POSTGRES_RUNTIME_PATH, default=/tmp/embeddedpostgres"`
	DataPath    string                `env:"EMBEDDED_POSTGRES_DATA_PATH, default=/tmp/embeddedpostgres/data" comment:"The path that will be used for the Postgres data directory. If you want to persist data between restarts, set this variable to a path that is not inside the EMBEDDED_POSTGRES_RUNTIME_PATH."`
	RunPg       bool                  `env:"EMBEDDED_POSTGRES_RUN, default=true" comment:"Set this variable to false if you want to disable the embedded postgres database."`
	PgxConfig   *pgx.ConnectionConfig `env:",prefix=PG_"` // This is the connection config for the pgx module
}

func registerHooks(
	lifecycle fx.Lifecycle,
	config ModuleConfig,
	postgres *embeddedpostgres.EmbeddedPostgres,
	migrateCommand *migrator.Migrate,
) {
	if !config.RunPg {
		return
	}
	lifecycle.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				fmt.Println("Starting embedded postgres...")
				err := postgres.Start()
				if err != nil {
					return err
				}
				defer func() {
					if err != nil {
						_ = postgres.Stop()
					}
				}()
				pwd := "postgres"
				if config.PgxConfig.User == "postgres" {
					pwd = config.PgxConfig.Password
				}
				cfg := pgx.ConnectionConfig{
					Host:     config.PgxConfig.Host,
					Port:     config.PgxConfig.Port,
					User:     "postgres",
					Password: pwd,
					Database: "postgres",
					SslMode:  "disable",
				}

				pg, err := pgxpool.New(ctx, cfg.Dsn())
				if err != nil {
					return err
				}
				defer pg.Close()

				err = createDatabase(ctx, config, pg)
				if err != nil {
					return err
				}

				err = createUser(ctx, config, pg)
				if err != nil {
					return err
				}

				err = migrateCommand.Invoke(ctx)
				if err != nil {
					return err
				}

				return nil
			},
			OnStop: func(ctx context.Context) error {
				fmt.Println("Finishing embedded postgres...")
				if err := postgres.Stop(); err != nil {
					return err
				}
				return nil
			},
		},
	)
}

func createDatabase(ctx context.Context, config ModuleConfig, pg *pgxpool.Pool) error {
	if config.PgxConfig.Database != "postgres" {
		getSql := fmt.Sprintf(
			`
			SELECT count(*) FROM pg_database WHERE datname = '%s';
		`, config.PgxConfig.Database,
		)
		var count int
		err := pg.QueryRow(ctx, getSql).Scan(&count)
		if err != nil {
			fmt.Println(color.RedString("Error checking if database exists: %v", err))
			return err
		}
		if count > 0 {
			return nil
		}
		fmt.Println("Creating database " + config.PgxConfig.Database + "...")
		sql := fmt.Sprintf("CREATE DATABASE %s;", config.PgxConfig.Database)
		_, err = pg.Exec(ctx, sql)
		if err != nil {
			fmt.Println(color.RedString("Error creating database: %v", err))
			return err
		}
	}
	return nil
}

func createUser(ctx context.Context, config ModuleConfig, pg *pgxpool.Pool) error {
	if config.PgxConfig.User != "postgres" {
		sql := fmt.Sprintf(
			`DO
$$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = '%s') THEN
      CREATE USER %s WITH PASSWORD '%s';
      GRANT ALL PRIVILEGES ON DATABASE %s TO %s;
   END IF;
END
$$;`,
			config.PgxConfig.User,
			config.PgxConfig.User,
			config.PgxConfig.Password,
			config.PgxConfig.Database,
			config.PgxConfig.User,
		)
		_, err := pg.Exec(ctx, sql)
		if err != nil {
			fmt.Println(color.RedString("Error creating user: %v", err))
			return err
		}
	}
	return nil
}

func NewEmbeddedPostgres(config ModuleConfig) *embeddedpostgres.EmbeddedPostgres {
	pwd := "postgres"
	if config.PgxConfig.Password != "" && config.PgxConfig.User == "postgres" {
		pwd = config.PgxConfig.Password
	}
	return embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().
			Password(pwd).
			Version(embeddedpostgres.V16).
			RuntimePath(config.RuntimePath).
			DataPath(config.DataPath).
			Port(uint32(config.PgxConfig.Port)).
			StartTimeout(45 * time.Second).
			StartParameters(map[string]string{"max_connections": "200"}),
	)
}

func NewModule() *module.Module {
	return module.NewModule("embedded pg").
		AddDependencies(
			pgx.NewModule(),
			migrator.NewModule(),
		).
		AddProviders(
			NewEmbeddedPostgres,
		).
		AddInvokes(
			registerHooks,
		).
		InitConfig(&ModuleConfig{})
}

func NewManifestModule() module.ManifestModule {
	return module.NewManifestModule(
		NewModule(),
		"github.com/go-modulus/modulus/db/embedded",
		"A wrapper for the github.com/fergusstrange/embedded-postgres package to integrate it into the Modulus framework. This package starts the embedded postgres database and creates the user and the database mentioned in PG_* vars. It works together with the github.com/go-modulus/modulus/db/pgx and migration packages.",
		"1.0.0",
	)
}
