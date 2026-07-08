# atago Behavior Specs
## Summary
2 suites · 7 scenarios
## Contents
- [ffmpeg + changes (single-artifact encode)](#ffmpeg--changes-single-artifact-encode) — 1 scenario
  - [lavfi source encodes exactly one output file](#scenario-lavfi-source-encodes-exactly-one-output-file)
- [ffmpeg / ffprobe (media pipeline)](#ffmpeg--ffprobe-media-pipeline) — 6 scenarios
  - [lavfi synthesizes a video file](#scenario-lavfi-synthesizes-a-video-file)
  - [ffprobe exposes the stream JSON contract](#scenario-ffprobe-exposes-the-stream-json-contract)
  - [extracted frames are the right image and deterministic](#scenario-extracted-frames-are-the-right-image-and-deterministic)
  - [transcode to webm and re-probe the codec](#scenario-transcode-to-webm-and-re-probe-the-codec)
  - [a missing input file fails with a not-found error](#scenario-a-missing-input-file-fails-with-a-not-found-error)
  - [ffprobe on non-media data reports invalid input](#scenario-ffprobe-on-non-media-data-reports-invalid-input)
## ffmpeg + changes (single-artifact encode)
Source: `test/e2e/thirdparty/ffmpeg/changes.atago.yaml`
### Scenario: lavfi source encodes exactly one output file
_only when `ffmpeg -version` succeeds_
#### When
```shell
ffmpeg -v error -f lavfi -i testsrc=duration=0.1:size=64x64 -y out.mp4
```
#### Then
- exit code is `0`
- the step changed exactly created `out.mp4`, modified nothing, deleted nothing
## ffmpeg / ffprobe (media pipeline)
Source: `test/e2e/thirdparty/ffmpeg/ffmpeg.atago.yaml`
### Scenario: lavfi synthesizes a video file
_only when `ffmpeg -version` succeeds_
#### When
```shell
ffmpeg -v error -f lavfi -i testsrc=duration=1:size=320x240:rate=10 out.mp4
```
#### Then
- exit code is `0`
- file `out.mp4` exists
#### Generated artifacts
- `out.mp4`
### Scenario: ffprobe exposes the stream JSON contract
_only when `ffmpeg -version` succeeds_
#### When
```shell
ffmpeg -v error -f lavfi -i testsrc=duration=1:size=320x240:rate=10 out.mp4
ffprobe -v error -print_format json -show_streams out.mp4
```
#### Then
- after `ffmpeg -v error -f lavfi -i testsrc=duration=1:size=320x240:rate=10 out.mp4`:
  - exit code is `0`
- after `ffprobe -v error -print_format json -show_streams out.mp4`:
  - exit code is `0`
  - stdout at `$.streams[0].width` equals `320`
  - stdout at `$.streams[0].height` equals `240`
  - stdout at `$.streams[0].codec_type` equals `video`
  - stdout at `$.streams` has length 1
### Scenario: extracted frames are the right image and deterministic
_only when `ffmpeg -version` succeeds_
#### When
```shell
ffmpeg -v error -f lavfi -i testsrc=duration=1:size=320x240:rate=10 out.mp4
ffmpeg -v error -i out.mp4 -frames:v 1 frame.png
ffmpeg -v error -i out.mp4 -frames:v 1 frame2.png
```
#### Then
- after `ffmpeg -v error -f lavfi -i testsrc=duration=1:size=320x240:rate=10 out.mp4`:
  - exit code is `0`
- after `ffmpeg -v error -i out.mp4 -frames:v 1 frame.png`:
  - exit code is `0`
  - image `frame.png` is `png`, width 320
  - image `frame.png` height 240
- after `ffmpeg -v error -i out.mp4 -frames:v 1 frame2.png`:
  - exit code is `0`
  - image `frame2.png` similar to `${workdir}/frame.png`
#### Generated artifacts
- `frame.png`
- `frame2.png`
### Scenario: transcode to webm and re-probe the codec
_only when `ffmpeg -version` succeeds_
#### When
```shell
ffmpeg -v error -f lavfi -i testsrc=duration=1:size=320x240:rate=10 out.mp4
ffmpeg -v error -i out.mp4 -c:v libvpx-vp9 -b:v 100k out.webm
ffprobe -v error -print_format json -show_streams out.webm
```
#### Then
- after `ffmpeg -v error -f lavfi -i testsrc=duration=1:size=320x240:rate=10 out.mp4`:
  - exit code is `0`
- after `ffmpeg -v error -i out.mp4 -c:v libvpx-vp9 -b:v 100k out.webm`:
  - exit code is `0`
  - file `out.webm` exists
- after `ffprobe -v error -print_format json -show_streams out.webm`:
  - exit code is `0`
  - stdout at `$.streams[0].codec_name` equals `vp9`
#### Generated artifacts
- `out.webm`
### Scenario: a missing input file fails with a not-found error
_only when `ffmpeg -version` succeeds_
#### When
```shell
ffmpeg -v error -i no_such_input.mp4 out.mp4
```
#### Then
- exit code is one of `1`, `254`
- stderr contains `No such file or directory`
### Scenario: ffprobe on non-media data reports invalid input
_only when `ffmpeg -version` succeeds_
#### Given
- Fixture file `corrupt.mp4` is created.
#### When
```shell
ffprobe -v error corrupt.mp4
```
#### Then
- exit code is one of `1`, `254`
- stderr contains `Invalid data`
