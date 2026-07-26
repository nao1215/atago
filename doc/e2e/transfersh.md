# atago Behavior Specs
## Summary
1 suite · 3 scenarios
## Contents
- [transfer.sh (self-hosted file sharing)](#transfersh-self-hosted-file-sharing) — 3 scenarios
  - [the binary reports its version](#scenario-the-binary-reports-its-version)
  - [a binary upload round-trips byte-for-byte](#scenario-a-binary-upload-round-trips-byte-for-byte)
  - [a browser-style multipart upload is accepted too](#scenario-a-browser-style-multipart-upload-is-accepted-too)
## transfer.sh (self-hosted file sharing)
[transfer.sh](https://github.com/dutchcoders/transfer.sh) takes a file and
gives you a URL to fetch it back from. The only guarantee that matters is
that what comes back is what went in.

So this suite tests it with binary data rather than text. A PNG is uploaded
as a raw request body, the share URL is captured from the response, the file
is downloaded back to disk, and the bytes are compared to the original —
both as a valid image and byte for byte. Text survives a lot of accidental
mangling that binary does not.

The browser-style multipart upload path is exercised as well, since that is
how a file arrives from a web form rather than from `curl`.

Source: `test/e2e/thirdparty/transfersh/transfersh.atago.yaml`
### Scenario: the binary reports its version
#### When
```shell
transfer.sh version
```
#### Then
- exit code is `0`
- stdout contains `transfer.sh`
### Scenario: a binary upload round-trips byte-for-byte
#### Given
- Background service `transfersh` is started: `transfer.sh --provider local --basedir storage --temp-path tmp --listener 127.0.0.1:18210`.
- Fixture file `storage/.keep` is created.
- Fixture file `tmp/.keep` is created.
- Fixture file `pixel.png` is created.
#### When
```shell
# HTTP PUT /pixel.png
# capture ${share_url} from the response body
# HTTP GET ${share_url}
cmp pixel.png downloaded.png
```
#### Then
- after `HTTP PUT /pixel.png`:
  - HTTP status is `200`
  - body contains `http://127.0.0.1:18210/`
- after `HTTP GET ${share_url}`:
  - HTTP status is `200`
  - image `downloaded.png` is `png`, width 1, height 1
- after `cmp pixel.png downloaded.png`:
  - exit code is `0`
#### Generated artifacts
- `downloaded.png`
### Scenario: a browser-style multipart upload is accepted too
#### Given
- Background service `transfersh` is started: `transfer.sh --provider local --basedir storage --temp-path tmp --listener 127.0.0.1:18211`.
- Fixture file `storage/.keep` is created.
- Fixture file `tmp/.keep` is created.
- Fixture file `notes.txt` is created.
#### Inputs
_Fixture `notes.txt`:_
```text
shared through a multipart form
```
#### When
```shell
# HTTP POST /
# capture ${share_url} from the response body
# HTTP GET ${share_url}
# HTTP GET /no-such-token/notes.txt
```
#### Then
- after `HTTP POST /`:
  - HTTP status is `200`
  - body contains `http://127.0.0.1:18211/`
- after `HTTP GET ${share_url}`:
  - HTTP status is `200`
  - body contains `shared through a multipart form`
- after `HTTP GET /no-such-token/notes.txt`:
  - HTTP status is `404`
