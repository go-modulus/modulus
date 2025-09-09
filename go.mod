module github.com/go-modulus/modulus

go 1.25

toolchain go1.25.0

replace github.com/vorlif/xspreak => github.com/go-modulus/xspreak v0.6.0

require (
	braces.dev/errtrace v0.4.0
	github.com/99designs/gqlgen v0.17.78
	github.com/amacneil/dbmate/v2 v2.28.0
	github.com/brianvoe/gofakeit/v7 v7.6.0
	github.com/c2h5oh/datasize v0.0.0-20231215233829-aa82cc1e6500
	github.com/fatih/color v1.18.0
	github.com/fatih/structs v1.1.0
	github.com/fergusstrange/embedded-postgres v1.32.0
	github.com/ggicci/httpin v0.20.1
	github.com/gkampitakis/go-snaps v0.5.14
	github.com/go-chi/chi/v5 v5.2.3
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/graph-gophers/dataloader/v7 v7.1.2
	github.com/iancoleman/strcase v0.3.0
	github.com/jackc/pgx/v5 v5.7.6
	github.com/jonboulle/clockwork v0.5.0
	github.com/laher/mergefs v0.1.1
	github.com/manifoldco/promptui v0.9.0
	github.com/rakyll/gotest v0.0.6
	github.com/ravilushqa/otelgqlgen v0.19.0
	github.com/rs/cors v1.11.1
	github.com/rs/xid v1.6.0
	github.com/samber/slog-formatter v1.2.0
	github.com/samber/slog-multi v1.5.0
	github.com/samber/slog-zap/v2 v2.6.2
	github.com/sethvargo/go-envconfig v1.3.0
	github.com/stretchr/testify v1.11.1
	github.com/subosito/gotenv v1.6.0
	github.com/thanhpk/randstr v1.0.6
	github.com/urfave/cli/v2 v2.27.7
	github.com/vektah/gqlparser/v2 v2.5.30
	github.com/vektra/mockery/v2 v2.53.5
	github.com/vorlif/spreak v0.6.0
	github.com/xinguang/go-recaptcha v1.0.1
	go.temporal.io/sdk v1.36.0
	go.temporal.io/sdk/contrib/opentelemetry v0.6.0
	go.uber.org/fx v1.24.0
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.42.0
	golang.org/x/oauth2 v0.31.0
	golang.org/x/text v0.29.0
	golang.org/x/tools v0.36.0
	google.golang.org/grpc v1.75.0
	gopkg.in/guregu/null.v4 v4.0.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cloud.google.com/go/compute/metadata v0.8.0 // indirect
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/chigopher/pathlib v0.19.1 // indirect
	github.com/chzyer/readline v1.5.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/ggicci/owl v0.8.2 // indirect
	github.com/gkampitakis/ciinfo v0.3.2 // indirect
	github.com/gkampitakis/go-diff v1.3.2 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/goccy/go-yaml v1.18.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.2 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/copier v0.4.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/maruel/natural v1.1.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-sqlite3 v1.14.32 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/nexus-rpc/sdk-go v0.4.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/locafero v0.10.0 // indirect
	github.com/samber/lo v1.51.0 // indirect
	github.com/samber/slog-common v0.19.0 // indirect
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/sosodev/duration v1.3.1 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.14.0 // indirect
	github.com/spf13/cast v1.9.2 // indirect
	github.com/spf13/cobra v1.9.1 // indirect
	github.com/spf13/pflag v1.0.7 // indirect
	github.com/spf13/viper v1.20.1 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	github.com/xrash/smetrics v0.0.0-20250705151800-55b8f293f342 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib v1.37.0 // indirect
	go.opentelemetry.io/otel v1.37.0 // indirect
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.37.0 // indirect
	go.temporal.io/api v1.52.0 // indirect
	go.uber.org/dig v1.19.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/mod v0.27.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/term v0.35.0 // indirect
	golang.org/x/time v0.12.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250818200422-3122310a409c // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250818200422-3122310a409c // indirect
	google.golang.org/protobuf v1.36.8 // indirect
)
