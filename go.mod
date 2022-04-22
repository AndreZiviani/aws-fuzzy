module github.com/AndreZiviani/aws-fuzzy

go 1.17

require (
	github.com/AndreZiviani/fzf-wrapper/v2 v2.0.0-20210523171949-c28ead4a0f86
	github.com/aws/aws-sdk-go-v2 v1.14.0
	github.com/aws/aws-sdk-go-v2/config v1.13.1
	github.com/aws/aws-sdk-go-v2/credentials v1.8.0
	github.com/aws/aws-sdk-go-v2/service/configservice v1.5.1
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.7.1
	github.com/aws/aws-sdk-go-v2/service/networkmanager v1.2.1
	github.com/aws/aws-sdk-go-v2/service/sso v1.9.0
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.10.0
	github.com/common-fate/granted v0.1.14
	github.com/faabiosr/cachego v0.16.1
	github.com/gdamore/tcell/v2 v2.3.3
	github.com/go-echarts/go-echarts/v2 v2.2.4
	github.com/jessevdk/go-flags v1.5.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/rivo/tview v0.0.0-20210521091241-1fd4a5b7aab3
	github.com/uber/jaeger-client-go v2.29.0+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible
	gopkg.in/ini.v1 v1.62.0
)

require (
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/99designs/keyring v1.2.0 // indirect
	github.com/AlecAivazis/survey/v2 v2.3.2 // indirect
	github.com/BurntSushi/toml v1.0.0 // indirect
	github.com/HdrHistogram/hdrhistogram-go v1.0.1 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.10.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.7.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.14.0 // indirect
	github.com/aws/smithy-go v1.11.0 // indirect
	github.com/bigkevmcd/go-configparser v0.0.0-20210106142102-909504547ead // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0-20190314233015-f79a8a8ca69d // indirect
	github.com/danieljoos/wincred v1.1.2 // indirect
	github.com/dvsekhvalnov/jose2go v1.5.0 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/gdamore/encoding v1.0.0 // indirect
	github.com/gdamore/tcell v1.4.0 // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/junegunn/fzf v0.0.0-20210523023155-7e5aa1e2a5d7 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.12 // indirect
	github.com/mattn/go-shellwords v1.0.11 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/saracen/walker v0.1.2 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/smartystreets/assertions v1.0.0 // indirect
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/urfave/cli/v2 v2.3.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/crypto v0.0.0-20220131195533-30dcbda58838 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20220224120231-95c6836cb0e7 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

//replace github.com/go-echarts/go-echarts/v2 => ../go-echarts

replace github.com/go-echarts/go-echarts/v2 => github.com/AndreZiviani/go-echarts/v2 v2.2.17

//replace github.com/AndreZiviani/fzf-wrapper/v2 => /home/andre/git/fzf-wrapper
