# atago Behavior Specs
## Summary
1 suite · 5 scenarios
## Contents
- [coredns (self-hosted DNS server)](#coredns-self-hosted-dns-server) — 5 scenarios
  - [the binary reports its version](#scenario-the-binary-reports-its-version)
  - [an authored zone is served authoritatively](#scenario-an-authored-zone-is-served-authoritatively)
  - [missing names and foreign zones get the right RCODEs](#scenario-missing-names-and-foreign-zones-get-the-right-rcodes)
  - [the health plugin answers over HTTP while DNS serves](#scenario-the-health-plugin-answers-over-http-while-dns-serves)
  - [a broken Corefile is rejected at startup](#scenario-a-broken-corefile-is-rejected-at-startup)
## coredns (self-hosted DNS server)
Source: `test/e2e/thirdparty/coredns/coredns.atago.yaml`
### Scenario: the binary reports its version
#### When
```shell
coredns -version
```
#### Then
- exit code is `0`
- stdout matches `/CoreDNS-[0-9]+\.[0-9]+\.[0-9]+/`
### Scenario: an authored zone is served authoritatively
#### Given
- Background service `coredns` is started: `coredns -conf Corefile`.
- Fixture file `Corefile` is created.
- Fixture file `zones/example.test.zone` is created.
#### Inputs
_Fixture `Corefile`:_
```
example.test:18150 {
    bind 127.0.0.1
    file zones/example.test.zone
}
```
_Fixture `zones/example.test.zone`:_
```
$ORIGIN example.test.
@   3600 IN SOA ns.example.test. admin.example.test. (2026070201 7200 3600 1209600 3600)
@   3600 IN NS  ns.example.test.
ns  3600 IN A   127.0.0.1
www 3600 IN A   192.0.2.10
txt 3600 IN TXT "issued-by=atago"
alias 3600 IN CNAME www.example.test.
```
#### When
```shell
dig @127.0.0.1 -p 18150 www.example.test A +noall +comments +answer
dig @127.0.0.1 -p 18150 txt.example.test TXT +short
dig @127.0.0.1 -p 18150 alias.example.test A +short
```
#### Then
- after `dig @127.0.0.1 -p 18150 www.example.test A +noall +comments +answer`:
  - exit code is `0`
  - stdout contains `status: NOERROR`
  - stdout matches `/flags:[^;]*\baa\b/`
  - stdout contains `192.0.2.10`
- after `dig @127.0.0.1 -p 18150 txt.example.test TXT +short`:
  - exit code is `0`
  - stdout contains `issued-by=atago`
- after `dig @127.0.0.1 -p 18150 alias.example.test A +short`:
  - exit code is `0`
  - stdout contains `www.example.test.`, `192.0.2.10`
### Scenario: missing names and foreign zones get the right RCODEs
#### Given
- Background service `coredns` is started: `coredns -conf Corefile`.
- Fixture file `Corefile` is created.
- Fixture file `zones/example.test.zone` is created.
#### Inputs
_Fixture `Corefile`:_
```
example.test:18151 {
    bind 127.0.0.1
    file zones/example.test.zone
}
```
_Fixture `zones/example.test.zone`:_
```
$ORIGIN example.test.
@   3600 IN SOA ns.example.test. admin.example.test. (2026070201 7200 3600 1209600 3600)
@   3600 IN NS  ns.example.test.
ns  3600 IN A   127.0.0.1
www 3600 IN A   192.0.2.10
```
#### When
```shell
dig @127.0.0.1 -p 18151 no-such-name.example.test A +noall +comments
dig @127.0.0.1 -p 18151 www.example.com A +noall +comments
```
#### Then
- after `dig @127.0.0.1 -p 18151 no-such-name.example.test A +noall +comments`:
  - exit code is `0`
  - stdout contains `status: NXDOMAIN`
- after `dig @127.0.0.1 -p 18151 www.example.com A +noall +comments`:
  - exit code is `0`
  - stdout contains `status: REFUSED`
### Scenario: the health plugin answers over HTTP while DNS serves
#### Given
- Background service `coredns` is started: `coredns -conf Corefile`.
- Fixture file `Corefile` is created.
- Fixture file `zones/example.test.zone` is created.
#### Inputs
_Fixture `Corefile`:_
```
example.test:18152 {
    bind 127.0.0.1
    file zones/example.test.zone
    health 127.0.0.1:18153
}
```
_Fixture `zones/example.test.zone`:_
```
$ORIGIN example.test.
@   3600 IN SOA ns.example.test. admin.example.test. (2026070201 7200 3600 1209600 3600)
@   3600 IN NS  ns.example.test.
ns  3600 IN A   127.0.0.1
www 3600 IN A   192.0.2.10
```
#### When
```shell
# HTTP GET /health
dig @127.0.0.1 -p 18152 www.example.test A +short
```
#### Then
- after `HTTP GET /health`:
  - HTTP status is `200`
  - body contains `OK`
- after `dig @127.0.0.1 -p 18152 www.example.test A +short`:
  - exit code is `0`
  - stdout contains `192.0.2.10`
### Scenario: a broken Corefile is rejected at startup
#### Given
- Fixture file `Corefile` is created.
#### Inputs
_Fixture `Corefile`:_
```
example.test:18159 {
    no_such_plugin_xyz
}
```
#### When
```shell
coredns -conf Corefile
```
#### Then
- exit code is not `0`
- stderr contains `Unknown directive 'no_such_plugin_xyz'`