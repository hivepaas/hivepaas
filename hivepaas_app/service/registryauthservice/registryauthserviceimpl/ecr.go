package registryauthserviceimpl

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// ecrToken is a token ECR gave: the password docker logs in with, as user AWS,
// and when it stops working.
type ecrToken struct {
	password  string
	expiresAt time.Time
}

// ecrKeys is what signs in to ECR, as AWS takes it: the key auth's keys, and
// the credential's region and role. keyVer is the key auth's version.
type ecrKeys struct {
	region  string
	keyID   string
	secret  string
	roleARN string
	keyVer  int
}

// ecrTokens gets ECR tokens; the tests give a fake.
type ecrTokens interface {
	Token(ctx context.Context, keys *ecrKeys) (*ecrToken, error)
}

// awsECR asks AWS: GetAuthorizationToken with the keys, after assuming the role
// when one is named.
type awsECR struct{}

// ecrCallTimeout bounds one call to AWS, so that a deploy waiting on a token is
// not held by an AWS that does not answer.
const ecrCallTimeout = 30 * time.Second

var errNoAuthorizationData = errors.New("ECR answered no authorization data")

func (awsECR) Token(ctx context.Context, keys *ecrKeys) (*ecrToken, error) {
	ctx, cancel := context.WithTimeout(ctx, ecrCallTimeout)
	defer cancel()

	cfg := aws.Config{
		Region:      keys.region,
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(keys.keyID, keys.secret, "")),
	}
	if keys.roleARN != "" {
		cfg.Credentials = aws.NewCredentialsCache(stscreds.NewAssumeRoleProvider(sts.NewFromConfig(cfg), keys.roleARN,
			func(o *stscreds.AssumeRoleOptions) { o.RoleSessionName = "hivepaas-registry" }))
	}
	out, err := ecr.NewFromConfig(cfg).GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return nil, err //nolint:wrapcheck // worded by the caller, with the credential's address
	}
	if len(out.AuthorizationData) == 0 || out.AuthorizationData[0].AuthorizationToken == nil {
		return nil, errNoAuthorizationData
	}
	data := out.AuthorizationData[0]
	password, err := decodeECRToken(aws.ToString(data.AuthorizationToken))
	if err != nil {
		return nil, err
	}
	return &ecrToken{password: password, expiresAt: aws.ToTime(data.ExpiresAt)}, nil
}

var errMalformedToken = errors.New("ECR answered a token that is not base64 of AWS:<password>")

// decodeECRToken is the password in an ECR authorization token: base64 of
// "AWS:<password>".
func decodeECRToken(token string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", errMalformedToken
	}
	user, password, ok := strings.Cut(string(raw), ":")
	if !ok || user != ecrUsername || password == "" {
		return "", errMalformedToken
	}
	return password, nil
}
