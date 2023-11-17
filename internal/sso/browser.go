package sso

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AndreZiviani/aws-fuzzy/internal/afconfig"
	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	"github.com/common-fate/clio"
	gbrowser "github.com/common-fate/granted/pkg/browser"
	"github.com/common-fate/granted/pkg/testable"
	opentracing "github.com/opentracing/opentracing-go"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func NewBrowser(browser string, ssoBrowser string, verbose bool) *Browser {
	b := Browser{
		Browser:    browser,
		SSOBrowser: ssoBrowser,
		Verbose:    verbose,
	}

	return &b
}

func (p *Browser) Execute(ctx context.Context) error {
	closer, err := tracing.InitTracing()
	if err != nil {
		fmt.Printf("failed to initialize tracing, %s\n", err)
	}
	defer closer.Close()

	tracer := opentracing.GlobalTracer()
	spanSso, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "ssobrowsercmd")
	defer spanSso.Finish()

	withStdio := survey.WithStdio(os.Stdin, os.Stderr, os.Stderr)
	browser := p.Browser
	ssoBrowser := p.SSOBrowser

	conf, err := afconfig.NewLoadedConfig()
	if err != nil {
		return err
	}

	if browser == "" {
		browser, err = gbrowser.HandleManualBrowserSelection()
		if err != nil {
			return err
		}

		err, browserKey, browserPath := p.GetBrowserSelection(withStdio, browser)
		if err != nil {
			return err
		}

		conf.DefaultBrowser = browserKey
		conf.CustomBrowserPath = browserPath

		err = conf.Save()
		if err != nil {
			return err
		}

		clio.Successf("aws-fuzzy will default to using %s", browserKey)
	}

	if ssoBrowser == "" {
		bpIn := survey.Confirm{
			Message: "Use a different browser than your default browser for SSO login?",
			Default: false,
			Help:    "For example, if you normally use a password manager in Chrome for your AWS login but Chrome is not your default browser, you would choose to use Chrome for SSO logins.",
		}
		var confirm bool
		err := testable.AskOne(&bpIn, &confirm, withStdio)
		if err != nil {
			return err
		}

		browserPath := ""
		browserKey := ""

		if confirm {
			ssoBrowser, err = gbrowser.HandleManualBrowserSelection()
			if err != nil {
				return err
			}

			err, browserKey, browserPath = p.GetBrowserSelection(withStdio, ssoBrowser)
			if err != nil {
				return err
			}

			clio.Successf("aws-fuzzy will use %s for SSO login prompts.", browserKey)
		}

		conf.CustomSSOBrowserPath = browserPath

		err = conf.Save()
		if err != nil {
			return err
		}
	}

	return nil

}

func (p *Browser) GetBrowserSelection(stdio survey.AskOpt, browserName string) (error, string, string) {
	var browserPath string
	browserKey := gbrowser.GetBrowserKey(browserName)
	title := cases.Title(language.AmericanEnglish)
	browserTitle := title.String(strings.ToLower(browserKey))

	// detect installation
	if browserKey != gbrowser.FirefoxStdoutKey && browserKey != gbrowser.StdoutKey {

		customBrowserPath, detected := gbrowser.DetectInstallation(browserKey)
		if !detected {
			clio.Warnf("aws-fuzzy could not detect an existing installation of %s at known installation paths for your system", browserTitle)
			clio.Info("If you have already installed this browser, you can specify the path to the executable manually")
			validPath := false
			for !validPath {
				// prompt for custom path
				bpIn := survey.Input{Message: fmt.Sprintf("Please enter the full path to your browser installation for %s:", browserTitle)}
				clio.NewLine()
				err := testable.AskOne(&bpIn, &customBrowserPath, stdio)
				if err != nil {
					return err, "", ""
				}
				if _, err := os.Stat(customBrowserPath); err == nil {
					validPath = true
				} else {
					clio.Error("The path you entered is not valid")
				}
			}
		}
		browserPath = customBrowserPath

		if browserKey == gbrowser.FirefoxKey {
			err := gbrowser.RunFirefoxExtensionPrompts(browserPath, browserName)
			if err != nil {
				return err, "", ""
			}
		}
	}

	return nil, browserKey, browserPath
}
