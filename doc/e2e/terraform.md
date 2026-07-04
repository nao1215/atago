# atago Behavior Specs
## Summary
1 suite · 6 scenarios
## Contents
- [terraform (offline via the builtin terraform_data resource)](#terraform-offline-via-the-builtin-terraform_data-resource) — 6 scenarios
  - [init downloads nothing and validate reports valid](#scenario-init-downloads-nothing-and-validate-reports-valid)
  - [plan -detailed-exitcode reports the change contract](#scenario-plan--detailed-exitcode-reports-the-change-contract)
  - [apply exposes state JSON and a captured output](#scenario-apply-exposes-state-json-and-a-captured-output)
  - [destroy empties the state](#scenario-destroy-empties-the-state)
  - [fmt -check exits 3 on a misformatted file](#scenario-fmt--check-exits-3-on-a-misformatted-file)
  - [a broken configuration is rejected](#scenario-a-broken-configuration-is-rejected)
## terraform (offline via the builtin terraform_data resource)
Source: `test/e2e/thirdparty/terraform/terraform.atago.yaml`
### Scenario: init downloads nothing and validate reports valid
#### Given
- Fixture file `main.tf` is created.
#### Inputs
_Fixture `main.tf`:_
```text
resource "terraform_data" "greeting" {
  input = "hello from atago"
}
output "message" {
  value = terraform_data.greeting.output
}
```
#### When
```shell
terraform init
terraform validate -json
```
#### Then
- after `terraform init`:
  - exit code is `0`
  - stdout does not contain `Downloading`
- after `terraform validate -json`:
  - exit code is `0`
  - stdout at `$.valid` equals `true`
### Scenario: plan -detailed-exitcode reports the change contract
#### Given
- Fixture file `main.tf` is created.
#### Inputs
_Fixture `main.tf`:_
```text
resource "terraform_data" "greeting" {
  input = "hello from atago"
}
output "message" {
  value = terraform_data.greeting.output
}
```
#### When
```shell
terraform init
terraform plan -detailed-exitcode
terraform apply -auto-approve
terraform plan -detailed-exitcode
```
#### Then
- after `terraform init`:
  - exit code is `0`
- after `terraform plan -detailed-exitcode`:
  - exit code is one of `2`
- after `terraform apply -auto-approve`:
  - exit code is `0`
- after `terraform plan -detailed-exitcode`:
  - exit code is `0`
### Scenario: apply exposes state JSON and a captured output
#### Given
- Fixture file `main.tf` is created.
#### Inputs
_Fixture `main.tf`:_
```text
resource "terraform_data" "greeting" {
  input = "hello from atago"
}
output "message" {
  value = terraform_data.greeting.output
}
```
#### When
```shell
terraform init
terraform apply -auto-approve
terraform show -json
terraform output -raw message
# capture ${message} from stdout
echo captured: ${message}
```
#### Then
- after `terraform init`:
  - exit code is `0`
- after `terraform apply -auto-approve`:
  - exit code is `0`
- after `terraform show -json`:
  - exit code is `0`
  - stdout at `$.values.root_module.resources[0].address` equals `terraform_data.greeting`
- after `echo captured: ${message}`:
  - exit code is `0`
  - stdout contains `captured: hello from atago`
### Scenario: destroy empties the state
#### Given
- Fixture file `main.tf` is created.
#### Inputs
_Fixture `main.tf`:_
```text
resource "terraform_data" "greeting" {
  input = "hello from atago"
}
```
#### When
```shell
terraform init
terraform apply -auto-approve
terraform destroy -auto-approve
terraform show -json
```
#### Then
- after `terraform init`:
  - exit code is `0`
- after `terraform apply -auto-approve`:
  - exit code is `0`
- after `terraform destroy -auto-approve`:
  - exit code is `0`
- after `terraform show -json`:
  - exit code is `0`
  - stdout does not contain `terraform_data.greeting`
### Scenario: fmt -check exits 3 on a misformatted file
#### Given
- Fixture file `messy.tf` is created.
#### Inputs
_Fixture `messy.tf`:_
```text
output   "x" {
value="terraform_data"
}
```
#### When
```shell
terraform fmt -check messy.tf
```
#### Then
- exit code is one of `3`
### Scenario: a broken configuration is rejected
#### Given
- Fixture file `main.tf` is created.
#### Inputs
_Fixture `main.tf`:_
```text
resource "terraform_data" "broken" {
  input =
}
```
#### When
```shell
terraform plan
```
#### Then
- exit code is one of `1`
- stderr contains `Error:`