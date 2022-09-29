module github.com/AndreZiviani/aws-fuzzy

go 1.17

require (
	github.com/AndreZiviani/fzf-wrapper/v2 v2.0.0-20220531134234-4dd6b5a9c480
	github.com/aws/aws-sdk-go-v2 v1.16.16
	github.com/aws/aws-sdk-go-v2/config v1.17.7
	github.com/aws/aws-sdk-go-v2/credentials v1.12.20
	github.com/aws/aws-sdk-go-v2/service/configservice v1.26.1
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.61.0
	github.com/aws/aws-sdk-go-v2/service/networkmanager v1.15.5
	github.com/aws/aws-sdk-go-v2/service/ssm v1.30.0
	github.com/aws/aws-sdk-go-v2/service/sso v1.11.23
	github.com/aws/aws-sdk-go-v2/service/sts v1.16.19
	github.com/common-fate/granted v0.3.1
	github.com/faabiosr/cachego v0.17.0
	github.com/gdamore/tcell/v2 v2.5.3
	github.com/gjbae1212/go-wraperror v0.7.0
	github.com/go-echarts/go-echarts/v2 v2.2.4
	github.com/jessevdk/go-flags v1.5.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/rivo/tview v0.0.0-20220916081518-2e69b7385a37
	github.com/uber/jaeger-client-go v2.30.0+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible
	gopkg.in/ini.v1 v1.67.0
)

require (
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/99designs/keyring v1.2.1 // indirect
	github.com/AlecAivazis/survey/v2 v2.3.6 // indirect
	github.com/BurntSushi/toml v1.2.0 // indirect
	github.com/HdrHistogram/hdrhistogram-go v1.0.1 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.24 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.13.5 // indirect
	github.com/aws/smithy-go v1.13.3 // indirect
	github.com/bigkevmcd/go-configparser v0.0.0-20220905190504-3f080373a501 // indirect
	github.com/common-fate/cf-protos v0.0.0-20220401154300-ca50a01c2dba // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/danieljoos/wincred v1.1.2 // indirect
	github.com/dvsekhvalnov/jose2go v1.5.0 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/gdamore/encoding v1.0.0 // indirect
	github.com/gdamore/tcell v1.4.0 // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/hako/durafmt v0.0.0-20210608085754-5c1018a4e16b // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/joho/godotenv v1.4.0 // indirect
	github.com/junegunn/fzf v0.0.0-20220928142231-04d0b0223fd6 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/mattn/go-shellwords v1.0.12 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rivo/uniseg v0.4.2 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/saracen/walker v0.1.3 // indirect
	github.com/segmentio/ksuid v1.0.4 // indirect
	github.com/urfave/cli/v2 v2.17.1 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.23.0 // indirect
	golang.org/x/net v0.0.0-20220927171203-f486391704dc // indirect
	golang.org/x/sync v0.0.0-20220923202941-7f9b1623fab7 // indirect
	golang.org/x/sys v0.0.0-20220928140112-f11e5e49a4ec // indirect
	golang.org/x/term v0.0.0-20220919170432-7a66f970e087 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220929141241-1ce7b20da813 // indirect
	google.golang.org/grpc v1.49.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)

//replace github.com/go-echarts/go-echarts/v2 => ../go-echarts

replace github.com/go-echarts/go-echarts/v2 => github.com/AndreZiviani/go-echarts/v2 v2.2.17

//replace github.com/AndreZiviani/fzf-wrapper/v2 => /home/andre/git/fzf-wrapper
