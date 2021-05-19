module github.com/AndreZiviani/aws-fuzzy

go 1.14

require (
	github.com/AndreZiviani/fzf-wrapper/v2 v2.0.0-20210519173630-cf442a4d1d88
	github.com/HdrHistogram/hdrhistogram-go v1.0.1 // indirect
	github.com/aws/aws-sdk-go-v2 v1.5.0
	github.com/aws/aws-sdk-go-v2/config v1.2.0
	github.com/aws/aws-sdk-go-v2/credentials v1.2.0
	github.com/aws/aws-sdk-go-v2/service/configservice v1.5.0
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.7.0
	github.com/aws/aws-sdk-go-v2/service/networkmanager v1.2.0
	github.com/aws/aws-sdk-go-v2/service/sso v1.2.0
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.2.0
	github.com/faabiosr/cachego v0.16.1
	github.com/gdamore/tcell/v2 v2.3.3
	github.com/go-echarts/go-echarts/v2 v2.2.4
	github.com/jessevdk/go-flags v1.5.0
	github.com/junegunn/fzf v0.0.0-20210519094350-9ef825d2fd50 // indirect
	github.com/mattn/go-sqlite3 v1.14.7
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rivo/tview v0.0.0-20210514202809-22dbf8415b04
	github.com/smartystreets/assertions v1.0.0 // indirect
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/uber/jaeger-client-go v2.28.0+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/sys v0.0.0-20210514084401-e8d321eab015 // indirect
	golang.org/x/term v0.0.0-20210503060354-a79de5458b56 // indirect
	gopkg.in/ini.v1 v1.62.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200605160147-a5ece683394c // indirect
)

//replace github.com/go-echarts/go-echarts/v2 => ../go-echarts

replace github.com/go-echarts/go-echarts/v2 => github.com/AndreZiviani/go-echarts/v2 v2.2.17
