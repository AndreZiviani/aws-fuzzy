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
	"github.com/pkg/errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (p *Browser) Execute(args []string) error {

	ctx := context.Background()

	closer, err := tracing.InitTracing()
	if err != nil {
		fmt.Printf("failed to initialize tracing, %s\n", err)
	}
	defer closer.Close()

	tracer := opentracing.GlobalTracer()
	spanSso, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "ssobrowsercmd")
	defer spanSso.Finish()

	browser := p.Browser

	if browser == "" {
		browser, err = gbrowser.HandleManualBrowserSelection()
		if err != nil {
			return err
		}
	}

	return p.ConfigureBrowserSelection(browser, "")

}

func (p *Browser) ConfigureBrowserSelection(browserName string, path string) error {
	browserKey := gbrowser.GetBrowserKey(browserName)
	withStdio := survey.WithStdio(os.Stdin, os.Stderr, os.Stderr)
	title := cases.Title(language.AmericanEnglish)
	browserTitle := title.String(strings.ToLower(browserKey))
	// We allow users to configure a custom install path is we cannot detect the installation
	browserPath := path
	// detect installation
	if browserKey != gbrowser.FirefoxStdoutKey && browserKey != gbrowser.StdoutKey {

		if browserPath != "" {
			_, err := os.Stat(browserPath)
			if err != nil {
				return errors.Wrap(err, "provided path is invalid")
			}
		} else {
			customBrowserPath, detected := gbrowser.DetectInstallation(browserKey)
			if !detected {
				clio.Warnf("aws-fuzzy could not detect an existing installation of %s at known installation paths for your system", browserTitle)
				clio.Info("If you have already installed this browser, you can specify the path to the executable manually")
				validPath := false
				for !validPath {
					// prompt for custom path
					bpIn := survey.Input{Message: fmt.Sprintf("Please enter the full path to your browser installation for %s:", browserTitle)}
					clio.NewLine()
					err := testable.AskOne(&bpIn, &customBrowserPath, withStdio)
					if err != nil {
						return err
					}
					if _, err := os.Stat(customBrowserPath); err == nil {
						validPath = true
					} else {
						clio.Error("The path you entered is not valid")
					}
				}
			}
			browserPath = customBrowserPath
		}

		if browserKey == gbrowser.FirefoxKey {
			err := gbrowser.RunFirefoxExtensionPrompts(browserPath)
			if err != nil {
				return err
			}
		}
	}
	//save the detected browser as the default
	conf, err := afconfig.NewLoadedConfig()
	if err != nil {
		return err
	}

	conf.DefaultBrowser = browserKey
	conf.CustomBrowserPath = browserPath
	err = conf.Save()
	if err != nil {
		return err
	}
	clio.Successf("Granted will default to using %s", browserTitle)
	return nil
}
