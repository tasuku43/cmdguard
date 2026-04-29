package semantic

var awsSchema = Schema{
	Command:     "aws",
	order:       20,
	Description: "AWS CLI service, operation, profile, region, endpoint, and dry-run matching.",
	Parser:      "aws",
	Fields: []Field{
		stringField("service", "AWS service name such as s3 or iam."),
		stringListField("service_in", "Allowed AWS services."),
		stringField("operation", "AWS operation name."),
		stringListField("operation_in", "Allowed AWS operations."),
		stringField("profile", "AWS profile selected by --profile or AWS_PROFILE."),
		stringListField("profile_in", "Allowed AWS profiles."),
		stringField("region", "AWS region selected by --region or environment."),
		stringListField("region_in", "Allowed AWS regions."),
		stringField("endpoint_url", "Exact --endpoint-url value."),
		stringField("endpoint_url_prefix", "--endpoint-url prefix."),
		boolField("dry_run", "True when --dry-run is present, false when --no-dry-run is present, and unset when neither form is recognized."),
		boolField("no_cli_pager", "True when --no-cli-pager is present."),
		stringListField("flags_contains", "Parser-recognized AWS option tokens that must be present; this does not scan raw argv words."),
		stringListField("flags_prefixes", "Parser-recognized AWS option tokens that must start with these prefixes; this depends on the AWS parser."),
	},
	Examples: []Example{
		{Title: "Ask for IAM writes", YAML: `permission:
  ask:
    - command:
        name: aws
        semantic:
          service: iam`},
	},
}

func init() {
	RegisterSchema(awsSchema)
}
