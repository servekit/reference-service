module github.com/servekit/reference-service

go 1.26.6

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.12-20260825204119-511051f7f437.2 // indirect
	buf.build/go/protovalidate v1.2.0
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.3.3
	github.com/joho/godotenv v1.5.1
	github.com/robfig/cron/v3 v3.0.1
	// Contract repo (servekit mode). Pinned at ../api HEAD when scaffolded; the
	// new domain's generated code is not in that commit yet, so after pushing
	// ../api bump it: go get github.com/servekit/api/gen/go@<commit> && go mod tidy
	github.com/servekit/api/gen/go v0.0.0-20260907134616-c735fd847cc4
	// go-common carries the provider-contract stack (dbx/redisx.Connect,
	// grpcx/clienttest) and the security bumps; keep this in step with the
	// latest go-common main when the scaffold regenerates demo-service.
	github.com/servekit/go-common v0.0.0-20260907134535-5a601f9f7499
	github.com/stretchr/testify v1.11.1
	google.golang.org/genproto/googleapis/api v0.0.0-20260904194346-d0f1323225a4 // indirect
	google.golang.org/grpc v1.83.2
	google.golang.org/protobuf v1.36.12
)

require (
	github.com/nyaruka/phonenumbers v1.8.1
	golang.org/x/text v0.41.0
)

require (
	cel.dev/expr v0.25.2 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/google/cel-go v0.28.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.30.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/sagikazarmark/locafero v0.11.0 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spf13/viper v1.21.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.45.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/exp v0.0.0-20250813145105-42675adae3e6 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260825221802-da73d73af1c5 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
