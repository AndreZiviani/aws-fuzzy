module github.com/AndreZiviani/aws-fuzzy

go 1.14

require (
	github.com/HdrHistogram/hdrhistogram-go v1.0.1 // indirect
	github.com/aws/aws-sdk-go-v2 v1.2.0
	github.com/aws/aws-sdk-go-v2/config v1.1.1
	github.com/aws/aws-sdk-go-v2/credentials v1.1.1
	github.com/aws/aws-sdk-go-v2/service/configservice v1.1.1
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.1.1
	github.com/aws/aws-sdk-go-v2/service/networkmanager v1.1.1
	github.com/aws/aws-sdk-go-v2/service/sso v1.1.1
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.1.1
	github.com/faabiosr/cachego v0.16.1
	github.com/go-echarts/go-echarts/v2 v2.2.3
	github.com/jessevdk/go-flags v1.4.0
	github.com/ktr0731/go-fuzzyfinder v0.3.2
	github.com/mattn/go-sqlite3 v1.14.6
	github.com/nsf/termbox-go v0.0.0-20210114135735-d04385b850e8 // indirect
	github.com/opentracing/opentracing-go v1.2.0
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.4.0+incompatible
	go.uber.org/atomic v1.7.0 // indirect
	gopkg.in/ini.v1 v1.62.0
)

replace github.com/go-echarts/go-echarts/v2 => github.com/AndreZiviani/go-echarts/v2 v2.2.13
