module github.com/imharish-sivakumar/modern-oauth2-system/oauth2-clients-setup

go 1.22.2

require (
	github.com/imharish-sivakumar/modern-oauth2-system/aws-utils v0.0.0
	github.com/imharish-sivakumar/modern-oauth2-system/service-utils v0.0.0
)

replace (
	github.com/imharish-sivakumar/modern-oauth2-system/aws-utils v0.0.0 => ../../aws-utils
	github.com/imharish-sivakumar/modern-oauth2-system/service-utils v0.0.0 => ../../service-utils
	github.com/imharish-sivakumar/modern-oauth2-system/cisauth-proto => ../../cisauth-proto
)

require (
	github.com/aws/aws-sdk-go-v2 v1.32.3 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.28.1 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.42 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.18 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.22 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.22 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.34.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.24.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.28.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.32.3 // indirect
	github.com/aws/smithy-go v1.22.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.4 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.22.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/crypto v0.26.0 // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
	golang.org/x/text v0.17.0 // indirect
)
