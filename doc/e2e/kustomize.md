# atago Behavior Specs
## Summary
2 suites · 18 scenarios
## Contents
- [kustomize + changes (kustomization file authoring)](#kustomize--changes-kustomization-file-authoring) — 5 scenarios
  - [create writes exactly the kustomization file](#scenario-create-writes-exactly-the-kustomization-file)
  - [edit set namespace modifies only the kustomization file](#scenario-edit-set-namespace-modifies-only-the-kustomization-file)
  - [edit set image records the image override in the file](#scenario-edit-set-image-records-the-image-override-in-the-file)
  - [edit fix migrates the deprecated commonLabels field to labels](#scenario-edit-fix-migrates-the-deprecated-commonlabels-field-to-labels)
  - [edit without a kustomization file fails](#scenario-edit-without-a-kustomization-file-fails)
- [kustomize (declarative Kubernetes config)](#kustomize-declarative-kubernetes-config) — 13 scenarios
  - [version prints a semantic version](#scenario-version-prints-a-semantic-version)
  - [build applies name prefix, namespace, labels and image tag](#scenario-build-applies-name-prefix-namespace-labels-and-image-tag)
  - [configMapGenerator appends a deterministic content hash](#scenario-configmapgenerator-appends-a-deterministic-content-hash)
  - [disableNameSuffixHash drops the generated hash suffix](#scenario-disablenamesuffixhash-drops-the-generated-hash-suffix)
  - [secretGenerator base64-encodes values and never leaks the plaintext](#scenario-secretgenerator-base64-encodes-values-and-never-leaks-the-plaintext)
  - [an overlay JSON6902-patches a base without editing it](#scenario-an-overlay-json6902-patches-a-base-without-editing-it)
  - [a strategic-merge patch overrides only the fields it names](#scenario-a-strategic-merge-patch-overrides-only-the-fields-it-names)
  - [an empty resources list renders nothing](#scenario-an-empty-resources-list-renders-nothing)
  - [resource order does not change the rendered output](#scenario-resource-order-does-not-change-the-rendered-output)
  - [rendering is deterministic across repeated builds](#scenario-rendering-is-deterministic-across-repeated-builds)
  - [building a directory without a kustomization fails cleanly](#scenario-building-a-directory-without-a-kustomization-fails-cleanly)
  - [a missing referenced resource is reported on stderr](#scenario-a-missing-referenced-resource-is-reported-on-stderr)
  - [the load restrictor refuses to read a file outside the root](#scenario-the-load-restrictor-refuses-to-read-a-file-outside-the-root)
## kustomize + changes (kustomization file authoring)
Source: `test/e2e/thirdparty/kustomize/changes.atago.yaml`
### Scenario: create writes exactly the kustomization file
_only when `kustomize version` succeeds_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
kustomize create
```
#### Then
- exit code is `0`
- the step changed exactly created `kustomization.yaml`, modified nothing, deleted nothing
- file `kustomization.yaml` contains `kind: Kustomization`
### Scenario: edit set namespace modifies only the kustomization file
_only when `kustomize version` succeeds_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
kustomize create
kustomize edit set namespace staging
```
#### Then
- after `kustomize create`:
  - exit code is `0`
- after `kustomize edit set namespace staging`:
  - exit code is `0`
  - the step changed exactly created nothing, modified `kustomization.yaml`, deleted nothing
  - file `kustomization.yaml` contains `namespace: staging`
### Scenario: edit set image records the image override in the file
_only when `kustomize version` succeeds_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
kustomize create
kustomize edit set image nginx=nginx:3.0
```
#### Then
- after `kustomize create`:
  - exit code is `0`
- after `kustomize edit set image nginx=nginx:3.0`:
  - exit code is `0`
  - the step changed exactly created nothing, modified `kustomization.yaml`, deleted nothing
  - file `kustomization.yaml` contains `name: nginx`, `newTag: "3.0"`
### Scenario: edit fix migrates the deprecated commonLabels field to labels
_only when `kustomize version` succeeds_
#### Given
- Fixture file `kustomization.yaml` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `kustomization.yaml`:_
```text
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
commonLabels:
  app: x
```
#### When
```shell
kustomize edit fix
```
#### Then
- exit code is `0`
- the step changed exactly created nothing, modified `kustomization.yaml`, deleted nothing
- file `kustomization.yaml` contains `labels:`
- file `kustomization.yaml` is checked
### Scenario: edit without a kustomization file fails
_only when `kustomize version` succeeds_
#### When
```shell
kustomize edit set namespace staging
```
#### Then
- exit code is `1`
- stderr contains `Missing kustomization file`
## kustomize (declarative Kubernetes config)
Source: `test/e2e/thirdparty/kustomize/kustomize.atago.yaml`
### Scenario: version prints a semantic version
_only when `kustomize version` succeeds_
#### When
```shell
kustomize version
```
#### Then
- exit code is `0`
- stdout matches `/^v[0-9]+\.[0-9]+\.[0-9]+/`
### Scenario: build applies name prefix, namespace, labels and image tag
_only when `kustomize version` succeeds_
#### Given
- Fixture file `base/deploy.yaml` is created.
- Fixture file `base/kustomization.yaml` is created.
#### Inputs
_Fixture `base/deploy.yaml`:_
```text
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: web
        image: nginx:1.0
```
_Fixture `base/kustomization.yaml`:_
```text
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: prod-
namespace: production
labels:
- pairs:
    app: myapp
  includeSelectors: true
images:
- name: nginx
  newTag: "2.0"
resources:
- deploy.yaml
```
#### When
```shell
kustomize build base
```
#### Then
- exit code is `0`
- stdout contains `name: prod-web`, `namespace: production`, `app: myapp`, `image: nginx:2.0`
- stdout matches `/(?s)selector:\s+matchLabels:\s+app: myapp/`
### Scenario: configMapGenerator appends a deterministic content hash
_only when `kustomize version` succeeds_
#### Given
- Fixture file `kustomization.yaml` is created.
#### Inputs
_Fixture `kustomization.yaml`:_
```text
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: settings
  literals:
  - LOG_LEVEL=info
```
#### When
```shell
kustomize build .
```
#### Then
- exit code is `0`
- stdout contains `kind: ConfigMap`, `LOG_LEVEL: info`
- stdout matches `/name: settings-[a-z0-9]+/`
### Scenario: disableNameSuffixHash drops the generated hash suffix
_only when `kustomize version` succeeds_
#### Given
- Fixture file `kustomization.yaml` is created.
#### Inputs
_Fixture `kustomization.yaml`:_
```text
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
generatorOptions:
  disableNameSuffixHash: true
configMapGenerator:
- name: settings
  literals:
  - LOG_LEVEL=info
```
#### When
```shell
kustomize build .
```
#### Then
- exit code is `0`
- stdout contains `name: settings`
- stdout does not match `/name: settings-[a-z0-9]+/`
### Scenario: secretGenerator base64-encodes values and never leaks the plaintext
_only when `kustomize version` succeeds_
#### Given
- Fixture file `kustomization.yaml` is created.
#### Inputs
_Fixture `kustomization.yaml`:_
```text
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- name: creds
  literals:
  - greeting=konnichiwa
```
#### When
```shell
kustomize build .
```
#### Then
- exit code is `0`
- stdout contains `kind: Secret`, `type: Opaque`
- stdout contains `greeting: a29ubmljaGl3YQ==`
- stdout does not contain `konnichiwa`
### Scenario: an overlay JSON6902-patches a base without editing it
_only when `kustomize version` succeeds_
#### Given
- Fixture file `base/deploy.yaml` is created.
- Fixture file `base/kustomization.yaml` is created.
- Fixture file `overlay/kustomization.yaml` is created.
#### Inputs
_Fixture `base/deploy.yaml`:_
```text
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: web
        image: nginx:1.0
```
_Fixture `base/kustomization.yaml`:_
```text
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- deploy.yaml
```
_Fixture `overlay/kustomization.yaml`:_
```text
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- ../base
patches:
- target:
    kind: Deployment
    name: web
  patch: |-
    - op: replace
      path: /spec/replicas
      value: 5
```
#### When
```shell
kustomize build overlay
kustomize build base
```
#### Then
- after `kustomize build overlay`:
  - exit code is `0`
  - stdout contains `replicas: 5`
- after `kustomize build base`:
  - exit code is `0`
  - stdout contains `replicas: 1`
### Scenario: a strategic-merge patch overrides only the fields it names
_only when `kustomize version` succeeds_
#### Given
- Fixture file `deploy.yaml` is created.
- Fixture file `replicas-patch.yaml` is created.
- Fixture file `kustomization.yaml` is created.
#### Inputs
_Fixture `deploy.yaml`:_
```text
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: web
        image: nginx:1.0
```
_Fixture `replicas-patch.yaml`:_
```text
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
spec:
  replicas: 4
```
_Fixture `kustomization.yaml`:_
```text
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- deploy.yaml
patches:
- path: replicas-patch.yaml
```
#### When
```shell
kustomize build .
```
#### Then
- exit code is `0`
- stdout contains `replicas: 4`, `image: nginx:1.0`
### Scenario: an empty resources list renders nothing
_only when `kustomize version` succeeds_
#### Given
- Fixture file `kustomization.yaml` is created.
#### Inputs
_Fixture `kustomization.yaml`:_
```text
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources: []
```
#### When
```shell
kustomize build .
```
#### Then
- exit code is `0`
- stdout is empty
### Scenario: resource order does not change the rendered output
_only when `kustomize version` succeeds_
#### Given
- Fixture file `a/cm-a.yaml` is created.
- Fixture file `a/cm-b.yaml` is created.
- Fixture file `a/kustomization.yaml` is created.
- Fixture file `b/cm-a.yaml` is created.
- Fixture file `b/cm-b.yaml` is created.
- Fixture file `b/kustomization.yaml` is created.
#### Inputs
_Fixture `a/cm-a.yaml`:_
```text
apiVersion: v1
kind: ConfigMap
metadata:
  name: a
data:
  k: v
```
_Fixture `a/cm-b.yaml`:_
```text
apiVersion: v1
kind: ConfigMap
metadata:
  name: b
data:
  k: v
```
_Fixture `a/kustomization.yaml`:_
```text
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- cm-a.yaml
- cm-b.yaml
```
_Fixture `b/cm-a.yaml`:_
```text
apiVersion: v1
kind: ConfigMap
metadata:
  name: a
data:
  k: v
```
_Fixture `b/cm-b.yaml`:_
```text
apiVersion: v1
kind: ConfigMap
metadata:
  name: b
data:
  k: v
```
_Fixture `b/kustomization.yaml`:_
```text
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- cm-b.yaml
- cm-a.yaml
```
#### When
```shell
kustomize build a -o out-a.yaml
kustomize build b -o out-b.yaml
cmp out-a.yaml out-b.yaml
```
#### Then
- after `kustomize build a -o out-a.yaml`:
  - exit code is `0`
- after `kustomize build b -o out-b.yaml`:
  - exit code is `0`
- after `cmp out-a.yaml out-b.yaml`:
  - exit code is `0`
### Scenario: rendering is deterministic across repeated builds
_only when `kustomize version` succeeds_
#### Given
- Fixture file `kustomization.yaml` is created.
#### Inputs
_Fixture `kustomization.yaml`:_
```text
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: a-
configMapGenerator:
- name: c
  literals:
  - K=V
```
#### When
```shell
kustomize build . -o first.yaml
kustomize build . -o second.yaml
cmp first.yaml second.yaml
```
#### Then
- after `kustomize build . -o first.yaml`:
  - exit code is `0`
- after `kustomize build . -o second.yaml`:
  - exit code is `0`
- after `cmp first.yaml second.yaml`:
  - exit code is `0`
### Scenario: building a directory without a kustomization fails cleanly
_only when `kustomize version` succeeds_
#### Given
- Fixture file `empty/.keep` is created.
#### When
```shell
kustomize build empty
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `unable to find one of 'kustomization.yaml'`
### Scenario: a missing referenced resource is reported on stderr
_only when `kustomize version` succeeds_
#### Given
- Fixture file `kustomization.yaml` is created.
#### Inputs
_Fixture `kustomization.yaml`:_
```text
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- missing.yaml
```
#### When
```shell
kustomize build .
```
#### Then
- exit code is `1`
- stderr contains `accumulating resources`
### Scenario: the load restrictor refuses to read a file outside the root
_only when `kustomize version` succeeds_
#### Given
- Fixture file `outside.yaml` is created.
- Fixture file `root/kustomization.yaml` is created.
#### Inputs
_Fixture `outside.yaml`:_
```text
apiVersion: v1
kind: ConfigMap
metadata:
  name: outside
data:
  k: v
```
_Fixture `root/kustomization.yaml`:_
```text
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- ../outside.yaml
```
#### When
```shell
kustomize build root
```
#### Then
- exit code is `1`
- stderr contains `security`
