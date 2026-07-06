# atago Behavior Specs
## Summary
2 suites · 8 scenarios
## Contents
- [ecspresso + deploy (real ECS deploy against a mock)](#ecspresso--deploy-real-ecs-deploy-against-a-mock) — 1 scenario
  - [deploy registers the task definition, creates the service, and rolls it](#scenario-deploy-registers-the-task-definition-creates-the-service-and-rolls-it)
- [ecspresso (Amazon ECS deploy tool)](#ecspresso-amazon-ecs-deploy-tool) — 7 scenarios
  - [version prints without error](#scenario-version-prints-without-error)
  - [render substitutes an env template function](#scenario-render-substitutes-an-env-template-function)
  - [render falls back to the template default when the env is unset](#scenario-render-falls-back-to-the-template-default-when-the-env-is-unset)
  - [render evaluates a jsonnet task definition with an external variable](#scenario-render-evaluates-a-jsonnet-task-definition-with-an-external-variable)
  - [render config resolves the defaults](#scenario-render-config-resolves-the-defaults)
  - [an undefined must_env fails the render](#scenario-an-undefined-must_env-fails-the-render)
  - [a missing config file fails cleanly](#scenario-a-missing-config-file-fails-cleanly)
## ecspresso + deploy (real ECS deploy against a mock)
Source: `test/e2e/thirdparty/ecspresso/deploy.atago.yaml`
### Scenario: deploy registers the task definition, creates the service, and rolls it
#### Given
- Background service `moto` is started: `moto_server -H 127.0.0.1 -p 15111`.
- Fixture file `ecspresso.yml` is created.
- Fixture file `taskdef.json` is created.
- Fixture file `servicedef.json` is created.
#### Inputs
_Fixture `ecspresso.yml`:_
```text
region: us-east-1
cluster: demo
service: web
task_definition: taskdef.json
service_definition: servicedef.json
```
_Fixture `taskdef.json`:_
```text
{
  "family": "web",
  "containerDefinitions": [
    { "name": "web", "image": "nginx:1.0", "cpu": 128, "memory": 256, "essential": true }
  ]
}
```
_Fixture `servicedef.json`:_
```text
{ "launchType": "EC2", "desiredCount": 2, "schedulingStrategy": "REPLICA" }
```
#### When
```shell
aws --endpoint-url http://127.0.0.1:15111 ecs create-cluster --cluster-name demo
ecspresso deploy --config ecspresso.yml --no-wait
aws --endpoint-url http://127.0.0.1:15111 ecs describe-services --cluster demo --services web --query "services[0].status" --output text
aws --endpoint-url http://127.0.0.1:15111 ecs describe-services --cluster demo --services web --query "services[0].taskDefinition" --output text
aws --endpoint-url http://127.0.0.1:15111 ecs describe-services --cluster demo --services web --query "services[0].desiredCount" --output text
sed 's/nginx:1.0/nginx:2.0/' taskdef.json > taskdef.new && mv taskdef.new taskdef.json
ecspresso deploy --config ecspresso.yml --no-wait
aws --endpoint-url http://127.0.0.1:15111 ecs describe-services --cluster demo --services web --query "services[0].taskDefinition" --output text
aws --endpoint-url http://127.0.0.1:15111 ecs list-task-definitions --query "length(taskDefinitionArns)" --output text
```
#### Then
- after `aws --endpoint-url http://127.0.0.1:15111 ecs create-cluster --cluster-name demo`:
  - exit code is `0`
- after `ecspresso deploy --config ecspresso.yml --no-wait`:
  - exit code is `0`
- after `aws --endpoint-url http://127.0.0.1:15111 ecs describe-services --cluster demo --services web --query "services[0].status" --output text`:
  - exit code is `0`
  - stdout contains `ACTIVE`
- after `aws --endpoint-url http://127.0.0.1:15111 ecs describe-services --cluster demo --services web --query "services[0].taskDefinition" --output text`:
  - exit code is `0`
  - stdout contains `web:1`
- after `aws --endpoint-url http://127.0.0.1:15111 ecs describe-services --cluster demo --services web --query "services[0].desiredCount" --output text`:
  - exit code is `0`
  - stdout contains `2`
- after `sed 's/nginx:1.0/nginx:2.0/' taskdef.json > taskdef.new && mv taskdef.new taskdef.json`:
  - exit code is `0`
- after `ecspresso deploy --config ecspresso.yml --no-wait`:
  - exit code is `0`
- after `aws --endpoint-url http://127.0.0.1:15111 ecs describe-services --cluster demo --services web --query "services[0].taskDefinition" --output text`:
  - exit code is `0`
  - stdout contains `web:2`
- after `aws --endpoint-url http://127.0.0.1:15111 ecs list-task-definitions --query "length(taskDefinitionArns)" --output text`:
  - exit code is `0`
  - stdout contains `2`
## ecspresso (Amazon ECS deploy tool)
Source: `test/e2e/thirdparty/ecspresso/ecspresso.atago.yaml`
### Scenario: version prints without error
#### When
```shell
ecspresso version
```
#### Then
- exit code is `0`
### Scenario: render substitutes an env template function
#### Given
- Fixture file `ecspresso.yml` is created.
- Fixture file `taskdef.json` is created.
- Environment variables are set: IMAGE_TAG.
#### Inputs
_Fixture `ecspresso.yml`:_
```text
region: ap-northeast-1
cluster: demo
service: web
task_definition: taskdef.json
```
_Fixture `taskdef.json`:_
```text
{
  "family": "web",
  "containerDefinitions": [
    {
      "name": "web",
      "image": "nginx:{{ env `IMAGE_TAG` `latest` }}",
      "cpu": 128,
      "memory": 256
    }
  ]
}
```
#### When
```shell
ecspresso render --config ecspresso.yml taskdef
```
#### Then
- exit code is `0`
- stdout at `$.containerDefinitions[0].image` equals `nginx:1.2.3`
### Scenario: render falls back to the template default when the env is unset
#### Given
- Fixture file `ecspresso.yml` is created.
- Fixture file `taskdef.json` is created.
#### Inputs
_Fixture `ecspresso.yml`:_
```text
region: ap-northeast-1
cluster: demo
service: web
task_definition: taskdef.json
```
_Fixture `taskdef.json`:_
```text
{
  "family": "web",
  "containerDefinitions": [
    {
      "name": "web",
      "image": "nginx:{{ env `IMAGE_TAG` `latest` }}",
      "cpu": 128,
      "memory": 256
    }
  ]
}
```
#### When
```shell
ecspresso render --config ecspresso.yml taskdef
```
#### Then
- exit code is `0`
- stdout at `$.containerDefinitions[0].image` equals `nginx:latest`
### Scenario: render evaluates a jsonnet task definition with an external variable
#### Given
- Fixture file `ecspresso.yml` is created.
- Fixture file `task.jsonnet` is created.
#### Inputs
_Fixture `ecspresso.yml`:_
```text
region: ap-northeast-1
cluster: demo
service: web
task_definition: task.jsonnet
```
_Fixture `task.jsonnet`:_
```text
{
  family: "web",
  containerDefinitions: [
    {
      name: "web",
      image: "nginx:" + std.extVar("tag"),
      cpu: 64,
      memory: 128,
    },
  ],
}
```
#### When
```shell
ecspresso render --config ecspresso.yml --ext-str tag=9.9.9 taskdef
```
#### Then
- exit code is `0`
- stdout at `$.containerDefinitions[0].image` equals `nginx:9.9.9`
### Scenario: render config resolves the defaults
#### Given
- Fixture file `ecspresso.yml` is created.
- Fixture file `taskdef.json` is created.
#### Inputs
_Fixture `ecspresso.yml`:_
```text
region: ap-northeast-1
cluster: demo
service: web
task_definition: taskdef.json
```
_Fixture `taskdef.json`:_
```text
{"family": "web", "containerDefinitions": []}
```
#### When
```shell
ecspresso render --config ecspresso.yml config
```
#### Then
- exit code is `0`
- stdout contains `cluster: demo`, `timeout:`
### Scenario: an undefined must_env fails the render
#### Given
- Fixture file `ecspresso.yml` is created.
- Fixture file `taskdef.json` is created.
#### Inputs
_Fixture `ecspresso.yml`:_
```text
region: ap-northeast-1
cluster: demo
service: web
task_definition: taskdef.json
```
_Fixture `taskdef.json`:_
```text
{
  "family": "web",
  "containerDefinitions": [
    {
      "name": "web",
      "image": "nginx:{{ must_env `REQUIRED_TAG` }}",
      "cpu": 1,
      "memory": 1
    }
  ]
}
```
#### When
```shell
ecspresso render --config ecspresso.yml taskdef
```
#### Then
- exit code is `2`
- stderr contains `REQUIRED_TAG is not defined`
### Scenario: a missing config file fails cleanly
#### When
```shell
ecspresso render --config nonexistent.yml taskdef
```
#### Then
- exit code is `1`
- stderr contains `failed to load config file`