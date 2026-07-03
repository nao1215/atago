# atago Behavior Specs
## Summary
1 suite · 6 scenarios
## Contents
- [redis (server + client, signal-step testbed)](#redis-server--client-signal-step-testbed) — 6 scenarios
  - [server boots (log readiness) and answers PING](#scenario-server-boots-log-readiness-and-answers-ping)
  - [server boots (port readiness) and round-trips SET/GET/INCR](#scenario-server-boots-port-readiness-and-round-trips-setgetincr)
  - [EXPIRE and TTL report lifetime state](#scenario-expire-and-ttl-report-lifetime-state)
  - [connecting to a closed port fails loudly](#scenario-connecting-to-a-closed-port-fails-loudly)
  - [an unknown command reports ERR without killing the server](#scenario-an-unknown-command-reports-err-without-killing-the-server)
  - [SHUTDOWN NOSAVE stops the server and PING starts failing](#scenario-shutdown-nosave-stops-the-server-and-ping-starts-failing)
## redis (server + client, signal-step testbed)
Source: `test/e2e/thirdparty/redis/redis.atago.yaml`
### Scenario: server boots (log readiness) and answers PING
_skipped on windows_
#### Given
- Background service `redis` is started: `redis-server --port 16379 --save '' --appendonly no`.
#### When
```shell
redis-cli -p 16379 PING
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: server boots (port readiness) and round-trips SET/GET/INCR
_skipped on windows_
#### Given
- Background service `redis` is started: `redis-server --port 16380 --save '' --appendonly no`.
#### When
```shell
redis-cli -p 16380 SET greeting hello
redis-cli -p 16380 GET greeting
redis-cli -p 16380 INCR counter
redis-cli -p 16380 INCR counter
```
#### Then
- after `redis-cli -p 16380 SET greeting hello`:
  - exit code is `0`
  - stdout equals an exact value
- after `redis-cli -p 16380 GET greeting`:
  - exit code is `0`
  - stdout equals an exact value
- after `redis-cli -p 16380 INCR counter`:
  - exit code is `0`
  - stdout equals an exact value
- after `redis-cli -p 16380 INCR counter`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: EXPIRE and TTL report lifetime state
_skipped on windows_
#### Given
- Background service `redis` is started: `redis-server --port 16381 --save '' --appendonly no`.
#### When
```shell
redis-cli -p 16381 SET k v
redis-cli -p 16381 TTL k
redis-cli -p 16381 EXPIRE k 100
redis-cli -p 16381 TTL k
redis-cli -p 16381 TTL no-such-key
```
#### Then
- after `redis-cli -p 16381 SET k v`:
  - exit code is `0`
- after `redis-cli -p 16381 TTL k`:
  - exit code is `0`
  - stdout equals an exact value
- after `redis-cli -p 16381 EXPIRE k 100`:
  - exit code is `0`
  - stdout equals an exact value
- after `redis-cli -p 16381 TTL k`:
  - exit code is `0`
  - stdout matches `/^(100|9[0-9])\b/`
- after `redis-cli -p 16381 TTL no-such-key`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: connecting to a closed port fails loudly
_skipped on windows_
#### When
```shell
redis-cli -p 16999 PING
```
#### Then
- exit code is `1`
- stderr contains `Connection refused`
### Scenario: an unknown command reports ERR without killing the server
_skipped on windows_
#### Given
- Background service `redis` is started: `redis-server --port 16382 --save '' --appendonly no`.
#### When
```shell
redis-cli -p 16382 NOSUCHCOMMAND arg; echo "rc=$?"
redis-cli -p 16382 PING
```
#### Then
- after `redis-cli -p 16382 NOSUCHCOMMAND arg; echo "rc=$?"`:
  - stdout contains `ERR unknown command`
- after `redis-cli -p 16382 PING`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: SHUTDOWN NOSAVE stops the server and PING starts failing
_skipped on windows_
#### Given
- Background service `redis` is started: `redis-server --port 16383 --save '' --appendonly no`.
#### When
```shell
redis-cli -p 16383 PING
redis-cli -p 16383 SHUTDOWN NOSAVE
redis-cli -p 16383 PING
```
#### Then
- after `redis-cli -p 16383 PING`:
  - exit code is `0`
  - stdout equals an exact value
- after `redis-cli -p 16383 PING`:
  - exit code is `1`
  - stderr contains `Connection refused`