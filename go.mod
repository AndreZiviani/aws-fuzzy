module github.com/AndreZiviani/aws-fuzzy

go 1.17

require (
	github.com/AndreZiviani/fzf-wrapper/v2 v2.0.0-20210523171949-c28ead4a0f86
	github.com/aws/aws-sdk-go-v2 v1.6.0
	github.com/aws/aws-sdk-go-v2/config v1.3.0
	github.com/aws/aws-sdk-go-v2/credentials v1.2.1
	github.com/aws/aws-sdk-go-v2/service/configservice v1.5.1
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.7.1
	github.com/aws/aws-sdk-go-v2/service/networkmanager v1.2.1
	github.com/aws/aws-sdk-go-v2/service/sso v1.2.1
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.2.1
	github.com/faabiosr/cachego v0.16.1
	github.com/gdamore/tcell/v2 v2.3.3
	github.com/go-echarts/go-echarts/v2 v2.2.4
	github.com/jessevdk/go-flags v1.5.0
	github.com/mattn/go-sqlite3 v1.14.7
	github.com/opentracing/opentracing-go v1.2.0
	github.com/rivo/tview v0.0.0-20210521091241-1fd4a5b7aab3
	github.com/uber/jaeger-client-go v2.29.0+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible
	gopkg.in/ini.v1 v1.62.0
)

require (
	github.com/HdrHistogram/hdrhistogram-go v1.0.1 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.4.1 // indirect
	github.com/aws/smithy-go v1.4.0 // indirect
	github.com/fatih/color v1.11.0 // indirect
	github.com/gdamore/encoding v1.0.0 // indirect
	github.com/gdamore/tcell v1.4.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/junegunn/fzf v0.0.0-20210523023155-7e5aa1e2a5d7 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mattn/go-runewidth v0.0.12 // indirect
	github.com/mattn/go-shellwords v1.0.11 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/saracen/walker v0.1.2 // indirect
	github.com/smartystreets/assertions v1.0.0 // indirect
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.6.1 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210521203332-0cec03c779c1 // indirect
	golang.org/x/term v0.0.0-20210503060354-a79de5458b56 // indirect
	golang.org/x/text v0.3.6 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200605160147-a5ece683394c // indirect
)

//replace github.com/go-echarts/go-echarts/v2 => ../go-echarts

replace github.com/go-echarts/go-echarts/v2 => github.com/AndreZiviani/go-echarts/v2 v2.2.17

//replace github.com/AndreZiviani/fzf-wrapper/v2 => /home/andre/git/fzf-wrapper
