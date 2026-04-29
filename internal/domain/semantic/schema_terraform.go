package semantic

var terraformSchema = Schema{
	Command:     "terraform",
	order:       80,
	Description: "Terraform subcommands, workspace/state subcommands, and high-risk infrastructure flags.",
	Parser:      "terraform",
	Fields: []Field{
		stringField("subcommand", "Terraform subcommand such as init, validate, plan, apply, destroy, state, or workspace."),
		stringListField("subcommand_in", "Allowed Terraform subcommands."),
		stringField("global_chdir", "Directory selected by global -chdir."),
		stringField("workspace_subcommand", "Workspace subcommand such as list, show, select, new, or delete."),
		stringListField("workspace_subcommand_in", "Allowed workspace subcommands."),
		stringField("state_subcommand", "State subcommand such as list, show, mv, rm, pull, push, or replace-provider."),
		stringListField("state_subcommand_in", "Allowed state subcommands."),
		boolField("target", "True when -target is present."),
		stringListField("targets_contains", "Targets selected by -target that must be present."),
		boolField("replace", "True when -replace is present."),
		stringListField("replaces_contains", "Replace addresses selected by -replace that must be present."),
		boolField("destroy", "True for terraform destroy or plan -destroy."),
		boolField("auto_approve", "True when -auto-approve is present."),
		boolField("input", "Value selected by -input=true/false."),
		boolField("lock", "Value selected by -lock=true/false."),
		boolField("refresh", "Value selected by -refresh=true/false."),
		boolField("refresh_only", "True when -refresh-only is present."),
		stringField("out", "Plan output selected by -out."),
		stringField("plan_file", "Plan file argument used by terraform apply <planfile>."),
		stringListField("var_files_contains", "Variable files selected by -var-file that must be present."),
		boolField("vars", "True when -var is present."),
		boolField("backend", "Value selected by terraform init -backend=true/false."),
		boolField("upgrade", "True when terraform init -upgrade is present."),
		boolField("reconfigure", "True when terraform init -reconfigure is present."),
		boolField("migrate_state", "True when terraform init -migrate-state is present."),
		boolField("recursive", "True when terraform fmt -recursive is present."),
		boolField("check", "True when terraform fmt -check is present."),
		boolField("json", "True when -json is present."),
		boolField("force", "True when a known -force flag is present."),
		stringListField("flags_contains", "Parser-recognized terraform option tokens that must be present; this does not scan raw argv words."),
		stringListField("flags_prefixes", "Parser-recognized terraform option tokens that must start with these prefixes; this depends on the terraform parser."),
	},
	Examples: []Example{
		{Title: "Allow read-only Terraform commands", YAML: `permission:
  allow:
    - command:
        name: terraform
        semantic:
          subcommand_in: [validate, plan, show, output]`},
		{Title: "Ask before apply", YAML: `permission:
  ask:
    - command:
        name: terraform
        semantic:
          subcommand: apply`},
		{Title: "Deny auto-approved destroy", YAML: `permission:
  deny:
    - command:
        name: terraform
        semantic:
          subcommand: destroy
          auto_approve: true`},
	},
}

func init() {
	RegisterSchema(terraformSchema)
}
