package sso

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	opentracing "github.com/opentracing/opentracing-go"
)

// AWS Creds consumed by credential_process must adhere to this schema
// https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sourcing-external.html
type awsCredsStdOut struct {
	Version         int    `json:"Version"`
	AccessKeyID     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken,omitempty"`
	Expiration      string `json:"Expiration,omitempty"`
}

func NewCredentialProcess(profile, token string, verbose bool) *CredentialProcess {
	cp := CredentialProcess{
		Profile: profile,
		MFATOTP: token,
		Verbose: verbose,
	}

	return &cp
}

func (p *CredentialProcess) Execute(ctx context.Context) error {
	closer, err := tracing.InitTracing()
	if err != nil {
		fmt.Printf("failed to initialize tracing, %s\n", err)
	}
	defer closer.Close()

	tracer := opentracing.GlobalTracer()
	spanSso, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "ssocredentialprocesscmd")
	defer spanSso.Finish()

	login := Login{Profile: p.Profile}
	creds, err := login.GetCredentials(ctx)
	if err != nil {
		return err
	}

	out := awsCredsStdOut{
		Version:         1,
		AccessKeyID:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
		SessionToken:    creds.SessionToken,
	}
	if creds.CanExpire {
		out.Expiration = creds.Expires.Format(time.RFC3339)
	}

	jsonOut, err := json.Marshal(out)
	if err != nil {
		return fmt.Errorf("marshalling session credentials\n")
	}

	fmt.Println(string(jsonOut))
	return nil

}
