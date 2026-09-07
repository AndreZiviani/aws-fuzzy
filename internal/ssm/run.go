package ssm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsssm "github.com/aws/aws-sdk-go-v2/service/ssm"
	awsssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/urfave/cli/v2"
)

const (
	// SSM truncates inline invocation output at 24000 bytes
	maxInlineOutput = 24000

	// exit code used by timeout(1) when the command did not finish in time
	exitCodeTimeout = 124

	pollInterval = time.Second

	// how long to keep polling after the command's own timeout expires, to
	// give the agent a chance to report the result
	pollGrace = 30 * time.Second
)

func NewRun(profile, region, instance, command string, timeout int, nonInteractive bool) *Run {
	run := Run{
		Profile:        profile,
		Region:         region,
		Instance:       instance,
		Command:        command,
		Timeout:        timeout,
		NonInteractive: nonInteractive,
	}

	return &run
}

// DoRun sends the command to a single instance and waits for it to finish,
// printing its output and returning an error carrying the remote exit code.
func (p *Run) DoRun(ctx context.Context, cfg aws.Config, id string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ssmrun")
	defer span.Finish()

	ssmclient := awsssm.NewFromConfig(cfg)

	command, err := ssmclient.SendCommand(ctx, &awsssm.SendCommandInput{
		DocumentName:   aws.String(docRunShellScript),
		InstanceIds:    []string{id},
		TimeoutSeconds: aws.Int32(int32(p.Timeout)),
		Parameters: map[string][]string{
			"commands": {p.Command},
		},
	})
	if err != nil {
		return err
	}

	invocation, err := p.waitForCommand(ctx, ssmclient, aws.ToString(command.Command.CommandId), id)
	if err != nil {
		return err
	}

	return printInvocation(invocation)
}

// waitForCommand polls until the invocation reaches a terminal status.
//
// The SDK's CommandExecutedWaiter is deliberately not used here: it returns an
// error on a Failed invocation and discards the output, and printing the
// remote stderr of a failed command is the whole point of this subcommand.
func (p *Run) waitForCommand(ctx context.Context, ssmclient *awsssm.Client, commandID, id string) (*awsssm.GetCommandInvocationOutput, error) {
	deadline := time.Now().Add(time.Duration(p.Timeout)*time.Second + pollGrace)

	input := &awsssm.GetCommandInvocationInput{
		CommandId:  aws.String(commandID),
		InstanceId: aws.String(id),
	}

	for {
		invocation, err := ssmclient.GetCommandInvocation(ctx, input)
		if err != nil {
			// the invocation is not visible immediately after SendCommand
			var notFound *awsssmtypes.InvocationDoesNotExist
			if !errors.As(err, &notFound) {
				return nil, err
			}
		} else if isTerminal(invocation.Status) {
			return invocation, nil
		}

		if time.Now().After(deadline) {
			return nil, cli.Exit(
				fmt.Sprintf("timed out waiting for command %s on %s", commandID, id),
				exitCodeTimeout,
			)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}

func isTerminal(status awsssmtypes.CommandInvocationStatus) bool {
	switch status {
	case awsssmtypes.CommandInvocationStatusSuccess,
		awsssmtypes.CommandInvocationStatusFailed,
		awsssmtypes.CommandInvocationStatusCancelled,
		awsssmtypes.CommandInvocationStatusTimedOut:
		return true
	}

	return false
}

// printInvocation writes the command output to our own stdout and stderr and
// translates the remote result into an exit code.
func printInvocation(invocation *awsssm.GetCommandInvocationOutput) error {
	stdout := aws.ToString(invocation.StandardOutputContent)
	stderr := aws.ToString(invocation.StandardErrorContent)

	fmt.Fprint(os.Stdout, stdout)
	fmt.Fprint(os.Stderr, stderr)

	if len(stdout) >= maxInlineOutput || len(stderr) >= maxInlineOutput {
		fmt.Fprintf(os.Stderr, "\naws-fuzzy: output truncated by SSM at %d bytes\n", maxInlineOutput)
	}

	if invocation.Status == awsssmtypes.CommandInvocationStatusTimedOut {
		return cli.Exit("aws-fuzzy: command timed out on the instance", exitCodeTimeout)
	}

	code := int(invocation.ResponseCode)
	if code == 0 && invocation.Status != awsssmtypes.CommandInvocationStatusSuccess {
		// cancelled, or failed without reporting a code
		code = 1
	}

	if code != 0 {
		return cli.Exit("", code)
	}

	return nil
}

func (p *Run) Execute(ctx context.Context) error {
	closer, err := tracing.InitTracing()
	if err != nil {
		fmt.Printf("failed to initialize tracing, %s\n", err)
	}
	defer func() { _ = closer.Close() }()

	tracer := opentracing.GlobalTracer()
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "ssm")

	if err := validateTarget(p.Instance, p.NonInteractive); err != nil {
		return err
	}

	cfg, err := newConfig(ctx, p.Profile, p.Region)
	if err != nil {
		return err
	}

	instance, err := resolveInstance(ctx, cfg, p.Instance, p.NonInteractive)
	if err != nil {
		return err
	}

	span.Finish()

	return p.DoRun(ctx, cfg, aws.ToString(instance.InstanceId))
}
