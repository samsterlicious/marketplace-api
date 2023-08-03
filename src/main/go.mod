module samster.link/main

go 1.20

require (
	github.com/aws/aws-lambda-go v1.41.0
	sammy.link/marketplace v0.0.0-00010101000000-000000000000
)

require (
	github.com/aws/aws-sdk-go v1.44.315 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
)

replace sammy.link/marketplace => ../marketplace
