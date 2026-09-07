# SSM: direct instance targeting and non-interactive execution

Date: 2026-09-04
Status: approved, not yet implemented

## Problem

The `ssm` module always picks its target through the interactive fzf TUI.
Two consequences:

1. You cannot name an instance you already know. Every `ssm session` and
   `ssm portforward` invocation forces a trip through the picker.
2. The module cannot be scripted. There is no way to run a command on an
   instance and collect its output, and every existing subcommand blocks on a
   terminal UI.

## Goals

- Accept a target directly on `ssm session` and `ssm portforward`.
- Add `ssm run`: send one shell command to one instance, print its output,
  exit with the remote exit code.
- Keep every subcommand usable in a script, with no path that can block on a
  terminal.

## Non-goals

- Pagination of `GetInstances` (an existing `TODO` in `common.go`).
- Fan-out to several instances, or targeting by tag filter.
- Offloading command output to S3 or CloudWatch Logs.
- Any change to `tcell.go` or to the embedded session-manager-plugin path.

## Design

### 1. Shared AWS config helper

`session.go` and `portforward.go` each build the AWS config twice: once in
`Execute`, then again inside `DoSsm` / `DoPortForward`. A third subcommand
would make that three duplicates of the same six lines.

Add to `internal/ssm/common.go`:

```go
func newConfig(ctx context.Context, profile, region string) (aws.Config, error)
```

It wraps `sso.Login{Profile: profile}` -> `GetCredentials` -> `sso.NewAwsConfig`
with `config.WithRegion(region)`.

`DoSsm`, `DoPortForward` and the new `DoRun` take `cfg aws.Config` as a
parameter instead of building their own. Each `Execute` builds the config once
and passes it down.

### 2. Target resolution

New file `internal/ssm/target.go`.

```go
func resolveInstance(ctx context.Context, cfg aws.Config, query string, noTui bool) (*Instance, error)
func matchInstances(instances []Instance, query string) []Instance
```

`resolveInstance` behaviour:

- `query == ""`: call `GetInstances`, show the full TUI. This is today's
  behaviour, unchanged.
- `query != ""`: call `GetInstances`, then `matchInstances`. One hit is
  returned directly. Zero hits is an error naming the query. More than one hit
  opens the TUI seeded with only those candidates.
- `noTui == true`: any case that would open the TUI becomes an error instead.
  For zero hits the message names the query; for an ambiguous query it lists
  each candidate's instance id and `Name` tag so the caller can re-run against
  a specific id.

`matchInstances` is pure and does no I/O. It tries four tiers in order and
returns the first tier that has any hits:

1. exact match on instance id
2. exact match on private IP address
3. exact match on the `Name` tag, case-insensitive
4. substring match on the `Name` tag, case-insensitive

Fast path: when `query` matches `^i-[0-9a-f]+$`, `resolveInstance` passes it as
an `InstanceIds` filter to `DescribeInstanceInformation` rather than
enumerating every managed instance in the account. This requires
`GetInstances` to accept an optional instance-id filter; extend its signature
to `GetInstances(ctx, cfg, instanceIDs []string)` and pass `nil` from the
no-query path.

### 3. `tui()` signature

`tui()` currently takes `*ec2.DescribeInstancesOutput`, which prevents feeding
it a filtered subset. Change it to `tui(instances []Instance) (*Instance, error)`
and move the reservation flattening into a helper:

```go
func flattenInstances(out *ec2.DescribeInstancesOutput) []Instance
```

`NewFzfData` then takes `[]Instance`. `tcell.go` is untouched.

### 4. `ssm run`

New file `internal/ssm/run.go`, following the struct / `New*` / `Execute` /
`Do*` shape of the existing subcommands.

```
aws-fuzzy ssm run -i <id|name|ip> -c 'uptime' [--timeout 60] [--non-interactive]
```

`Run` struct fields: `Profile`, `Region`, `Instance`, `Command`, `Timeout`,
`NonInteractive`.

`DoRun` sends:

- `SendCommand` with `DocumentName: "AWS-RunShellScript"`,
  `Parameters: {"commands": [p.Command]}`, `InstanceIds: [id]`,
  `TimeoutSeconds: p.Timeout`.

Then it polls `GetCommandInvocation` in a hand-rolled loop until the
invocation reaches a terminal status (`Success`, `Failed`, `Cancelled`,
`TimedOut`). The SDK's `CommandExecutedWaiter` is deliberately not used: it
returns an error on `Failed` and discards the invocation output, and printing
stderr on a failed command is the primary purpose of this subcommand.

Poll loop details:

- The invocation is not visible immediately after `SendCommand`; treat
  `InvocationDoesNotExist` as "keep waiting", not as an error.
- Poll every second, capped by `p.Timeout` plus a small grace period for the
  agent to report the result.

Output and exit codes:

- `StandardOutputContent` to stdout, `StandardErrorContent` to stderr.
- Exit with the invocation's `ResponseCode`.
- A `TimedOut` invocation, or the poll loop exceeding its deadline, exits `124`
  to match `timeout(1)`.
- If either content field is at the 24 KB service cap, write a warning to
  stderr that the output was truncated.

Propagating an arbitrary remote exit code requires returning
`cli.Exit(msg, code)` (from `urfave/cli/v2`) rather than a plain error: a plain
error reaches `internal/cli/main.go`, which prints it and hardcodes
`os.Exit(1)`.

No change to `internal/cli/main.go` is needed. `urfave/cli` calls
`handleExitCoder` immediately after the subcommand's `Action` returns
(`command.go`), and for an `ExitCoder` that prints the message to stderr and
exits with the carried code, so `cli.Run`'s own `os.Exit(1)` is never reached.
Return `nil`, not `cli.Exit("", 0)`, on success.

### 5. CLI wiring

In `internal/ssm/main.go`:

- `session` and `portforward` gain
  `&cli.StringFlag{Name: "instance", Aliases: []string{"i"}}` and
  `&cli.BoolFlag{Name: "non-interactive"}`.
- New `run` subcommand with `profile`, `region`, `instance`,
  `non-interactive`, plus `&cli.StringFlag{Name: "command", Aliases: []string{"c"}, Required: true}`
  and `&cli.IntFlag{Name: "timeout", Value: 60}`.
- `NewSession` and `NewPortForward` gain the two new parameters; `NewRun` is
  added.

### 6. Error handling

- Resolution failures (zero hits, ambiguity under `--non-interactive`) return
  before any SSM API call, with a message that names what was searched for.
- A user aborting the TUI keeps returning the existing
  `"aborting by user request"` error.
- `SendCommand` rejecting the target (instance not registered with SSM) is
  surfaced as-is; resolution already guarantees the instance was in the
  `DescribeInstanceInformation` result, so this should be rare.

## Testing

The repository has no AWS mocking and only one existing test file
(`internal/awsprofile/pkce_test.go`), so tests target the pure logic:

- `matchInstances` table test over a fixture `[]Instance` covering: match by
  instance id, match by private IP, exact `Name` match, case-insensitive
  `Name` match, substring `Name` match, no match, and a multi-match that
  returns several candidates.
- Tier precedence: a query that is both an exact IP of one instance and a
  `Name` substring of another returns only the IP match.

The AWS-calling paths (`resolveInstance`, `DoRun`) stay manually verified,
consistent with the rest of the module.

## Documentation

`README.md` currently documents no `ssm` module at all. Add an `## SSM`
section covering `session`, `portforward` and `run`, including the 24 KB
output cap and the `--non-interactive` flag.
