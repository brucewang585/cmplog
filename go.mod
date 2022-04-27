module github.com/brucewang585/cmplog

go 1.17

require (
	github.com/stretchr/testify v1.7.1
	github.com/zeromicro/go-zero v1.3.1
	go.opentelemetry.io/otel/sdk v1.6.3
	go.opentelemetry.io/otel/trace v1.6.3
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	go.opentelemetry.io/otel v1.6.3
	go.uber.org/automaxprocs v1.4.0 // indirect
	golang.org/x/sys v0.0.0-20220227234510-4e6760a101f9 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/zeromicro/go-zero v1.3.1 => github.com/tal-tech/go-zero v1.3.1

replace go.opentelemetry.io/otel v1.6.3 => github.com/open-telemetry/opentelemetry-go v1.6.3
