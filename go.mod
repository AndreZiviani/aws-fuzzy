module github.com/AndreZiviani/aws-fuzzy

go 1.19

require (
	github.com/99designs/keyring v1.2.2
	github.com/AlecAivazis/survey/v2 v2.3.6
	github.com/AndreZiviani/fzf-wrapper/v2 v2.0.0-20220531134234-4dd6b5a9c480
	github.com/BurntSushi/toml v1.2.1
	github.com/aws/aws-sdk-go-v2 v1.18.0
	github.com/aws/aws-sdk-go-v2/config v1.18.24
	github.com/aws/aws-sdk-go-v2/credentials v1.13.23
	github.com/aws/aws-sdk-go-v2/service/configservice v1.32.0
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.97.0
	github.com/aws/aws-sdk-go-v2/service/networkmanager v1.17.12
	github.com/aws/aws-sdk-go-v2/service/resourceexplorer2 v1.0.3
	github.com/aws/aws-sdk-go-v2/service/ssm v1.36.4
	github.com/aws/aws-sdk-go-v2/service/sso v1.12.10
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.14.10
	github.com/aws/aws-sdk-go-v2/service/sts v1.19.0
	github.com/common-fate/clio v1.1.0
	github.com/common-fate/granted v0.11.1
	github.com/gdamore/tcell/v2 v2.6.0
	github.com/gjbae1212/go-wraperror v0.7.0
	github.com/go-echarts/go-echarts/v2 v2.2.6
	github.com/hako/durafmt v0.0.0-20210608085754-5c1018a4e16b
	github.com/jessevdk/go-flags v1.5.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/rivo/tview v0.0.0-20230504092913-51ba3688bcdd
	github.com/segmentio/ksuid v1.0.4
	github.com/uber/jaeger-client-go v2.30.0+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible
	github.com/urfave/cli/v2 v2.25.3
	golang.org/x/text v0.9.0
	gopkg.in/ini.v1 v1.67.0
)

require (
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/HdrHistogram/hdrhistogram-go v1.0.1 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.13.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.33 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.27 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.27 // indirect
	github.com/aws/smithy-go v1.13.5 // indirect
	github.com/common-fate/awsconfigfile v0.8.0 // indirect
	github.com/common-fate/useragent v0.1.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/danieljoos/wincred v1.1.2 // indirect
	github.com/dvsekhvalnov/jose2go v1.5.0 // indirect
	github.com/fatih/color v1.15.0 // indirect
	github.com/gdamore/encoding v1.0.0 // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/junegunn/fzf v0.0.0-20230505063303-94999101e358 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.18 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/mattn/go-shellwords v1.0.12 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/saracen/walker v0.1.3 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/sync v0.2.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/term v0.8.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

//replace github.com/go-echarts/go-echarts/v2 => ../go-echarts

replace github.com/go-echarts/go-echarts/v2 => github.com/AndreZiviani/go-echarts/v2 v2.2.17

//replace github.com/common-fate/granted => ../granted-fork

//replace github.com/AndreZiviani/fzf-wrapper/v2 => /home/andre/git/fzf-wrapper
