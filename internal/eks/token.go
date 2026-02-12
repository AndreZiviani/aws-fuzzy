package eks

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/AndreZiviani/aws-fuzzy/internal/securestorage"
	"github.com/AndreZiviani/aws-fuzzy/internal/sso"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/common-fate/clio"
)

const (
	tokenPrefix      = "k8s-aws-v1."
	tokenTTL         = 15 * time.Minute
	presignedURLTTL  = 60 // seconds, must be part of the signed URL for EKS to accept it
	clusterIDHeader  = "x-k8s-aws-id"
)

// ExecCredential matches the Kubernetes client.authentication.k8s.io/v1beta1 ExecCredential format
type ExecCredential struct {
	Kind       string               `json:"kind"`
	APIVersion string               `json:"apiVersion"`
	Spec       ExecCredentialSpec   `json:"spec"`
	Status     ExecCredentialStatus `json:"status"`
}

type ExecCredentialSpec struct{}

type ExecCredentialStatus struct {
	ExpirationTimestamp string `json:"expirationTimestamp"`
	Token               string `json:"token"`
}

func NewGetToken(profile, clusterName, region string, noCache, verbose bool) *GetToken {
	return &GetToken{
		Profile:     profile,
		ClusterName: clusterName,
		Region:      region,
		NoCache:     noCache,
		Verbose:     verbose,
	}
}

func (g *GetToken) Execute(ctx context.Context) error {
	if g.Verbose {
		clio.SetLevelFromString("debug")
	}

	// Check cache first
	cacheKey := g.Profile + ":" + g.ClusterName
	tokenStore := securestorage.NewSecureEKSTokenStorage()

	if !g.NoCache {
		cached, err := tokenStore.GetValidEKSToken(cacheKey)
		if err == nil && cached != nil {
			return g.outputExecCredential(cached.Token, cached.Expiration)
		}
	}

	// Get AWS credentials using the existing SSO login flow
	login := sso.Login{Profile: g.Profile}
	creds, err := login.GetCredentials(ctx)
	if err != nil {
		return fmt.Errorf("failed to get AWS credentials: %w", err)
	}

	// Build AWS config with credentials and optional region
	configOpts := []func(*config.LoadOptions) error{}
	if g.Region != "" {
		configOpts = append(configOpts, config.WithRegion(g.Region))
	}
	cfg, err := sso.NewAwsConfig(ctx, creds, configOpts...)
	if err != nil {
		return fmt.Errorf("failed to create AWS config: %w", err)
	}

	// Generate the EKS token
	token, expiration, err := g.generateToken(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to generate EKS token: %w", err)
	}

	// Cache the token
	tokenStore.StoreEKSToken(cacheKey, securestorage.EKSToken{
		Token:      token,
		Expiration: expiration,
	})

	return g.outputExecCredential(token, expiration)
}

func (g *GetToken) generateToken(ctx context.Context, cfg aws.Config) (string, time.Time, error) {
	stsClient := sts.NewFromConfig(cfg)
	presignClient := sts.NewPresignClient(stsClient)

	presigned, err := presignClient.PresignGetCallerIdentity(ctx,
		&sts.GetCallerIdentityInput{},
		func(po *sts.PresignOptions) {
			po.ClientOptions = append(po.ClientOptions, func(o *sts.Options) {
				o.APIOptions = append(o.APIOptions, addEKSPresignMiddleware(g.ClusterName))
			})
		},
	)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("presigning GetCallerIdentity: %w", err)
	}

	// Parse the expiration from the presigned URL
	expiration, err := parsePresignedURLExpiration(presigned.URL)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("parsing presigned URL expiration: %w", err)
	}

	// Base64url-encode the presigned URL (no padding)
	encoded := base64.RawURLEncoding.EncodeToString([]byte(presigned.URL))
	token := tokenPrefix + encoded

	return token, expiration, nil
}

func (g *GetToken) outputExecCredential(token string, expiration time.Time) error {
	cred := ExecCredential{
		Kind:       "ExecCredential",
		APIVersion: "client.authentication.k8s.io/v1beta1",
		Spec:       ExecCredentialSpec{},
		Status: ExecCredentialStatus{
			ExpirationTimestamp: expiration.UTC().Format(time.RFC3339),
			Token:               token,
		},
	}

	out, err := json.MarshalIndent(cred, "", "    ")
	if err != nil {
		return fmt.Errorf("marshalling ExecCredential: %w", err)
	}

	fmt.Println(string(out))
	return nil
}

// parsePresignedURLExpiration extracts the signing time from the presigned URL's
// X-Amz-Date parameter and adds the known EKS token TTL.
func parsePresignedURLExpiration(presignedURL string) (time.Time, error) {
	parsed, err := url.Parse(presignedURL)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing URL: %w", err)
	}

	amzDate := parsed.Query().Get("X-Amz-Date")
	if amzDate == "" {
		return time.Time{}, fmt.Errorf("missing X-Amz-Date in presigned URL")
	}

	signTime, err := time.Parse("20060102T150405Z", amzDate)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing X-Amz-Date: %w", err)
	}

	return signTime.Add(tokenTTL), nil
}

// addEKSPresignMiddleware adds the x-k8s-aws-id header and X-Amz-Expires query
// parameter to the STS request before signing, so they become part of the
// presigned URL's signature. Both are required by the EKS authenticator.
func addEKSPresignMiddleware(clusterName string) func(*middleware.Stack) error {
	return func(stack *middleware.Stack) error {
		return stack.Build.Add(middleware.BuildMiddlewareFunc(
			"AddEKSPresignParams",
			func(ctx context.Context, in middleware.BuildInput, next middleware.BuildHandler) (
				middleware.BuildOutput, middleware.Metadata, error,
			) {
				req, ok := in.Request.(*smithyhttp.Request)
				if ok {
					req.Header.Set(clusterIDHeader, clusterName)
					query := req.URL.Query()
					query.Set("X-Amz-Expires", strconv.Itoa(presignedURLTTL))
					req.URL.RawQuery = query.Encode()
				}
				return next.HandleBuild(ctx, in)
			},
		), middleware.Before)
	}
}
