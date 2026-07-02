# atago Behavior Specs
## Summary
540 suites · 1522 scenarios
## Contents
- [mimixbox ar](#mimixbox-ar) — 2 scenarios
  - [lists members](#scenario-lists-members)
  - [extracts a member](#scenario-extracts-a-member)
- [mimixbox bunzip2](#mimixbox-bunzip2) — 1 scenario
  - [decompresses a .bz2 file to stdout with -c](#scenario-decompresses-a-bz2-file-to-stdout-with--c)
- [mimixbox bzcat](#mimixbox-bzcat) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help)
- [mimixbox bzip2](#mimixbox-bzip2) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-1)
- [mimixbox compress and uncompress](#mimixbox-compress-and-uncompress) — 1 scenario
  - [round-trips a file through compress and uncompress](#scenario-round-trips-a-file-through-compress-and-uncompress)
- [mimixbox cpio](#mimixbox-cpio) — 2 scenarios
  - [round-trips a file through -o and -i](#scenario-round-trips-a-file-through--o-and--i)
  - [lists archive contents with -i -t](#scenario-lists-archive-contents-with--i--t)
- [mimixbox bzip2, lzop and Debian package applets](#mimixbox-bzip2-lzop-and-debian-package-applets) — 7 scenarios
  - [bzip2 | bzip2 -dc round-trips data](#scenario-bzip2--bzip2--dc-round-trips-data)
  - [lzop | lzopcat round-trips data](#scenario-lzop--lzopcat-round-trips-data)
  - [lzop | unlzop -c round-trips data](#scenario-lzop--unlzop--c-round-trips-data)
  - [dpkg-deb -c lists package contents](#scenario-dpkg-deb--c-lists-package-contents)
  - [dpkg-deb -f prints a control field](#scenario-dpkg-deb--f-prints-a-control-field)
  - [dpkg -x extracts the data tarball path-safely](#scenario-dpkg--x-extracts-the-data-tarball-path-safely)
  - [dpkg rejects unsupported database operations](#scenario-dpkg-rejects-unsupported-database-operations)
- [mimixbox dpkg-deb](#mimixbox-dpkg-deb) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-2)
- [mimixbox dpkg](#mimixbox-dpkg) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-3)
- [mimixbox gunzip](#mimixbox-gunzip) — 1 scenario
  - [decompresses a .gz file to stdout with -c](#scenario-decompresses-a-gz-file-to-stdout-with--c)
- [mimixbox archival commands expose a dedicated --help helper](#mimixbox-archival-commands-expose-a-dedicated---help-helper) — 17 scenarios
  - [bzcat --help is structured](#scenario-bzcat---help-is-structured)
  - [bzip2 --help is structured](#scenario-bzip2---help-is-structured)
  - [dpkg --help is structured](#scenario-dpkg---help-is-structured)
  - [dpkg-deb --help is structured](#scenario-dpkg-deb---help-is-structured)
  - [lzcat --help is structured](#scenario-lzcat---help-is-structured)
  - [lzma --help is structured](#scenario-lzma---help-is-structured)
  - [lzopcat --help is structured](#scenario-lzopcat---help-is-structured)
  - [pipe_progress --help is structured](#scenario-pipe_progress---help-is-structured)
  - [rpm2cpio --help is structured](#scenario-rpm2cpio---help-is-structured)
  - [uncompress --help is structured](#scenario-uncompress---help-is-structured)
  - [unlzma --help is structured](#scenario-unlzma---help-is-structured)
  - [unlzop --help is structured](#scenario-unlzop---help-is-structured)
  - [unxz --help is structured](#scenario-unxz---help-is-structured)
  - [unzip --help is structured](#scenario-unzip---help-is-structured)
  - [xz --help is structured](#scenario-xz---help-is-structured)
  - [xzcat --help is structured](#scenario-xzcat---help-is-structured)
  - [zcat --help is structured](#scenario-zcat---help-is-structured)
- [mimixbox rpm and rpm2cpio](#mimixbox-rpm-and-rpm2cpio) — 3 scenarios
  - [queries the package identity with rpm -qp](#scenario-queries-the-package-identity-with-rpm--qp)
  - [lists package files with rpm -qpl](#scenario-lists-package-files-with-rpm--qpl)
  - [extracts the payload with rpm2cpio](#scenario-extracts-the-payload-with-rpm2cpio)
- [mimixbox rpm2cpio](#mimixbox-rpm2cpio) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-4)
- [mimixbox tar](#mimixbox-tar) — 2 scenarios
  - [creates and extracts an archive](#scenario-creates-and-extracts-an-archive)
  - [lists archive contents](#scenario-lists-archive-contents)
- [mimixbox uncompress](#mimixbox-uncompress) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-5)
- [mimixbox unlzma](#mimixbox-unlzma) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-6)
- [mimixbox unlzop](#mimixbox-unlzop) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-7)
- [mimixbox unxz](#mimixbox-unxz) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-8)
- [mimixbox unzip](#mimixbox-unzip) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-9)
- [mimixbox xz](#mimixbox-xz) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-10)
- [mimixbox xzcat](#mimixbox-xzcat) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-11)
- [mimixbox compression applets](#mimixbox-compression-applets) — 4 scenarios
  - [xz | xzcat round-trips data](#scenario-xz--xzcat-round-trips-data)
  - [lzma | unlzma round-trips data](#scenario-lzma--unlzma-round-trips-data)
  - [zcat decompresses a gzip file to stdout](#scenario-zcat-decompresses-a-gzip-file-to-stdout)
  - [pipe_progress passes stdin through to stdout](#scenario-pipe_progress-passes-stdin-through-to-stdout)
- [mimixbox zcat](#mimixbox-zcat) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-12)
- [mimixbox zip and unzip](#mimixbox-zip-and-unzip) — 2 scenarios
  - [lists a zipped file via unzip -l](#scenario-lists-a-zipped-file-via-unzip--l)
  - [round-trips a file through zip and unzip](#scenario-round-trips-a-file-through-zip-and-unzip)
- [mimixbox \[](#mimixbox-) — 2 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-13)
  - [documents its purpose in --help](#scenario-documents-its-purpose-in---help)
- [mimixbox \[\[](#mimixbox--1) — 2 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-14)
  - [documents its purpose in --help](#scenario-documents-its-purpose-in---help-1)
- [mimixbox ash](#mimixbox-ash) — 2 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-15)
  - [documents its purpose in --help](#scenario-documents-its-purpose-in---help-2)
- [mimixbox bash](#mimixbox-bash) — 2 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-16)
  - [documents its purpose in --help](#scenario-documents-its-purpose-in---help-3)
- [mimixbox busybox](#mimixbox-busybox) — 2 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-17)
  - [documents its purpose in --help](#scenario-documents-its-purpose-in---help-4)
- [mimixbox compat front-ends](#mimixbox-compat-front-ends) — 6 scenarios
  - [the \[ alias returns true for an existing file](#scenario-the--alias-returns-true-for-an-existing-file)
  - [the \[ alias returns false for a missing file](#scenario-the--alias-returns-false-for-a-missing-file)
  - [busybox dispatches to an applet](#scenario-busybox-dispatches-to-an-applet)
  - [busybox --list shows applets](#scenario-busybox---list-shows-applets)
  - [sh -c runs a command without a prompt](#scenario-sh--c-runs-a-command-without-a-prompt)
  - [bash reads a non-interactive script from stdin without a prompt](#scenario-bash-reads-a-non-interactive-script-from-stdin-without-a-prompt)
- [mimixbox cttyhack](#mimixbox-cttyhack) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-18)
- [mimixbox compat commands expose a dedicated --help helper](#mimixbox-compat-commands-expose-a-dedicated---help-helper) — 8 scenarios
  - [\[ --help is structured](#scenario----help-is-structured)
  - [\[\[ --help is structured](#scenario----help-is-structured-1)
  - [ash --help is structured](#scenario-ash---help-is-structured)
  - [bash --help is structured](#scenario-bash---help-is-structured)
  - [busybox --help is structured](#scenario-busybox---help-is-structured)
  - [cttyhack --help is structured](#scenario-cttyhack---help-is-structured)
  - [hush --help is structured](#scenario-hush---help-is-structured)
  - [unit --help is structured](#scenario-unit---help-is-structured)
- [mimixbox hush](#mimixbox-hush) — 2 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-19)
  - [documents its purpose in --help](#scenario-documents-its-purpose-in---help-5)
- [mimixbox sh](#mimixbox-sh) — 2 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-20)
  - [documents its purpose in --help](#scenario-documents-its-purpose-in---help-6)
- [mimixbox unit](#mimixbox-unit) — 2 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-21)
  - [documents its purpose in --help](#scenario-documents-its-purpose-in---help-7)
- [mimixbox adjtimex](#mimixbox-adjtimex) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-22)
- [mimixbox ascii](#mimixbox-ascii) — 2 scenarios
  - [prints 128 entries](#scenario-prints-128-entries)
  - [maps code 65 to A](#scenario-maps-code-65-to-a)
- [mimixbox bbconfig](#mimixbox-bbconfig) — 3 scenarios
  - [prints the version line](#scenario-prints-the-version-line)
  - [lists itself among the applets](#scenario-lists-itself-among-the-applets)
  - [rejects an unexpected argument](#scenario-rejects-an-unexpected-argument)
- [mimixbox beep](#mimixbox-beep) — 2 scenarios
  - [rejects a non-positive frequency](#scenario-rejects-a-non-positive-frequency)
  - [rejects a zero repeat count](#scenario-rejects-a-zero-repeat-count)
- [mimixbox chat](#mimixbox-chat) — 3 scenarios
  - [sends the reply after the expected string](#scenario-sends-the-reply-after-the-expected-string)
  - [requires a script](#scenario-requires-a-script)
  - [fails when an expected string never arrives](#scenario-fails-when-an-expected-string-never-arrives)
- [mimixbox chvt](#mimixbox-chvt) — 2 scenarios
  - [rejects a non-numeric VT](#scenario-rejects-a-non-numeric-vt)
  - [requires a VT number](#scenario-requires-a-vt-number)
- [mimixbox clear](#mimixbox-clear) — 2 scenarios
  - [prints usage with --help and exits 0](#scenario-prints-usage-with---help-and-exits-0)
  - [exits 0 when clearing the screen](#scenario-exits-0-when-clearing-the-screen)
- [mimixbox conspy](#mimixbox-conspy) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-23)
- [mimixbox deallocvt](#mimixbox-deallocvt) — 2 scenarios
  - [rejects a non-numeric VT](#scenario-rejects-a-non-numeric-vt-1)
  - [describes itself with --help](#scenario-describes-itself-with---help-24)
- [mimixbox dumpkmap](#mimixbox-dumpkmap) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-25)
- [mimixbox fgconsole](#mimixbox-fgconsole) — 2 scenarios
  - [fails without a virtual console](#scenario-fails-without-a-virtual-console)
  - [describes itself with --help](#scenario-describes-itself-with---help-26)
- [mimixbox console-tools --help contract](#mimixbox-console-tools---help-contract) — 11 scenarios
  - [adjtimex --help is structured](#scenario-adjtimex---help-is-structured)
  - [conspy --help is structured](#scenario-conspy---help-is-structured)
  - [dumpkmap --help is structured](#scenario-dumpkmap---help-is-structured)
  - [less --help is structured](#scenario-less---help-is-structured)
  - [loadfont --help is structured](#scenario-loadfont---help-is-structured)
  - [loadkmap --help is structured](#scenario-loadkmap---help-is-structured)
  - [microcom --help is structured](#scenario-microcom---help-is-structured)
  - [more --help is structured](#scenario-more---help-is-structured)
  - [openvt --help is structured](#scenario-openvt---help-is-structured)
  - [rx --help is structured](#scenario-rx---help-is-structured)
  - [setfont --help is structured](#scenario-setfont---help-is-structured)
- [mimixbox inotifyd](#mimixbox-inotifyd) — 2 scenarios
  - [runs the handler on a create event](#scenario-runs-the-handler-on-a-create-event)
  - [requires a handler and a file](#scenario-requires-a-handler-and-a-file)
- [mimixbox kbd_mode](#mimixbox-kbd_mode) — 2 scenarios
  - [rejects conflicting mode options](#scenario-rejects-conflicting-mode-options)
  - [describes itself with --help](#scenario-describes-itself-with---help-27)
- [mimixbox loadfont](#mimixbox-loadfont) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-28)
- [mimixbox loadkmap](#mimixbox-loadkmap) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-29)
- [mimixbox microcom](#mimixbox-microcom) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-30)
- [mimixbox openvt](#mimixbox-openvt) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-31)
- [mimixbox more / less](#mimixbox-more--less) — 3 scenarios
  - [more streams stdin through when stdout is not a terminal](#scenario-more-streams-stdin-through-when-stdout-is-not-a-terminal)
  - [less streams stdin through when stdout is not a terminal](#scenario-less-streams-stdin-through-when-stdout-is-not-a-terminal)
  - [more streams a file through](#scenario-more-streams-a-file-through)
- [mimixbox reset](#mimixbox-reset) — 1 scenario
  - [prints usage with --help and exits 0](#scenario-prints-usage-with---help-and-exits-0-1)
- [mimixbox resize](#mimixbox-resize) — 1 scenario
  - [shows the usage line for --help](#scenario-shows-the-usage-line-for---help)
- [mimixbox rfkill](#mimixbox-rfkill) — 2 scenarios
  - [lists devices cleanly](#scenario-lists-devices-cleanly)
  - [rejects an unknown command](#scenario-rejects-an-unknown-command)
- [mimixbox rx](#mimixbox-rx) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-32)
- [mimixbox setconsole](#mimixbox-setconsole) — 2 scenarios
  - [fails on an inaccessible device](#scenario-fails-on-an-inaccessible-device)
  - [describes itself with --help](#scenario-describes-itself-with---help-33)
- [mimixbox setfont](#mimixbox-setfont) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-34)
- [mimixbox setkeycodes](#mimixbox-setkeycodes) — 2 scenarios
  - [requires arguments in pairs](#scenario-requires-arguments-in-pairs)
  - [rejects an invalid scancode](#scenario-rejects-an-invalid-scancode)
- [mimixbox setlogcons](#mimixbox-setlogcons) — 2 scenarios
  - [rejects a non-numeric VT](#scenario-rejects-a-non-numeric-vt-2)
  - [describes itself with --help](#scenario-describes-itself-with---help-35)
- [mimixbox setserial](#mimixbox-setserial) — 3 scenarios
  - [echoes the parsed request with -g](#scenario-echoes-the-parsed-request-with--g)
  - [rejects an unknown parameter](#scenario-rejects-an-unknown-parameter)
  - [requires a device](#scenario-requires-a-device)
- [mimixbox showkey](#mimixbox-showkey) — 2 scenarios
  - [rejects conflicting modes](#scenario-rejects-conflicting-modes)
  - [fails deterministically without a console](#scenario-fails-deterministically-without-a-console)
- [mimixbox stty](#mimixbox-stty) — 1 scenario
  - [reports when standard input is not a terminal](#scenario-reports-when-standard-input-is-not-a-terminal)
- [mimixbox ts](#mimixbox-ts) — 1 scenario
  - [prefixes each line with a timestamp](#scenario-prefixes-each-line-with-a-timestamp)
- [mimixbox ttysize](#mimixbox-ttysize) — 2 scenarios
  - [prints width and height](#scenario-prints-width-and-height)
  - [prints just the width with w](#scenario-prints-just-the-width-with-w)
- [mimixbox coreutils slice](#mimixbox-coreutils-slice) — 4 scenarios
  - [factor prints prime factors](#scenario-factor-prints-prime-factors)
  - [tsort topologically sorts](#scenario-tsort-topologically-sorts)
  - [egrep uses extended regular expressions](#scenario-egrep-uses-extended-regular-expressions)
  - [fgrep matches fixed strings literally](#scenario-fgrep-matches-fixed-strings-literally)
- [mimixbox add-shell](#mimixbox-add-shell) — 2 scenarios
  - [prints usage with --help and exits 0](#scenario-prints-usage-with---help-and-exits-0-2)
  - [fails with a message when given no operand](#scenario-fails-with-a-message-when-given-no-operand)
- [mimixbox ischroot](#mimixbox-ischroot) — 1 scenario
  - [prints usage with --help and exits 0](#scenario-prints-usage-with---help-and-exits-0-3)
- [mimixbox mktemp](#mimixbox-mktemp) — 3 scenarios
  - [creates a regular file under the temp dir](#scenario-creates-a-regular-file-under-the-temp-dir)
  - [creates a directory](#scenario-creates-a-directory)
  - [mktemp -u only prints a name](#scenario-mktemp--u-only-prints-a-name)
- [mimixbox remove-shell](#mimixbox-remove-shell) — 2 scenarios
  - [prints usage with --help and exits 0](#scenario-prints-usage-with---help-and-exits-0-4)
  - [fails with a message when given no operand](#scenario-fails-with-a-message-when-given-no-operand-1)
- [mimixbox valid-shell](#mimixbox-valid-shell) — 2 scenarios
  - [prints usage with --help and exits 0](#scenario-prints-usage-with---help-and-exits-0-5)
  - [accepts a file listing existing shells](#scenario-accepts-a-file-listing-existing-shells)
- [mimixbox awk](#mimixbox-awk) — 4 scenarios
  - [prints a field](#scenario-prints-a-field)
  - [honors -F](#scenario-honors--f)
  - [selects a record with NR](#scenario-selects-a-record-with-nr)
  - [counts records in END](#scenario-counts-records-in-end)
- [mimixbox diff](#mimixbox-diff) — 3 scenarios
  - [reports a change in normal format](#scenario-reports-a-change-in-normal-format)
  - [is silent and succeeds for identical files](#scenario-is-silent-and-succeeds-for-identical-files)
  - [reports briefly with -q](#scenario-reports-briefly-with--q)
- [mimixbox ed](#mimixbox-ed) — 3 scenarios
  - [prints the buffer with size](#scenario-prints-the-buffer-with-size)
  - [appends a line and writes it](#scenario-appends-a-line-and-writes-it)
  - [substitutes text on a line](#scenario-substitutes-text-on-a-line)
- [mimixbox patch](#mimixbox-patch) — 1 scenario
  - [applies a unified diff](#scenario-applies-a-unified-diff)
- [mimixbox sed](#mimixbox-sed) — 4 scenarios
  - [substitutes the first match](#scenario-substitutes-the-first-match)
  - [substitutes globally](#scenario-substitutes-globally)
  - [deletes a line by number](#scenario-deletes-a-line-by-number)
  - [prints a single line with -n](#scenario-prints-a-single-line-with--n)
- [mimixbox vi](#mimixbox-vi) — 8 scenarios
  - [deletes a character and writes the file](#scenario-deletes-a-character-and-writes-the-file)
  - [inserts text and writes the file](#scenario-inserts-text-and-writes-the-file)
  - [creates a new file](#scenario-creates-a-new-file)
  - [treats an arrow-key escape sequence as a motion, not an edit](#scenario-treats-an-arrow-key-escape-sequence-as-a-motion-not-an-edit)
  - [duplicates a line with yy then p](#scenario-duplicates-a-line-with-yy-then-p)
  - [applies a count to an edit (2x)](#scenario-applies-a-count-to-an-edit-2x)
  - [undoes the last change with u](#scenario-undoes-the-last-change-with-u)
  - [searches with /pattern and moves to the next match with n](#scenario-searches-with-pattern-and-moves-to-the-next-match-with-n)
- [mimixbox devmem](#mimixbox-devmem) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-36)
- [mimixbox capability-gated applets](#mimixbox-capability-gated-applets) — 3 scenarios
  - [netctl: brctl addbr prints the plan then fails with a capability-gated backend error](#scenario-netctl-brctl-addbr-prints-the-plan-then-fails-with-a-capability-gated-backend-error)
  - [selinux: setenforce refuses to mutate SELinux state and exits non-zero](#scenario-selinux-setenforce-refuses-to-mutate-selinux-state-and-exits-non-zero)
  - [modutils: modprobe validates the module then fails on the CAP_SYS_MODULE gate](#scenario-modutils-modprobe-validates-the-module-then-fails-on-the-cap_sys_module-gate)
- [mimixbox getfattr](#mimixbox-getfattr) — 4 scenarios
  - [dumps a user attribute set by setfattr (or skips without xattr support)](#scenario-dumps-a-user-attribute-set-by-setfattr-or-skips-without-xattr-support)
  - [fails when no file operand is given](#scenario-fails-when-no-file-operand-is-given)
  - [prints usage for --help](#scenario-prints-usage-for---help)
  - [prints the version line for --version](#scenario-prints-the-version-line-for---version)
- [mimixbox compression --help contract](#mimixbox-compression---help-contract) — 12 scenarios
  - [xz --help exposes the documented sections](#scenario-xz---help-exposes-the-documented-sections)
  - [unxz --help exposes the documented sections](#scenario-unxz---help-exposes-the-documented-sections)
  - [xzcat --help exposes the documented sections](#scenario-xzcat---help-exposes-the-documented-sections)
  - [lzma --help exposes the documented sections](#scenario-lzma---help-exposes-the-documented-sections)
  - [unlzma --help exposes the documented sections](#scenario-unlzma---help-exposes-the-documented-sections)
  - [lzcat --help exposes the documented sections](#scenario-lzcat---help-exposes-the-documented-sections)
  - [lzop --help exposes the documented sections](#scenario-lzop---help-exposes-the-documented-sections)
  - [unlzop --help exposes the documented sections](#scenario-unlzop---help-exposes-the-documented-sections)
  - [lzopcat --help exposes the documented sections](#scenario-lzopcat---help-exposes-the-documented-sections)
  - [zcat --help exposes the documented sections](#scenario-zcat---help-exposes-the-documented-sections)
  - [bzcat --help exposes the documented sections](#scenario-bzcat---help-exposes-the-documented-sections)
  - [unit --help exposes the documented sections](#scenario-unit---help-exposes-the-documented-sections)
- [mimixbox --help exit-status contract](#mimixbox---help-exit-status-contract) — 37 scenarios
  - [ash --help exposes the documented sections](#scenario-ash---help-exposes-the-documented-sections)
  - [bash --help exposes the documented sections](#scenario-bash---help-exposes-the-documented-sections)
  - [bc --help exposes the documented sections](#scenario-bc---help-exposes-the-documented-sections)
  - [busybox --help exposes the documented sections](#scenario-busybox---help-exposes-the-documented-sections)
  - [cttyhack --help exposes the documented sections](#scenario-cttyhack---help-exposes-the-documented-sections)
  - [dc --help exposes the documented sections](#scenario-dc---help-exposes-the-documented-sections)
  - [ed --help exposes the documented sections](#scenario-ed---help-exposes-the-documented-sections)
  - [hd --help exposes the documented sections](#scenario-hd---help-exposes-the-documented-sections)
  - [hexdump --help exposes the documented sections](#scenario-hexdump---help-exposes-the-documented-sections)
  - [hush --help exposes the documented sections](#scenario-hush---help-exposes-the-documented-sections)
  - [iostat --help exposes the documented sections](#scenario-iostat---help-exposes-the-documented-sections)
  - [ipcs --help exposes the documented sections](#scenario-ipcs---help-exposes-the-documented-sections)
  - [last --help exposes the documented sections](#scenario-last---help-exposes-the-documented-sections)
  - [less --help exposes the documented sections](#scenario-less---help-exposes-the-documented-sections)
  - [lsblk --help exposes the documented sections](#scenario-lsblk---help-exposes-the-documented-sections)
  - [lspci --help exposes the documented sections](#scenario-lspci---help-exposes-the-documented-sections)
  - [lsusb --help exposes the documented sections](#scenario-lsusb---help-exposes-the-documented-sections)
  - [mbsh --help exposes the documented sections](#scenario-mbsh---help-exposes-the-documented-sections)
  - [minips --help exposes the documented sections](#scenario-minips---help-exposes-the-documented-sections)
  - [more --help exposes the documented sections](#scenario-more---help-exposes-the-documented-sections)
  - [mpstat --help exposes the documented sections](#scenario-mpstat---help-exposes-the-documented-sections)
  - [nmeter --help exposes the documented sections](#scenario-nmeter---help-exposes-the-documented-sections)
  - [pipe_progress --help exposes the documented sections](#scenario-pipe_progress---help-exposes-the-documented-sections)
  - [powertop --help exposes the documented sections](#scenario-powertop---help-exposes-the-documented-sections)
  - [ps --help exposes the documented sections](#scenario-ps---help-exposes-the-documented-sections)
  - [pstree --help exposes the documented sections](#scenario-pstree---help-exposes-the-documented-sections)
  - [sh --help exposes the documented sections](#scenario-sh---help-exposes-the-documented-sections)
  - [smemcap --help exposes the documented sections](#scenario-smemcap---help-exposes-the-documented-sections)
  - [top --help exposes the documented sections](#scenario-top---help-exposes-the-documented-sections)
  - [uptime --help exposes the documented sections](#scenario-uptime---help-exposes-the-documented-sections)
  - [users --help exposes the documented sections](#scenario-users---help-exposes-the-documented-sections)
  - [uudecode --help exposes the documented sections](#scenario-uudecode---help-exposes-the-documented-sections)
  - [uuencode --help exposes the documented sections](#scenario-uuencode---help-exposes-the-documented-sections)
  - [vi --help exposes the documented sections](#scenario-vi---help-exposes-the-documented-sections)
  - [vmstat --help exposes the documented sections](#scenario-vmstat---help-exposes-the-documented-sections)
  - [w --help exposes the documented sections](#scenario-w---help-exposes-the-documented-sections)
  - [wall --help exposes the documented sections](#scenario-wall---help-exposes-the-documented-sections)
- [mimixbox embedded --help helpers](#mimixbox-embedded---help-helpers) — 12 scenarios
  - [devmem --help is structured](#scenario-devmem---help-is-structured)
  - [i2cdetect --help is structured](#scenario-i2cdetect---help-is-structured)
  - [i2cdump --help is structured](#scenario-i2cdump---help-is-structured)
  - [i2cget --help is structured](#scenario-i2cget---help-is-structured)
  - [i2cset --help is structured](#scenario-i2cset---help-is-structured)
  - [partprobe --help is structured](#scenario-partprobe---help-is-structured)
  - [raidautorun --help is structured](#scenario-raidautorun---help-is-structured)
  - [readahead --help is structured](#scenario-readahead---help-is-structured)
  - [resume --help is structured](#scenario-resume---help-is-structured)
  - [seedrng --help is structured](#scenario-seedrng---help-is-structured)
  - [volname --help is structured](#scenario-volname---help-is-structured)
  - [watchdog --help is structured](#scenario-watchdog---help-is-structured)
- [mimixbox --help Notes contract](#mimixbox---help-notes-contract) — 12 scenarios
  - [acpid --help exposes the documented sections](#scenario-acpid---help-exposes-the-documented-sections)
  - [brctl --help exposes the documented sections](#scenario-brctl---help-exposes-the-documented-sections)
  - [crond --help exposes the documented sections](#scenario-crond---help-exposes-the-documented-sections)
  - [ifenslave --help exposes the documented sections](#scenario-ifenslave---help-exposes-the-documented-sections)
  - [mkfs.reiser --help exposes the documented sections](#scenario-mkfsreiser---help-exposes-the-documented-sections)
  - [nbd-client --help exposes the documented sections](#scenario-nbd-client---help-exposes-the-documented-sections)
  - [ssl_server --help exposes the documented sections](#scenario-ssl_server---help-exposes-the-documented-sections)
  - [tunctl --help exposes the documented sections](#scenario-tunctl---help-exposes-the-documented-sections)
  - [vconfig --help exposes the documented sections](#scenario-vconfig---help-exposes-the-documented-sections)
  - [zcip --help exposes the documented sections](#scenario-zcip---help-exposes-the-documented-sections)
  - [\[ --help exposes the documented sections](#scenario----help-exposes-the-documented-sections)
  - [\[\[ --help exposes the documented sections](#scenario----help-exposes-the-documented-sections-1)
- [mimixbox structured --help sections](#mimixbox-structured---help-sections) — 91 scenarios
  - [ln --help exposes the documented sections](#scenario-ln---help-exposes-the-documented-sections)
  - [log-collect --help exposes the documented sections](#scenario-log-collect---help-exposes-the-documented-sections)
  - [logname --help exposes the documented sections](#scenario-logname---help-exposes-the-documented-sections)
  - [md5sum --help exposes the documented sections](#scenario-md5sum---help-exposes-the-documented-sections)
  - [mkdir --help exposes the documented sections](#scenario-mkdir---help-exposes-the-documented-sections)
  - [mkfifo --help exposes the documented sections](#scenario-mkfifo---help-exposes-the-documented-sections)
  - [mknod --help exposes the documented sections](#scenario-mknod---help-exposes-the-documented-sections)
  - [mktemp --help exposes the documented sections](#scenario-mktemp---help-exposes-the-documented-sections)
  - [mountpoint --help exposes the documented sections](#scenario-mountpoint---help-exposes-the-documented-sections)
  - [mv --help exposes the documented sections](#scenario-mv---help-exposes-the-documented-sections)
  - [nc --help exposes the documented sections](#scenario-nc---help-exposes-the-documented-sections)
  - [netcat --help exposes the documented sections](#scenario-netcat---help-exposes-the-documented-sections)
  - [nl --help exposes the documented sections](#scenario-nl---help-exposes-the-documented-sections)
  - [nohup --help exposes the documented sections](#scenario-nohup---help-exposes-the-documented-sections)
  - [nproc --help exposes the documented sections](#scenario-nproc---help-exposes-the-documented-sections)
  - [nyancat --help exposes the documented sections](#scenario-nyancat---help-exposes-the-documented-sections)
  - [od --help exposes the documented sections](#scenario-od---help-exposes-the-documented-sections)
  - [paste --help exposes the documented sections](#scenario-paste---help-exposes-the-documented-sections)
  - [patch --help exposes the documented sections](#scenario-patch---help-exposes-the-documented-sections)
  - [path --help exposes the documented sections](#scenario-path---help-exposes-the-documented-sections)
  - [pidof --help exposes the documented sections](#scenario-pidof---help-exposes-the-documented-sections)
  - [ping --help exposes the documented sections](#scenario-ping---help-exposes-the-documented-sections)
  - [posixer --help exposes the documented sections](#scenario-posixer---help-exposes-the-documented-sections)
  - [poweroff --help exposes the documented sections](#scenario-poweroff---help-exposes-the-documented-sections)
  - [printenv --help exposes the documented sections](#scenario-printenv---help-exposes-the-documented-sections)
  - [pwcrack --help exposes the documented sections](#scenario-pwcrack---help-exposes-the-documented-sections)
  - [pwgen --help exposes the documented sections](#scenario-pwgen---help-exposes-the-documented-sections)
  - [pwscore --help exposes the documented sections](#scenario-pwscore---help-exposes-the-documented-sections)
  - [readlink --help exposes the documented sections](#scenario-readlink---help-exposes-the-documented-sections)
  - [realpath --help exposes the documented sections](#scenario-realpath---help-exposes-the-documented-sections)
  - [reboot --help exposes the documented sections](#scenario-reboot---help-exposes-the-documented-sections)
  - [remove-shell --help exposes the documented sections](#scenario-remove-shell---help-exposes-the-documented-sections)
  - [reset --help exposes the documented sections](#scenario-reset---help-exposes-the-documented-sections)
  - [resize --help exposes the documented sections](#scenario-resize---help-exposes-the-documented-sections)
  - [rev --help exposes the documented sections](#scenario-rev---help-exposes-the-documented-sections)
  - [rm --help exposes the documented sections](#scenario-rm---help-exposes-the-documented-sections)
  - [rmdir --help exposes the documented sections](#scenario-rmdir---help-exposes-the-documented-sections)
  - [rpm --help exposes the documented sections](#scenario-rpm---help-exposes-the-documented-sections)
  - [rpm2cpio --help exposes the documented sections](#scenario-rpm2cpio---help-exposes-the-documented-sections)
  - [sddf --help exposes the documented sections](#scenario-sddf---help-exposes-the-documented-sections)
  - [sed --help exposes the documented sections](#scenario-sed---help-exposes-the-documented-sections)
  - [seq --help exposes the documented sections](#scenario-seq---help-exposes-the-documented-sections)
  - [serial --help exposes the documented sections](#scenario-serial---help-exposes-the-documented-sections)
  - [sha1sum --help exposes the documented sections](#scenario-sha1sum---help-exposes-the-documented-sections)
  - [sha256sum --help exposes the documented sections](#scenario-sha256sum---help-exposes-the-documented-sections)
  - [sha384sum --help exposes the documented sections](#scenario-sha384sum---help-exposes-the-documented-sections)
  - [sha3sum --help exposes the documented sections](#scenario-sha3sum---help-exposes-the-documented-sections)
  - [sha512sum --help exposes the documented sections](#scenario-sha512sum---help-exposes-the-documented-sections)
  - [shred --help exposes the documented sections](#scenario-shred---help-exposes-the-documented-sections)
  - [shuf --help exposes the documented sections](#scenario-shuf---help-exposes-the-documented-sections)
  - [sl --help exposes the documented sections](#scenario-sl---help-exposes-the-documented-sections)
  - [sleep --help exposes the documented sections](#scenario-sleep---help-exposes-the-documented-sections)
  - [sort --help exposes the documented sections](#scenario-sort---help-exposes-the-documented-sections)
  - [speaker --help exposes the documented sections](#scenario-speaker---help-exposes-the-documented-sections)
  - [split --help exposes the documented sections](#scenario-split---help-exposes-the-documented-sections)
  - [stat --help exposes the documented sections](#scenario-stat---help-exposes-the-documented-sections)
  - [strings --help exposes the documented sections](#scenario-strings---help-exposes-the-documented-sections)
  - [sync --help exposes the documented sections](#scenario-sync---help-exposes-the-documented-sections)
  - [tac --help exposes the documented sections](#scenario-tac---help-exposes-the-documented-sections)
  - [tar --help exposes the documented sections](#scenario-tar---help-exposes-the-documented-sections)
  - [tee --help exposes the documented sections](#scenario-tee---help-exposes-the-documented-sections)
  - [timeout --help exposes the documented sections](#scenario-timeout---help-exposes-the-documented-sections)
  - [touch --help exposes the documented sections](#scenario-touch---help-exposes-the-documented-sections)
  - [tr --help exposes the documented sections](#scenario-tr---help-exposes-the-documented-sections)
  - [truncate --help exposes the documented sections](#scenario-truncate---help-exposes-the-documented-sections)
  - [tty --help exposes the documented sections](#scenario-tty---help-exposes-the-documented-sections)
  - [uname --help exposes the documented sections](#scenario-uname---help-exposes-the-documented-sections)
  - [uncompress --help exposes the documented sections](#scenario-uncompress---help-exposes-the-documented-sections)
  - [unexpand --help exposes the documented sections](#scenario-unexpand---help-exposes-the-documented-sections)
  - [uniq --help exposes the documented sections](#scenario-uniq---help-exposes-the-documented-sections)
  - [unix2dos --help exposes the documented sections](#scenario-unix2dos---help-exposes-the-documented-sections)
  - [unlink --help exposes the documented sections](#scenario-unlink---help-exposes-the-documented-sections)
  - [unshadow --help exposes the documented sections](#scenario-unshadow---help-exposes-the-documented-sections)
  - [unzip --help exposes the documented sections](#scenario-unzip---help-exposes-the-documented-sections)
  - [uuidgen --help exposes the documented sections](#scenario-uuidgen---help-exposes-the-documented-sections)
  - [valid-shell --help exposes the documented sections](#scenario-valid-shell---help-exposes-the-documented-sections)
  - [watch --help exposes the documented sections](#scenario-watch---help-exposes-the-documented-sections)
  - [wc --help exposes the documented sections](#scenario-wc---help-exposes-the-documented-sections)
  - [which --help exposes the documented sections](#scenario-which---help-exposes-the-documented-sections)
  - [who --help exposes the documented sections](#scenario-who---help-exposes-the-documented-sections)
  - [whoami --help exposes the documented sections](#scenario-whoami---help-exposes-the-documented-sections)
  - [whris --help exposes the documented sections](#scenario-whris---help-exposes-the-documented-sections)
  - [xargs --help exposes the documented sections](#scenario-xargs---help-exposes-the-documented-sections)
  - [xxd --help exposes the documented sections](#scenario-xxd---help-exposes-the-documented-sections)
  - [yes --help exposes the documented sections](#scenario-yes---help-exposes-the-documented-sections)
  - [zip --help exposes the documented sections](#scenario-zip---help-exposes-the-documented-sections)
  - [zip-pwcrack --help exposes the documented sections](#scenario-zip-pwcrack---help-exposes-the-documented-sections)
  - [true --help exposes the documented sections](#scenario-true---help-exposes-the-documented-sections)
  - [test --help exposes the documented sections](#scenario-test---help-exposes-the-documented-sections)
  - [printf --help exposes the documented sections](#scenario-printf---help-exposes-the-documented-sections)
  - [pwd --help exposes the documented sections](#scenario-pwd---help-exposes-the-documented-sections)
- [mimixbox structured --help sections (2)](#mimixbox-structured---help-sections-2) — 60 scenarios
  - [add-shell --help exposes the documented sections](#scenario-add-shell---help-exposes-the-documented-sections)
  - [ar --help exposes the documented sections](#scenario-ar---help-exposes-the-documented-sections)
  - [arch --help exposes the documented sections](#scenario-arch---help-exposes-the-documented-sections)
  - [awk --help exposes the documented sections](#scenario-awk---help-exposes-the-documented-sections)
  - [banner --help exposes the documented sections](#scenario-banner---help-exposes-the-documented-sections)
  - [base32 --help exposes the documented sections](#scenario-base32---help-exposes-the-documented-sections)
  - [base64 --help exposes the documented sections](#scenario-base64---help-exposes-the-documented-sections)
  - [basename --help exposes the documented sections](#scenario-basename---help-exposes-the-documented-sections)
  - [bunzip2 --help exposes the documented sections](#scenario-bunzip2---help-exposes-the-documented-sections)
  - [cal --help exposes the documented sections](#scenario-cal---help-exposes-the-documented-sections)
  - [cat --help exposes the documented sections](#scenario-cat---help-exposes-the-documented-sections)
  - [chgrp --help exposes the documented sections](#scenario-chgrp---help-exposes-the-documented-sections)
  - [chmod --help exposes the documented sections](#scenario-chmod---help-exposes-the-documented-sections)
  - [chown --help exposes the documented sections](#scenario-chown---help-exposes-the-documented-sections)
  - [cksum --help exposes the documented sections](#scenario-cksum---help-exposes-the-documented-sections)
  - [clear --help exposes the documented sections](#scenario-clear---help-exposes-the-documented-sections)
  - [cmatrix --help exposes the documented sections](#scenario-cmatrix---help-exposes-the-documented-sections)
  - [cmp --help exposes the documented sections](#scenario-cmp---help-exposes-the-documented-sections)
  - [comm --help exposes the documented sections](#scenario-comm---help-exposes-the-documented-sections)
  - [compress --help exposes the documented sections](#scenario-compress---help-exposes-the-documented-sections)
  - [cowsay --help exposes the documented sections](#scenario-cowsay---help-exposes-the-documented-sections)
  - [cowthink --help exposes the documented sections](#scenario-cowthink---help-exposes-the-documented-sections)
  - [cpio --help exposes the documented sections](#scenario-cpio---help-exposes-the-documented-sections)
  - [cut --help exposes the documented sections](#scenario-cut---help-exposes-the-documented-sections)
  - [date --help exposes the documented sections](#scenario-date---help-exposes-the-documented-sections)
  - [dd --help exposes the documented sections](#scenario-dd---help-exposes-the-documented-sections)
  - [df --help exposes the documented sections](#scenario-df---help-exposes-the-documented-sections)
  - [diff --help exposes the documented sections](#scenario-diff---help-exposes-the-documented-sections)
  - [dirname --help exposes the documented sections](#scenario-dirname---help-exposes-the-documented-sections)
  - [dos2unix --help exposes the documented sections](#scenario-dos2unix---help-exposes-the-documented-sections)
  - [du --help exposes the documented sections](#scenario-du---help-exposes-the-documented-sections)
  - [egrep --help exposes the documented sections](#scenario-egrep---help-exposes-the-documented-sections)
  - [env --help exposes the documented sections](#scenario-env---help-exposes-the-documented-sections)
  - [expand --help exposes the documented sections](#scenario-expand---help-exposes-the-documented-sections)
  - [expr --help exposes the documented sections](#scenario-expr---help-exposes-the-documented-sections)
  - [fakemovie --help exposes the documented sections](#scenario-fakemovie---help-exposes-the-documented-sections)
  - [fgrep --help exposes the documented sections](#scenario-fgrep---help-exposes-the-documented-sections)
  - [fmt --help exposes the documented sections](#scenario-fmt---help-exposes-the-documented-sections)
  - [fold --help exposes the documented sections](#scenario-fold---help-exposes-the-documented-sections)
  - [fortune --help exposes the documented sections](#scenario-fortune---help-exposes-the-documented-sections)
  - [free --help exposes the documented sections](#scenario-free---help-exposes-the-documented-sections)
  - [ghrdc --help exposes the documented sections](#scenario-ghrdc---help-exposes-the-documented-sections)
  - [grep --help exposes the documented sections](#scenario-grep---help-exposes-the-documented-sections)
  - [groups --help exposes the documented sections](#scenario-groups---help-exposes-the-documented-sections)
  - [gunzip --help exposes the documented sections](#scenario-gunzip---help-exposes-the-documented-sections)
  - [gzip --help exposes the documented sections](#scenario-gzip---help-exposes-the-documented-sections)
  - [halt --help exposes the documented sections](#scenario-halt---help-exposes-the-documented-sections)
  - [head --help exposes the documented sections](#scenario-head---help-exposes-the-documented-sections)
  - [hostid --help exposes the documented sections](#scenario-hostid---help-exposes-the-documented-sections)
  - [hostname --help exposes the documented sections](#scenario-hostname---help-exposes-the-documented-sections)
  - [http-status-code --help exposes the documented sections](#scenario-http-status-code---help-exposes-the-documented-sections)
  - [id --help exposes the documented sections](#scenario-id---help-exposes-the-documented-sections)
  - [install --help exposes the documented sections](#scenario-install---help-exposes-the-documented-sections)
  - [ischroot --help exposes the documented sections](#scenario-ischroot---help-exposes-the-documented-sections)
  - [killall --help exposes the documented sections](#scenario-killall---help-exposes-the-documented-sections)
  - [lifegame --help exposes the documented sections](#scenario-lifegame---help-exposes-the-documented-sections)
  - [link --help exposes the documented sections](#scenario-link---help-exposes-the-documented-sections)
  - [echo --help exposes the documented sections](#scenario-echo---help-exposes-the-documented-sections)
  - [false --help exposes the documented sections](#scenario-false---help-exposes-the-documented-sections)
  - [kill --help exposes the documented sections](#scenario-kill---help-exposes-the-documented-sections)
- [mimixbox i2cdetect](#mimixbox-i2cdetect) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-37)
- [mimixbox i2cdump](#mimixbox-i2cdump) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-38)
- [mimixbox i2cget](#mimixbox-i2cget) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-39)
- [mimixbox i2cset](#mimixbox-i2cset) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-40)
- [mimixbox ifup](#mimixbox-ifup) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-41)
- [mimixbox insmod](#mimixbox-insmod) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-42)
- [mimixbox ip](#mimixbox-ip) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-43)
- [mimixbox ipaddr](#mimixbox-ipaddr) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-44)
- [mimixbox iplink](#mimixbox-iplink) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-45)
- [mimixbox ipneigh](#mimixbox-ipneigh) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-46)
- [mimixbox iproute](#mimixbox-iproute) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-47)
- [mimixbox iprule](#mimixbox-iprule) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-48)
- [mimixbox iptunnel](#mimixbox-iptunnel) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-49)
- [mimixbox less](#mimixbox-less) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-50)
- [mimixbox linux32](#mimixbox-linux32) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-51)
- [mimixbox linux64](#mimixbox-linux64) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-52)
- [mimixbox linuxrc](#mimixbox-linuxrc) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-53)
- [mimixbox load_policy](#mimixbox-load_policy) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-54)
- [mimixbox log-collect](#mimixbox-log-collect) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-55)
- [mimixbox lpd](#mimixbox-lpd) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-56)
- [mimixbox lpq](#mimixbox-lpq) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-57)
- [mimixbox lpr](#mimixbox-lpr) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-58)
- [mimixbox lpr_roundtrip](#mimixbox-lpr_roundtrip) — 1 scenario
  - [queues, lists, drains, and empties the spool](#scenario-queues-lists-drains-and-empties-the-spool)
- [mimixbox lsscsi](#mimixbox-lsscsi) — 3 scenarios
  - [lists SCSI devices from sysfs without error (empty is allowed)](#scenario-lists-scsi-devices-from-sysfs-without-error-empty-is-allowed)
  - [prints usage for --help](#scenario-prints-usage-for---help-1)
  - [prints the version line for --version](#scenario-prints-the-version-line-for---version-1)
- [mimixbox lzcat](#mimixbox-lzcat) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-59)
- [mimixbox lzma](#mimixbox-lzma) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-60)
- [mimixbox lzop](#mimixbox-lzop) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-61)
- [mimixbox lzopcat](#mimixbox-lzopcat) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-62)
- [mimixbox makedevs](#mimixbox-makedevs) — 4 scenarios
  - [creates the directory and file tree from a device table](#scenario-creates-the-directory-and-file-tree-from-a-device-table)
  - [fails without the -d table option](#scenario-fails-without-the--d-table-option)
  - [prints usage for --help](#scenario-prints-usage-for---help-2)
  - [prints the version line for --version](#scenario-prints-the-version-line-for---version-2)
- [mimixbox matchpathcon](#mimixbox-matchpathcon) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-63)
- [mimixbox mkdosfs](#mimixbox-mkdosfs) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-64)
- [mimixbox mkfs.ext2](#mimixbox-mkfsext2) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-65)
- [mimixbox mkfs.minix](#mimixbox-mkfsminix) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-66)
- [mimixbox mkfs.reiser](#mimixbox-mkfsreiser) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-67)
- [mimixbox mkfs.vfat](#mimixbox-mkfsvfat) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-68)
- [mimixbox modprobe](#mimixbox-modprobe) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-69)
- [mimixbox more](#mimixbox-more) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-70)
- [mimixbox nameif](#mimixbox-nameif) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-71)
- [mimixbox nbd-client](#mimixbox-nbd-client) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-72)
- [mimixbox partprobe](#mimixbox-partprobe) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-73)
- [mimixbox ping6](#mimixbox-ping6) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-74)
- [mimixbox pipe_progress](#mimixbox-pipe_progress) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-75)
- [mimixbox pkill](#mimixbox-pkill) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-76)
- [mimixbox poweroff](#mimixbox-poweroff) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-77)
- [mimixbox preexisting_tmp_root](#mimixbox-preexisting_tmp_root) — 2 scenarios
  - [allocates a usable per-run root that is not /tmp/mimixbox](#scenario-allocates-a-usable-per-run-root-that-is-not-tmpmimixbox)
  - [leaves a pre-existing /tmp/mimixbox file untouched (harness-specific)](#scenario-leaves-a-pre-existing-tmpmimixbox-file-untouched-harness-specific)
- [mimixbox raidautorun](#mimixbox-raidautorun) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-78)
- [mimixbox readahead](#mimixbox-readahead) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-79)
- [mimixbox reboot](#mimixbox-reboot) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-80)
- [mimixbox restorecon](#mimixbox-restorecon) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-81)
- [mimixbox resume](#mimixbox-resume) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-82)
- [mimixbox seedrng](#mimixbox-seedrng) — 2 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-83)
  - [documents its purpose in --help](#scenario-documents-its-purpose-in---help-8)
- [mimixbox setfattr](#mimixbox-setfattr) — 4 scenarios
  - [sets an attribute that getfattr can read back (or skips without xattr support)](#scenario-sets-an-attribute-that-getfattr-can-read-back-or-skips-without-xattr-support)
  - [rejects mutually exclusive -n and -x](#scenario-rejects-mutually-exclusive--n-and--x)
  - [prints usage for --help](#scenario-prints-usage-for---help-3)
  - [prints the version line for --version](#scenario-prints-the-version-line-for---version-3)
- [mimixbox volname](#mimixbox-volname) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-84)
- [mimixbox watchdog](#mimixbox-watchdog) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-85)
- [mimixbox chgrp](#mimixbox-chgrp) — 2 scenarios
  - [prints usage with --help and exits 0](#scenario-prints-usage-with---help-and-exits-0-6)
  - [fails with a message when given no operand](#scenario-fails-with-a-message-when-given-no-operand-2)
- [mimixbox chown](#mimixbox-chown) — 2 scenarios
  - [prints usage with --help and exits 0](#scenario-prints-usage-with---help-and-exits-0-7)
  - [fails with a message when given no operand](#scenario-fails-with-a-message-when-given-no-operand-3)
- [mimixbox cp](#mimixbox-cp) — 7 scenarios
  - [copy one file](#scenario-copy-one-file)
  - [copy directory recursively](#scenario-copy-directory-recursively)
  - [can not copy when src and dest are the same](#scenario-can-not-copy-when-src-and-dest-are-the-same)
  - [status failure when src and dest are the same](#scenario-status-failure-when-src-and-dest-are-the-same)
  - [copy three files at the same time](#scenario-copy-three-files-at-the-same-time)
  - [can not copy a directory without the recursive option](#scenario-can-not-copy-a-directory-without-the-recursive-option)
  - [can not copy a directory to root without authority](#scenario-can-not-copy-a-directory-to-root-without-authority)
- [mimixbox cp GNU flags](#mimixbox-cp-gnu-flags) — 7 scenarios
  - [copies into the target directory (-t equals destination-last form)](#scenario-copies-into-the-target-directory--t-equals-destination-last-form)
  - [rejects a directory destination with --no-target-directory (-T)](#scenario-rejects-a-directory-destination-with---no-target-directory--t)
  - [recreates the source path prefix with --parents](#scenario-recreates-the-source-path-prefix-with---parents)
  - [makes a backup before overwriting with --backup](#scenario-makes-a-backup-before-overwriting-with---backup)
  - [skips the copy when the destination is newer (-u)](#scenario-skips-the-copy-when-the-destination-is-newer--u)
  - [skips a newer file inside the tree with -ru (#940)](#scenario-skips-a-newer-file-inside-the-tree-with--ru-940)
  - [backs up a file inside the tree with -r --backup (#940)](#scenario-backs-up-a-file-inside-the-tree-with--r---backup-940)
- [mimixbox cp symlink handling](#mimixbox-cp-symlink-handling) — 3 scenarios
  - [cp -P copies the symlink as a link](#scenario-cp--p-copies-the-symlink-as-a-link)
  - [cp -L copies the link target as a regular file](#scenario-cp--l-copies-the-link-target-as-a-regular-file)
  - [cp -d preserves a symlink inside a copied tree](#scenario-cp--d-preserves-a-symlink-inside-a-copied-tree)
- [mimixbox fileutils help helpers](#mimixbox-fileutils-help-helpers) — 2 scenarios
  - [chgrp --help is structured](#scenario-chgrp---help-is-structured)
  - [chown --help is structured](#scenario-chown---help-is-structured)
- [mimixbox link](#mimixbox-link) — 1 scenario
  - [creates a hard link sharing contents](#scenario-creates-a-hard-link-sharing-contents)
- [mimixbox ln](#mimixbox-ln) — 3 scenarios
  - [ln creates a hard link to the same content](#scenario-ln-creates-a-hard-link-to-the-same-content)
  - [ln -s creates a symbolic link](#scenario-ln--s-creates-a-symbolic-link)
  - [ln with no operand reports an error](#scenario-ln-with-no-operand-reports-an-error)
- [mimixbox ln GNU flags](#mimixbox-ln-gnu-flags) — 2 scenarios
  - [ln -s --relative stores the target relative to the link location](#scenario-ln--s---relative-stores-the-target-relative-to-the-link-location)
  - [ln --target-directory links each operand into the directory](#scenario-ln---target-directory-links-each-operand-into-the-directory)
- [mimixbox ls](#mimixbox-ls) — 3 scenarios
  - [lists entries sorted, hiding dotfiles](#scenario-lists-entries-sorted-hiding-dotfiles)
  - [includes dotfiles with -a](#scenario-includes-dotfiles-with--a)
  - [marks directories with -F](#scenario-marks-directories-with--f)
- [mimixbox ls GNU flags](#mimixbox-ls-gnu-flags) — 12 scenarios
  - [colors directories with --color=always](#scenario-colors-directories-with---coloralways)
  - [emits no escapes with --color=never](#scenario-emits-no-escapes-with---colornever)
  - [appends / * @ with -F](#scenario-appends----with--f)
  - [omits * for executables with --file-type](#scenario-omits--for-executables-with---file-type)
  - [marks only dirs with --indicator-style=slash](#scenario-marks-only-dirs-with---indicator-styleslash)
  - [lists largest first with --sort=size](#scenario-lists-largest-first-with---sortsize)
  - [lists directories first with --group-directories-first](#scenario-lists-directories-first-with---group-directories-first)
  - [drops matches with --ignore](#scenario-drops-matches-with---ignore)
  - [drops matches with --hide](#scenario-drops-matches-with---hide)
  - [keeps hidden matches when -a is given](#scenario-keeps-hidden-matches-when--a-is-given)
  - [prints an inode number with -i](#scenario-prints-an-inode-number-with--i)
  - [scales sizes to 1024-byte blocks with -k](#scenario-scales-sizes-to-1024-byte-blocks-with--k)
- [mimixbox mkdir](#mimixbox-mkdir) — 7 scenarios
  - [make a single directory](#scenario-make-a-single-directory)
  - [make three directories](#scenario-make-three-directories)
  - [make a parent/child directory with -p](#scenario-make-a-parentchild-directory-with--p)
  - [make a directory from a pipe](#scenario-make-a-directory-from-a-pipe)
  - [print error without operand](#scenario-print-error-without-operand)
  - [print error with --parents and no operand](#scenario-print-error-with---parents-and-no-operand)
  - [make 1 and 3 but fail to make 2 at an unwritable path](#scenario-make-1-and-3-but-fail-to-make-2-at-an-unwritable-path)
- [mimixbox mkfifo](#mimixbox-mkfifo) — 5 scenarios
  - [make one named pipe with mode prw-r--r--](#scenario-make-one-named-pipe-with-mode-prw-r--r--)
  - [make three named pipes with mode prw-r--r--](#scenario-make-three-named-pipes-with-mode-prw-r--r--)
  - [print error for a non-existent path](#scenario-print-error-for-a-non-existent-path)
  - [print error when the same name already exists](#scenario-print-error-when-the-same-name-already-exists)
  - [make two pipes and report the one that failed](#scenario-make-two-pipes-and-report-the-one-that-failed)
- [mimixbox mountpoint](#mimixbox-mountpoint) — 1 scenario
  - [reports that / is a mountpoint](#scenario-reports-that--is-a-mountpoint)
- [mimixbox mv](#mimixbox-mv) — 10 scenarios
  - [rename a file](#scenario-rename-a-file)
  - [move a file into an inner directory](#scenario-move-a-file-into-an-inner-directory)
  - [move three files into an inner directory](#scenario-move-three-files-into-an-inner-directory)
  - [move three files where one does not exist](#scenario-move-three-files-where-one-does-not-exist)
  - [move a directory into a directory](#scenario-move-a-directory-into-a-directory)
  - [move three directories](#scenario-move-three-directories)
  - [move three directories where one does not exist](#scenario-move-three-directories-where-one-does-not-exist)
  - [moving a file onto itself fails](#scenario-moving-a-file-onto-itself-fails)
  - [overwrite a file with the same destination name](#scenario-overwrite-a-file-with-the-same-destination-name)
  - [overwrite with the backup option keeps a tilde copy](#scenario-overwrite-with-the-backup-option-keeps-a-tilde-copy)
- [mimixbox mv GNU flags](#mimixbox-mv-gnu-flags) — 2 scenarios
  - [mv --target-directory moves each source into the directory](#scenario-mv---target-directory-moves-each-source-into-the-directory)
  - [mv --update preserves a newer destination](#scenario-mv---update-preserves-a-newer-destination)
- [mimixbox readlink](#mimixbox-readlink) — 1 scenario
  - [prints the symlink target](#scenario-prints-the-symlink-target)
- [mimixbox readlink_gnu](#mimixbox-readlink_gnu) — 4 scenarios
  - [fails when -e is given a missing path](#scenario-fails-when--e-is-given-a-missing-path)
  - [succeeds when -e is given an existing symlink](#scenario-succeeds-when--e-is-given-an-existing-symlink)
  - [succeeds with -m on a missing path](#scenario-succeeds-with--m-on-a-missing-path)
  - [terminates output with NUL under -z](#scenario-terminates-output-with-nul-under--z)
- [mimixbox rm](#mimixbox-rm) — 6 scenarios
  - [remove one file](#scenario-remove-one-file)
  - [remove files using a wildcard](#scenario-remove-files-using-a-wildcard)
  - [remove three files at the same time](#scenario-remove-three-files-at-the-same-time)
  - [remove two files and report the missing one](#scenario-remove-two-files-and-report-the-missing-one)
  - [can not remove a directory without the recursive option](#scenario-can-not-remove-a-directory-without-the-recursive-option)
  - [remove a directory with the recursive option](#scenario-remove-a-directory-with-the-recursive-option)
- [mimixbox rm GNU flags](#mimixbox-rm-gnu-flags) — 4 scenarios
  - [refuses to recurse on / by default (--preserve-root)](#scenario-refuses-to-recurse-on--by-default---preserve-root)
  - [removes an ordinary directory recursively (guard does not interfere)](#scenario-removes-an-ordinary-directory-recursively-guard-does-not-interfere)
  - [removes a single-filesystem tree with --one-file-system](#scenario-removes-a-single-filesystem-tree-with---one-file-system)
  - [removes a tree when --no-preserve-root is given (safe target)](#scenario-removes-a-tree-when---no-preserve-root-is-given-safe-target)
- [mimixbox rmdir](#mimixbox-rmdir) — 4 scenarios
  - [removes an empty directory](#scenario-removes-an-empty-directory)
  - [fails on a non-empty directory](#scenario-fails-on-a-non-empty-directory)
  - [rmdir -p removes nested empty directories](#scenario-rmdir--p-removes-nested-empty-directories)
  - [rmdir with no operand reports an error](#scenario-rmdir-with-no-operand-reports-an-error)
- [mimixbox serial](#mimixbox-serial) — 3 scenarios
  - [adds a serial-number prefix to each file](#scenario-adds-a-serial-number-prefix-to-each-file)
  - [--dry-run does not rename anything](#scenario---dry-run-does-not-rename-anything)
  - [serial with no operand reports an error](#scenario-serial-with-no-operand-reports-an-error)
- [mimixbox shred](#mimixbox-shred) — 1 scenario
  - [overwrites and removes the file](#scenario-overwrites-and-removes-the-file)
- [mimixbox stat](#mimixbox-stat) — 1 scenario
  - [prints the size with a custom format](#scenario-prints-the-size-with-a-custom-format)
- [mimixbox stat GNU flags](#mimixbox-stat-gnu-flags) — 5 scenarios
  - [prints name and size via --printf with no trailing newline](#scenario-prints-name-and-size-via---printf-with-no-trailing-newline)
  - [interprets backslash escapes in --printf](#scenario-interprets-backslash-escapes-in---printf)
  - [appends a trailing newline for --format](#scenario-appends-a-trailing-newline-for---format)
  - [prints a single space-separated terse line](#scenario-prints-a-single-space-separated-terse-line)
  - [reports the size as the second terse field](#scenario-reports-the-size-as-the-second-terse-field)
- [mimixbox touch](#mimixbox-touch) — 3 scenarios
  - [make one file](#scenario-make-one-file)
  - [make three files at the same time](#scenario-make-three-files-at-the-same-time)
  - [make two files and fail to make one at an unwritable path](#scenario-make-two-files-and-fail-to-make-one-at-an-unwritable-path)
- [mimixbox touch GNU flags](#mimixbox-touch-gnu-flags) — 5 scenarios
  - [copies the reference file mtime (--reference)](#scenario-copies-the-reference-file-mtime---reference)
  - [sets a known time with --date](#scenario-sets-a-known-time-with---date)
  - [accepts --time=atime without error](#scenario-accepts---timeatime-without-error)
  - [rejects an invalid --time word](#scenario-rejects-an-invalid---time-word)
  - [changes the symlink itself with --no-dereference (-h)](#scenario-changes-the-symlink-itself-with---no-dereference--h)
- [mimixbox truncate](#mimixbox-truncate) — 1 scenario
  - [sets the file to the given size](#scenario-sets-the-file-to-the-given-size)
- [mimixbox unlink](#mimixbox-unlink) — 1 scenario
  - [removes a single file](#scenario-removes-a-single-file)
- [mimixbox alias parity](#mimixbox-alias-parity) — 4 scenarios
  - [egrep matches the same lines as grep -E over the same file](#scenario-egrep-matches-the-same-lines-as-grep--e-over-the-same-file)
  - [fgrep matches the same lines as grep -F over the same file](#scenario-fgrep-matches-the-same-lines-as-grep--f-over-the-same-file)
  - [netcat answers --help with exit 0 and an Examples section](#scenario-netcat-answers---help-with-exit-0-and-an-examples-section)
  - [nc answers --help with exit 0 and an Examples section](#scenario-nc-answers---help-with-exit-0-and-an-examples-section)
- [mimixbox egrep](#mimixbox-egrep) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-86)
- [mimixbox fgrep](#mimixbox-fgrep) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-87)
- [mimixbox find](#mimixbox-find) — 7 scenarios
  - [finds a file by -name](#scenario-finds-a-file-by--name)
  - [lists directories with -type d](#scenario-lists-directories-with--type-d)
  - [rejects an unknown predicate](#scenario-rejects-an-unknown-predicate)
  - [prints usage for --help](#scenario-prints-usage-for---help-4)
  - [lists an Options block with --help and --version in --help](#scenario-lists-an-options-block-with---help-and---version-in---help)
  - [documents the supported subset tokens in --help](#scenario-documents-the-supported-subset-tokens-in---help)
  - [prints the version line for --version](#scenario-prints-the-version-line-for---version-4)
- [mimixbox grep](#mimixbox-grep) — 4 scenarios
  - [matches lines from stdin](#scenario-matches-lines-from-stdin)
  - [matches lines from a file](#scenario-matches-lines-from-a-file)
  - [counts matching lines with -c](#scenario-counts-matching-lines-with--c)
  - [exits 1 when nothing matches](#scenario-exits-1-when-nothing-matches)
- [mimixbox grep GNU flags](#mimixbox-grep-gnu-flags) — 10 scenarios
  - [prints trailing context with -A1](#scenario-prints-trailing-context-with--a1)
  - [prints leading context with -B1](#scenario-prints-leading-context-with--b1)
  - [prints surrounding context with -C1](#scenario-prints-surrounding-context-with--c1)
  - [separates non-contiguous groups with --](#scenario-separates-non-contiguous-groups-with---)
  - [searches only included files with --include](#scenario-searches-only-included-files-with---include)
  - [skips excluded files with --exclude](#scenario-skips-excluded-files-with---exclude)
  - [skips excluded directories with --exclude-dir](#scenario-skips-excluded-directories-with---exclude-dir)
  - [highlights matches with --color=always](#scenario-highlights-matches-with---coloralways)
  - [prints byte offsets with -b](#scenario-prints-byte-offsets-with--b)
  - [prints files without a match with -L](#scenario-prints-files-without-a-match-with--l)
- [mimixbox findutils help helpers](#mimixbox-findutils-help-helpers) — 2 scenarios
  - [egrep --help is structured](#scenario-egrep---help-is-structured)
  - [fgrep --help is structured](#scenario-fgrep---help-is-structured)
- [mimixbox xargs](#mimixbox-xargs) — 3 scenarios
  - [appends stdin items to the command](#scenario-appends-stdin-items-to-the-command)
  - [splits into groups with -n](#scenario-splits-into-groups-with--n)
  - [substitutes with -I](#scenario-substitutes-with--i)
- [mimixbox xargs GNU flags](#mimixbox-xargs-gnu-flags) — 6 scenarios
  - [runs once per input line with -L 1](#scenario-runs-once-per-input-line-with--l-1)
  - [groups two input lines per invocation with -L 2](#scenario-groups-two-input-lines-per-invocation-with--l-2)
  - [splits a long input into multiple invocations with -s](#scenario-splits-a-long-input-into-multiple-invocations-with--s)
  - [keeps every item across the -s split invocations](#scenario-keeps-every-item-across-the--s-split-invocations)
  - [runs all batches concurrently with -P 4](#scenario-runs-all-batches-concurrently-with--p-4)
  - [runs all batches with -P 0](#scenario-runs-all-batches-with--p-0)
- [mimixbox lifegame CLI contract](#mimixbox-lifegame-cli-contract) — 1 scenario
  - [prints usage with --help and exits 0](#scenario-prints-usage-with---help-and-exits-0-8)
- [mimixbox banner](#mimixbox-banner) — 1 scenario
  - [prints five rows of art](#scenario-prints-five-rows-of-art)
- [mimixbox cmatrix](#mimixbox-cmatrix) — 1 scenario
  - [exits gracefully without a terminal](#scenario-exits-gracefully-without-a-terminal)
- [mimixbox cowsay](#mimixbox-cowsay) — 2 scenarios
  - [prints usage with --help and exits 0](#scenario-prints-usage-with---help-and-exits-0-9)
  - [renders the message in the speech bubble](#scenario-renders-the-message-in-the-speech-bubble)
- [mimixbox cowthink](#mimixbox-cowthink) — 1 scenario
  - [draws the thought-bubble connector](#scenario-draws-the-thought-bubble-connector)
- [mimixbox fakemovie](#mimixbox-fakemovie) — 2 scenarios
  - [prints usage with --help and exits 0](#scenario-prints-usage-with---help-and-exits-0-10)
  - [fails with a message when given no operand](#scenario-fails-with-a-message-when-given-no-operand-4)
- [mimixbox fortune](#mimixbox-fortune) — 1 scenario
  - [prints a single adage line](#scenario-prints-a-single-adage-line)
- [mimixbox jokeutils --help contract](#mimixbox-jokeutils---help-contract) — 1 scenario
  - [sl --help is structured](#scenario-sl---help-is-structured)
- [mimixbox nyancat](#mimixbox-nyancat) — 1 scenario
  - [exits gracefully without a terminal](#scenario-exits-gracefully-without-a-terminal-1)
- [mimixbox sl](#mimixbox-sl) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-88)
- [mimixbox acpid](#mimixbox-acpid) — 2 scenarios
  - [requires foreground mode](#scenario-requires-foreground-mode)
  - [describes itself with --help](#scenario-describes-itself-with---help-89)
- [mimixbox addgroup](#mimixbox-addgroup) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-90)
- [mimixbox adduser](#mimixbox-adduser) — 2 scenarios
  - [requires a user name](#scenario-requires-a-user-name)
  - [describes itself with --help](#scenario-describes-itself-with---help-91)
- [mimixbox bootchartd](#mimixbox-bootchartd) — 1 scenario
  - [records a proc_stat sample](#scenario-records-a-proc_stat-sample)
- [mimixbox chpasswd](#mimixbox-chpasswd) — 2 scenarios
  - [rejects an unknown method](#scenario-rejects-an-unknown-method)
  - [describes itself with --help](#scenario-describes-itself-with---help-92)
- [mimixbox chsh](#mimixbox-chsh) — 4 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-93)
  - [lists shells and exits successfully](#scenario-lists-shells-and-exits-successfully)
  - [rejects an unknown user](#scenario-rejects-an-unknown-user)
  - [rejects a relative shell path](#scenario-rejects-a-relative-shell-path)
- [mimixbox crond](#mimixbox-crond) — 2 scenarios
  - [requires foreground mode](#scenario-requires-foreground-mode-1)
  - [describes itself with --help](#scenario-describes-itself-with---help-94)
- [mimixbox crontab](#mimixbox-crontab) — 2 scenarios
  - [reports that interactive edit is unsupported](#scenario-reports-that-interactive-edit-is-unsupported)
  - [describes itself with --help](#scenario-describes-itself-with---help-95)
- [mimixbox cryptpw](#mimixbox-cryptpw) — 2 scenarios
  - [hashes a stdin password with sha-512](#scenario-hashes-a-stdin-password-with-sha-512)
  - [supports the md5 method](#scenario-supports-the-md5-method)
- [mimixbox delgroup](#mimixbox-delgroup) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-96)
- [mimixbox deluser](#mimixbox-deluser) — 2 scenarios
  - [requires a user name](#scenario-requires-a-user-name-1)
  - [describes itself with --help](#scenario-describes-itself-with---help-97)
- [mimixbox getty](#mimixbox-getty) — 2 scenarios
  - [prints the login prompt](#scenario-prints-the-login-prompt)
  - [requires a TTY argument](#scenario-requires-a-tty-argument)
- [mimixbox addgroup / delgroup](#mimixbox-addgroup--delgroup) — 2 scenarios
  - [addgroup requires a group name](#scenario-addgroup-requires-a-group-name)
  - [delgroup requires a group name](#scenario-delgroup-requires-a-group-name)
- [mimixbox loginutils --help helpers](#mimixbox-loginutils---help-helpers) — 6 scenarios
  - [addgroup --help is structured](#scenario-addgroup---help-is-structured)
  - [delgroup --help is structured](#scenario-delgroup---help-is-structured)
  - [linuxrc --help is structured](#scenario-linuxrc---help-is-structured)
  - [run-init --help is structured](#scenario-run-init---help-is-structured)
  - [run-parts --help is structured](#scenario-run-parts---help-is-structured)
  - [start-stop-daemon --help is structured](#scenario-start-stop-daemon---help-is-structured)
- [mimixbox init](#mimixbox-init) — 2 scenarios
  - [runs the inittab sysinit and wait actions](#scenario-runs-the-inittab-sysinit-and-wait-actions)
  - [fails on a missing inittab](#scenario-fails-on-a-missing-inittab)
- [mimixbox login](#mimixbox-login) — 2 scenarios
  - [fails for an unknown user](#scenario-fails-for-an-unknown-user)
  - [describes itself with --help](#scenario-describes-itself-with---help-98)
- [mimixbox mkpasswd](#mimixbox-mkpasswd) — 2 scenarios
  - [hashes with sha-512 and a fixed salt](#scenario-hashes-with-sha-512-and-a-fixed-salt)
  - [reads the password from stdin](#scenario-reads-the-password-from-stdin)
- [mimixbox nologin](#mimixbox-nologin) — 2 scenarios
  - [prints a refusal and exits non-zero](#scenario-prints-a-refusal-and-exits-non-zero)
  - [never runs a passed command](#scenario-never-runs-a-passed-command)
- [mimixbox passwd](#mimixbox-passwd) — 2 scenarios
  - [rejects conflicting flags](#scenario-rejects-conflicting-flags)
  - [describes itself with --help](#scenario-describes-itself-with---help-99)
- [mimixbox run-init](#mimixbox-run-init) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-100)
- [mimixbox run-parts](#mimixbox-run-parts) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-101)
- [mimixbox run-init](#mimixbox-run-init-1) — 2 scenarios
  - [requires NEW_ROOT and INIT](#scenario-requires-new_root-and-init)
  - [rejects a non-directory NEW_ROOT](#scenario-rejects-a-non-directory-new_root)
- [mimixbox run-parts](#mimixbox-run-parts-1) — 2 scenarios
  - [runs executables in alphabetical order](#scenario-runs-executables-in-alphabetical-order)
  - [requires a directory](#scenario-requires-a-directory)
- [mimixbox runlevel](#mimixbox-runlevel) — 1 scenario
  - [reports a runlevel or unknown](#scenario-reports-a-runlevel-or-unknown)
- [mimixbox start-stop-daemon](#mimixbox-start-stop-daemon) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-102)
- [mimixbox start-stop-daemon](#mimixbox-start-stop-daemon-1) — 2 scenarios
  - [starts and stops a background program](#scenario-starts-and-stops-a-background-program)
  - [requires a start or stop mode](#scenario-requires-a-start-or-stop-mode)
- [mimixbox su](#mimixbox-su) — 2 scenarios
  - [fails for an unknown user](#scenario-fails-for-an-unknown-user-1)
  - [describes itself with --help](#scenario-describes-itself-with---help-103)
- [mimixbox sulogin](#mimixbox-sulogin) — 2 scenarios
  - [rejects a wrong root password](#scenario-rejects-a-wrong-root-password)
  - [describes itself with --help](#scenario-describes-itself-with---help-104)
- [mimixbox vlock](#mimixbox-vlock) — 2 scenarios
  - [fails on a wrong password](#scenario-fails-on-a-wrong-password)
  - [describes itself with --help](#scenario-describes-itself-with---help-105)
- [mimixbox mailutils commands expose a dedicated --help helper](#mimixbox-mailutils-commands-expose-a-dedicated---help-helper) — 4 scenarios
  - [makemime --help is structured](#scenario-makemime---help-is-structured)
  - [popmaildir --help is structured](#scenario-popmaildir---help-is-structured)
  - [reformime --help is structured](#scenario-reformime---help-is-structured)
  - [sendmail --help is structured](#scenario-sendmail---help-is-structured)
- [mimixbox makemime](#mimixbox-makemime) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-106)
- [mimixbox popmaildir](#mimixbox-popmaildir) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-107)
- [mimixbox reformime](#mimixbox-reformime) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-108)
- [mimixbox sendmail](#mimixbox-sendmail) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-109)
- [mimixbox arp](#mimixbox-arp) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-110)
- [mimixbox arping](#mimixbox-arping) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-111)
- [mimixbox brctl](#mimixbox-brctl) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-112)
- [mimixbox dhcprelay](#mimixbox-dhcprelay) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-113)
- [mimixbox dnsd](#mimixbox-dnsd) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-114)
- [mimixbox dnsdomainname](#mimixbox-dnsdomainname) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-115)
- [mimixbox dumpleases](#mimixbox-dumpleases) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-116)
- [mimixbox ether-wake](#mimixbox-ether-wake) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-117)
- [mimixbox fakeidentd](#mimixbox-fakeidentd) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-118)
- [mimixbox ftpd](#mimixbox-ftpd) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-119)
- [mimixbox ftpget](#mimixbox-ftpget) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-120)
- [mimixbox ftpput](#mimixbox-ftpput) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-121)
- [mimixbox netutils --help helpers](#mimixbox-netutils---help-helpers) — 55 scenarios
  - [arp --help is structured](#scenario-arp---help-is-structured)
  - [arping --help is structured](#scenario-arping---help-is-structured)
  - [brctl --help is structured](#scenario-brctl---help-is-structured)
  - [dhcprelay --help is structured](#scenario-dhcprelay---help-is-structured)
  - [dnsd --help is structured](#scenario-dnsd---help-is-structured)
  - [dnsdomainname --help is structured](#scenario-dnsdomainname---help-is-structured)
  - [dumpleases --help is structured](#scenario-dumpleases---help-is-structured)
  - [ether-wake --help is structured](#scenario-ether-wake---help-is-structured)
  - [fakeidentd --help is structured](#scenario-fakeidentd---help-is-structured)
  - [ftpd --help is structured](#scenario-ftpd---help-is-structured)
  - [ftpget --help is structured](#scenario-ftpget---help-is-structured)
  - [ftpput --help is structured](#scenario-ftpput---help-is-structured)
  - [http-status-code --help is structured](#scenario-http-status-code---help-is-structured)
  - [httpd --help is structured](#scenario-httpd---help-is-structured)
  - [ifconfig --help is structured](#scenario-ifconfig---help-is-structured)
  - [ifdown --help is structured](#scenario-ifdown---help-is-structured)
  - [ifenslave --help is structured](#scenario-ifenslave---help-is-structured)
  - [ifplugd --help is structured](#scenario-ifplugd---help-is-structured)
  - [ifup --help is structured](#scenario-ifup---help-is-structured)
  - [inetd --help is structured](#scenario-inetd---help-is-structured)
  - [ip --help is structured](#scenario-ip---help-is-structured)
  - [ipaddr --help is structured](#scenario-ipaddr---help-is-structured)
  - [iplink --help is structured](#scenario-iplink---help-is-structured)
  - [ipneigh --help is structured](#scenario-ipneigh---help-is-structured)
  - [iproute --help is structured](#scenario-iproute---help-is-structured)
  - [iprule --help is structured](#scenario-iprule---help-is-structured)
  - [iptunnel --help is structured](#scenario-iptunnel---help-is-structured)
  - [nameif --help is structured](#scenario-nameif---help-is-structured)
  - [nbd-client --help is structured](#scenario-nbd-client---help-is-structured)
  - [netcat --help is structured](#scenario-netcat---help-is-structured)
  - [netstat --help is structured](#scenario-netstat---help-is-structured)
  - [nslookup --help is structured](#scenario-nslookup---help-is-structured)
  - [ntpd --help is structured](#scenario-ntpd---help-is-structured)
  - [ping6 --help is structured](#scenario-ping6---help-is-structured)
  - [pscan --help is structured](#scenario-pscan---help-is-structured)
  - [route --help is structured](#scenario-route---help-is-structured)
  - [slattach --help is structured](#scenario-slattach---help-is-structured)
  - [ssl_client --help is structured](#scenario-ssl_client---help-is-structured)
  - [ssl_server --help is structured](#scenario-ssl_server---help-is-structured)
  - [tc --help is structured](#scenario-tc---help-is-structured)
  - [tcpsvd --help is structured](#scenario-tcpsvd---help-is-structured)
  - [telnet --help is structured](#scenario-telnet---help-is-structured)
  - [telnetd --help is structured](#scenario-telnetd---help-is-structured)
  - [tftp --help is structured](#scenario-tftp---help-is-structured)
  - [tftpd --help is structured](#scenario-tftpd---help-is-structured)
  - [traceroute --help is structured](#scenario-traceroute---help-is-structured)
  - [traceroute6 --help is structured](#scenario-traceroute6---help-is-structured)
  - [tunctl --help is structured](#scenario-tunctl---help-is-structured)
  - [udhcpc --help is structured](#scenario-udhcpc---help-is-structured)
  - [udhcpc6 --help is structured](#scenario-udhcpc6---help-is-structured)
  - [udhcpd --help is structured](#scenario-udhcpd---help-is-structured)
  - [udpsvd --help is structured](#scenario-udpsvd---help-is-structured)
  - [vconfig --help is structured](#scenario-vconfig---help-is-structured)
  - [whois --help is structured](#scenario-whois---help-is-structured)
  - [zcip --help is structured](#scenario-zcip---help-is-structured)
- [mimixbox http-status-code](#mimixbox-http-status-code) — 2 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-122)
  - [looks up a status code by number](#scenario-looks-up-a-status-code-by-number)
- [mimixbox httpd](#mimixbox-httpd) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-123)
- [mimixbox http-status-code](#mimixbox-http-status-code-1) — 1 scenario
  - [explains a status code](#scenario-explains-a-status-code)
- [mimixbox ifconfig](#mimixbox-ifconfig) — 2 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-124)
  - [documents its purpose in --help](#scenario-documents-its-purpose-in---help-9)
- [mimixbox ifdown](#mimixbox-ifdown) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-125)
- [mimixbox ifenslave](#mimixbox-ifenslave) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-126)
- [mimixbox ifplugd](#mimixbox-ifplugd) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-127)
- [mimixbox inetd](#mimixbox-inetd) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-128)
- [mimixbox ipcalc](#mimixbox-ipcalc) — 2 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-129)
  - [documents its purpose in --help](#scenario-documents-its-purpose-in---help-10)
- [mimixbox nc loopback](#mimixbox-nc-loopback) — 1 scenario
  - [transfers data over TCP](#scenario-transfers-data-over-tcp)
- [mimixbox netcat](#mimixbox-netcat) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-130)
- [mimixbox netstat](#mimixbox-netstat) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-131)
- [mimixbox nslookup](#mimixbox-nslookup) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-132)
- [mimixbox ntpd](#mimixbox-ntpd) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-133)
- [mimixbox ping usage](#mimixbox-ping-usage) — 1 scenario
  - [reports an error when no host is given](#scenario-reports-an-error-when-no-host-is-given)
- [mimixbox pscan](#mimixbox-pscan) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-134)
- [mimixbox route](#mimixbox-route) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-135)
- [mimixbox slattach](#mimixbox-slattach) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-136)
- [mimixbox ssl_client](#mimixbox-ssl_client) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-137)
- [mimixbox ssl_server](#mimixbox-ssl_server) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-138)
- [mimixbox tc](#mimixbox-tc) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-139)
- [mimixbox tcpsvd](#mimixbox-tcpsvd) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-140)
- [mimixbox telnet](#mimixbox-telnet) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-141)
- [mimixbox telnetd](#mimixbox-telnetd) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-142)
- [mimixbox tftp](#mimixbox-tftp) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-143)
- [mimixbox tftpd](#mimixbox-tftpd) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-144)
- [mimixbox traceroute](#mimixbox-traceroute) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-145)
- [mimixbox traceroute6](#mimixbox-traceroute6) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-146)
- [mimixbox tunctl](#mimixbox-tunctl) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-147)
- [mimixbox udhcpc](#mimixbox-udhcpc) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-148)
- [mimixbox udhcpc6](#mimixbox-udhcpc6) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-149)
- [mimixbox udhcpd](#mimixbox-udhcpd) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-150)
- [mimixbox udpsvd](#mimixbox-udpsvd) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-151)
- [mimixbox vconfig](#mimixbox-vconfig) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-152)
- [mimixbox whois](#mimixbox-whois) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-153)
- [mimixbox whris usage](#mimixbox-whris-usage) — 1 scenario
  - [reports an error when no domain is given](#scenario-reports-an-error-when-no-domain-is-given)
- [mimixbox zcip](#mimixbox-zcip) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-154)
- [mimixbox halt](#mimixbox-halt) — 4 scenarios
  - [halt --help prints usage and lists the options](#scenario-halt---help-prints-usage-and-lists-the-options)
  - [poweroff --help prints usage](#scenario-poweroff---help-prints-usage)
  - [reboot --help prints usage](#scenario-reboot---help-prints-usage)
  - [halt --version prints the version](#scenario-halt---version-prints-the-version)
- [mimixbox pmutils --help contract](#mimixbox-pmutils---help-contract) — 2 scenarios
  - [poweroff --help is structured](#scenario-poweroff---help-is-structured)
  - [reboot --help is structured](#scenario-reboot---help-is-structured)
- [mimixbox printutils commands expose a dedicated --help helper](#mimixbox-printutils-commands-expose-a-dedicated---help-helper) — 3 scenarios
  - [lpd --help is structured](#scenario-lpd---help-is-structured)
  - [lpq --help is structured](#scenario-lpq---help-is-structured)
  - [lpr --help is structured](#scenario-lpr---help-is-structured)
- [mimixbox depmod](#mimixbox-depmod) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-155)
- [mimixbox fuser](#mimixbox-fuser) — 2 scenarios
  - [finds processes using the current directory](#scenario-finds-processes-using-the-current-directory)
  - [exits non-zero when nothing uses the file](#scenario-exits-non-zero-when-nothing-uses-the-file)
- [mimixbox procps --help contract](#mimixbox-procps---help-contract) — 9 scenarios
  - [depmod --help is structured](#scenario-depmod---help-is-structured)
  - [insmod --help is structured](#scenario-insmod---help-is-structured)
  - [lsmod --help is structured](#scenario-lsmod---help-is-structured)
  - [modinfo --help is structured](#scenario-modinfo---help-is-structured)
  - [modprobe --help is structured](#scenario-modprobe---help-is-structured)
  - [pkill --help is structured](#scenario-pkill---help-is-structured)
  - [pwdx --help is structured](#scenario-pwdx---help-is-structured)
  - [rmmod --help is structured](#scenario-rmmod---help-is-structured)
  - [uptime --help is structured](#scenario-uptime---help-is-structured)
- [mimixbox iostat](#mimixbox-iostat) — 2 scenarios
  - [prints the avg-cpu header](#scenario-prints-the-avg-cpu-header)
  - [prints the device table header](#scenario-prints-the-device-table-header)
- [mimixbox killall5](#mimixbox-killall5) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-156)
- [mimixbox klogd](#mimixbox-klogd) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-157)
- [mimixbox logger](#mimixbox-logger) — 1 scenario
  - [rejects an unknown facility](#scenario-rejects-an-unknown-facility)
- [mimixbox logread](#mimixbox-logread) — 2 scenarios
  - [prints a given log file](#scenario-prints-a-given-log-file)
  - [fails when no readable log is found](#scenario-fails-when-no-readable-log-is-found)
- [mimixbox lsmod](#mimixbox-lsmod) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-158)
- [mimixbox lsof](#mimixbox-lsof) — 2 scenarios
  - [lists the working directory of a process](#scenario-lists-the-working-directory-of-a-process)
  - [prints the column header](#scenario-prints-the-column-header)
- [mimixbox minips](#mimixbox-minips) — 2 scenarios
  - [prints the PID/USER/COMMAND header](#scenario-prints-the-pidusercommand-header)
  - [lists processes](#scenario-lists-processes)
- [mimixbox modinfo](#mimixbox-modinfo) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-159)
- [mimixbox mpstat](#mimixbox-mpstat) — 2 scenarios
  - [prints the CPU column header](#scenario-prints-the-cpu-column-header)
  - [prints the aggregate all row](#scenario-prints-the-aggregate-all-row)
- [mimixbox nmeter](#mimixbox-nmeter) — 2 scenarios
  - [expands a literal percent and copies text](#scenario-expands-a-literal-percent-and-copies-text)
  - [expands the total-memory directive](#scenario-expands-the-total-memory-directive)
- [mimixbox pgrep / pkill](#mimixbox-pgrep--pkill) — 2 scenarios
  - [finds a running process by name](#scenario-finds-a-running-process-by-name)
  - [exits non-zero when nothing matches](#scenario-exits-non-zero-when-nothing-matches)
- [mimixbox pmap](#mimixbox-pmap) — 2 scenarios
  - [prints a total line for a process map](#scenario-prints-a-total-line-for-a-process-map)
  - [rejects an invalid PID](#scenario-rejects-an-invalid-pid)
- [mimixbox powertop](#mimixbox-powertop) — 2 scenarios
  - [runs and exits zero](#scenario-runs-and-exits-zero)
  - [describes itself with --help](#scenario-describes-itself-with---help-160)
- [mimixbox ps](#mimixbox-ps) — 2 scenarios
  - [prints the standard header](#scenario-prints-the-standard-header)
  - [lists running processes](#scenario-lists-running-processes)
- [mimixbox pstree](#mimixbox-pstree) — 1 scenario
  - [builds a tree containing PID 1](#scenario-builds-a-tree-containing-pid-1)
- [mimixbox pwdx](#mimixbox-pwdx) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-161)
- [mimixbox rmmod](#mimixbox-rmmod) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-162)
- [mimixbox smemcap](#mimixbox-smemcap) — 1 scenario
  - [captures a tar containing meminfo](#scenario-captures-a-tar-containing-meminfo)
- [mimixbox sysctl](#mimixbox-sysctl) — 2 scenarios
  - [reads a kernel parameter](#scenario-reads-a-kernel-parameter)
  - [lists parameters with -a](#scenario-lists-parameters-with--a)
- [mimixbox syslogd](#mimixbox-syslogd) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-163)
- [mimixbox top](#mimixbox-top) — 2 scenarios
  - [prints the top summary line](#scenario-prints-the-top-summary-line)
  - [prints the tasks line](#scenario-prints-the-tasks-line)
- [mimixbox uptime](#mimixbox-uptime) — 2 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-164)
  - [prints an uptime/load line](#scenario-prints-an-uptimeload-line)
- [mimixbox uptime / pwdx](#mimixbox-uptime--pwdx) — 2 scenarios
  - [uptime shows the load averages](#scenario-uptime-shows-the-load-averages)
  - [pwdx prints a process working directory](#scenario-pwdx-prints-a-process-working-directory)
- [mimixbox vmstat](#mimixbox-vmstat) — 2 scenarios
  - [prints the column header](#scenario-prints-the-column-header-1)
  - [prints a numeric data row](#scenario-prints-a-numeric-data-row)
- [mimixbox chpst](#mimixbox-chpst) — 2 scenarios
  - [loads an environment directory](#scenario-loads-an-environment-directory)
  - [requires a program](#scenario-requires-a-program)
- [mimixbox envdir](#mimixbox-envdir) — 2 scenarios
  - [sets a variable from a directory file](#scenario-sets-a-variable-from-a-directory-file)
  - [requires a directory and a program](#scenario-requires-a-directory-and-a-program)
- [mimixbox envuidgid](#mimixbox-envuidgid) — 2 scenarios
  - [exports root uid and gid](#scenario-exports-root-uid-and-gid)
  - [fails for an unknown user](#scenario-fails-for-an-unknown-user-2)
- [mimixbox runsv](#mimixbox-runsv) — 2 scenarios
  - [requires a service directory](#scenario-requires-a-service-directory)
  - [describes itself with --help](#scenario-describes-itself-with---help-165)
- [mimixbox runsvdir](#mimixbox-runsvdir) — 2 scenarios
  - [requires a services directory](#scenario-requires-a-services-directory)
  - [describes itself with --help](#scenario-describes-itself-with---help-166)
- [mimixbox setuidgid](#mimixbox-setuidgid) — 2 scenarios
  - [fails for an unknown user](#scenario-fails-for-an-unknown-user-3)
  - [requires a program](#scenario-requires-a-program-1)
- [mimixbox softlimit](#mimixbox-softlimit) — 2 scenarios
  - [runs a program under the limits](#scenario-runs-a-program-under-the-limits)
  - [requires a program](#scenario-requires-a-program-2)
- [mimixbox sv](#mimixbox-sv) — 2 scenarios
  - [writes the up control character](#scenario-writes-the-up-control-character)
  - [reports a running service](#scenario-reports-a-running-service)
- [mimixbox svc](#mimixbox-svc) — 2 scenarios
  - [writes the down control character](#scenario-writes-the-down-control-character)
  - [requires a control command](#scenario-requires-a-control-command)
- [mimixbox svlogd](#mimixbox-svlogd) — 2 scenarios
  - [appends stdin to the current log](#scenario-appends-stdin-to-the-current-log)
  - [requires a directory](#scenario-requires-a-directory-1)
- [mimixbox svok](#mimixbox-svok) — 2 scenarios
  - [succeeds for a supervised service](#scenario-succeeds-for-a-supervised-service)
  - [returns 100 for an unsupervised service](#scenario-returns-100-for-an-unsupervised-service)
- [mimixbox chcon](#mimixbox-chcon) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-167)
- [mimixbox getenforce](#mimixbox-getenforce) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-168)
- [mimixbox getsebool](#mimixbox-getsebool) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-169)
- [mimixbox securityutils --help contract](#mimixbox-securityutils---help-contract) — 13 scenarios
  - [chcon --help is structured](#scenario-chcon---help-is-structured)
  - [getenforce --help is structured](#scenario-getenforce---help-is-structured)
  - [getsebool --help is structured](#scenario-getsebool---help-is-structured)
  - [load_policy --help is structured](#scenario-load_policy---help-is-structured)
  - [matchpathcon --help is structured](#scenario-matchpathcon---help-is-structured)
  - [restorecon --help is structured](#scenario-restorecon---help-is-structured)
  - [runcon --help is structured](#scenario-runcon---help-is-structured)
  - [selinuxenabled --help is structured](#scenario-selinuxenabled---help-is-structured)
  - [sestatus --help is structured](#scenario-sestatus---help-is-structured)
  - [setenforce --help is structured](#scenario-setenforce---help-is-structured)
  - [setfiles --help is structured](#scenario-setfiles---help-is-structured)
  - [setsebool --help is structured](#scenario-setsebool---help-is-structured)
  - [zip-pwcrack --help is structured](#scenario-zip-pwcrack---help-is-structured)
- [mimixbox pwcrack](#mimixbox-pwcrack) — 1 scenario
  - [finds a weak password in the wordlist](#scenario-finds-a-weak-password-in-the-wordlist)
- [mimixbox pwgen](#mimixbox-pwgen) — 1 scenario
  - [generates the requested number of passwords](#scenario-generates-the-requested-number-of-passwords)
- [mimixbox pwscore](#mimixbox-pwscore) — 1 scenario
  - [scores a common password as zero](#scenario-scores-a-common-password-as-zero)
- [mimixbox runcon](#mimixbox-runcon) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-170)
- [mimixbox selinuxenabled](#mimixbox-selinuxenabled) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-171)
- [mimixbox sestatus](#mimixbox-sestatus) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-172)
- [mimixbox setenforce](#mimixbox-setenforce) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-173)
- [mimixbox setfiles](#mimixbox-setfiles) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-174)
- [mimixbox setsebool](#mimixbox-setsebool) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-175)
- [mimixbox unshadow](#mimixbox-unshadow) — 1 scenario
  - [merges the shadow hash into the passwd line](#scenario-merges-the-shadow-hash-into-the-passwd-line)
- [mimixbox zip-pwcrack](#mimixbox-zip-pwcrack) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-176)
- [mimixbox zip-pwcrack](#mimixbox-zip-pwcrack-1) — 1 scenario
  - [recovers the ZIP password from the wordlist](#scenario-recovers-the-zip-password-from-the-wordlist)
- [mimixbox arch](#mimixbox-arch) — 1 scenario
  - [prints a non-empty machine name](#scenario-prints-a-non-empty-machine-name)
- [mimixbox base64](#mimixbox-base64) — 5 scenarios
  - [encodes standard input](#scenario-encodes-standard-input)
  - [encodes the file contents](#scenario-encodes-the-file-contents)
  - [decodes standard input](#scenario-decodes-standard-input)
  - [returns the original text (round trip)](#scenario-returns-the-original-text-round-trip)
  - [reports an error for a non-existent file](#scenario-reports-an-error-for-a-non-existent-file)
- [mimixbox basename](#mimixbox-basename) — 12 scenarios
  - [show test.txt](#scenario-show-testtxt)
  - [show test](#scenario-show-test)
  - [show .test](#scenario-show-test-1)
  - [show nao for a trailing slash](#scenario-show-nao-for-a-trailing-slash)
  - [show error without operand](#scenario-show-error-without-operand)
  - [show / for root](#scenario-show--for-root)
  - [show empty string](#scenario-show-empty-string)
  - [show error for extra operand](#scenario-show-error-for-extra-operand)
  - [show three basenames with -a](#scenario-show-three-basenames-with--a)
  - [show three basenames joined with -a -z](#scenario-show-three-basenames-joined-with--a--z)
  - [show basename without the suffix](#scenario-show-basename-without-the-suffix)
  - [show basename built from an environment variable](#scenario-show-basename-built-from-an-environment-variable)
- [mimixbox bc](#mimixbox-bc) — 4 scenarios
  - [respects operator precedence](#scenario-respects-operator-precedence)
  - [honors scale for division](#scenario-honors-scale-for-division)
  - [supports variables](#scenario-supports-variables)
  - [computes powers](#scenario-computes-powers)
- [mimixbox cal](#mimixbox-cal) — 1 scenario
  - [prints the month calendar](#scenario-prints-the-month-calendar)
- [mimixbox chmod](#mimixbox-chmod) — 3 scenarios
  - [sets the permission bits with an octal mode](#scenario-sets-the-permission-bits-with-an-octal-mode)
  - [adds owner execute to mode 600 with a symbolic mode](#scenario-adds-owner-execute-to-mode-600-with-a-symbolic-mode)
  - [reports an error on a missing file](#scenario-reports-an-error-on-a-missing-file)
- [mimixbox chroot](#mimixbox-chroot) — 3 scenarios
  - [prints usage with --help and exits 0](#scenario-prints-usage-with---help-and-exits-0-11)
  - [documents the --userspec identity option in --help](#scenario-documents-the---userspec-identity-option-in---help)
  - [fails with a message when given no operand](#scenario-fails-with-a-message-when-given-no-operand-5)
- [mimixbox cmp](#mimixbox-cmp) — 3 scenarios
  - [prints nothing and succeeds on identical files](#scenario-prints-nothing-and-succeeds-on-identical-files)
  - [reports the first differing byte and line on differing files](#scenario-reports-the-first-differing-byte-and-line-on-differing-files)
  - [cmp -s prints nothing but exits non-zero](#scenario-cmp--s-prints-nothing-but-exits-non-zero)
- [mimixbox cmp (GNU options)](#mimixbox-cmp-gnu-options) — 5 scenarios
  - [-n reports equality when the difference is past the byte limit](#scenario--n-reports-equality-when-the-difference-is-past-the-byte-limit)
  - [--bytes reports the difference within the byte limit](#scenario---bytes-reports-the-difference-within-the-byte-limit)
  - [-i skips the first N bytes of both files](#scenario--i-skips-the-first-n-bytes-of-both-files)
  - [-i N:M skips N bytes of file1 and M of file2](#scenario--i-nm-skips-n-bytes-of-file1-and-m-of-file2)
  - [-b prints the differing byte values in the message](#scenario--b-prints-the-differing-byte-values-in-the-message)
- [mimixbox cp (permission preservation)](#mimixbox-cp-permission-preservation) — 3 scenarios
  - [keeps the source file mode (execute bit)](#scenario-keeps-the-source-file-mode-execute-bit)
  - [keeps a private directory mode](#scenario-keeps-a-private-directory-mode)
  - [overwrites a read-only destination with -f](#scenario-overwrites-a-read-only-destination-with--f)
- [mimixbox cut](#mimixbox-cut) — 5 scenarios
  - [prints the chosen field](#scenario-prints-the-chosen-field)
  - [prints the chosen fields joined by the delimiter](#scenario-prints-the-chosen-fields-joined-by-the-delimiter)
  - [prints from the field to the end](#scenario-prints-from-the-field-to-the-end)
  - [prints the chosen character range](#scenario-prints-the-chosen-character-range)
  - [reports an error without a list](#scenario-reports-an-error-without-a-list)
- [mimixbox cut (GNU options)](#mimixbox-cut-gnu-options) — 4 scenarios
  - [--complement keeps the fields not selected](#scenario---complement-keeps-the-fields-not-selected)
  - [--complement keeps the bytes not selected](#scenario---complement-keeps-the-bytes-not-selected)
  - [-z splits and joins records on NUL (fields)](#scenario--z-splits-and-joins-records-on-nul-fields)
  - [-z cuts bytes from each NUL-delimited record](#scenario--z-cuts-bytes-from-each-nul-delimited-record)
- [mimixbox date](#mimixbox-date) — 4 scenarios
  - [formats the date portion of an epoch](#scenario-formats-the-date-portion-of-an-epoch)
  - [formats the time portion of an epoch](#scenario-formats-the-time-portion-of-an-epoch)
  - [prints a literal percent sign](#scenario-prints-a-literal-percent-sign)
  - [prints a four-digit year](#scenario-prints-a-four-digit-year)
- [mimixbox dc](#mimixbox-dc) — 4 scenarios
  - [performs integer division](#scenario-performs-integer-division)
  - [honors the precision register](#scenario-honors-the-precision-register)
  - [evaluates -e expressions](#scenario-evaluates--e-expressions)
  - [stores and loads registers](#scenario-stores-and-loads-registers)
- [mimixbox dd](#mimixbox-dd) — 3 scenarios
  - [reproduces the input (stdin to stdout)](#scenario-reproduces-the-input-stdin-to-stdout)
  - [copies only the requested blocks with count](#scenario-copies-only-the-requested-blocks-with-count)
  - [conv=ucase upper-cases the data](#scenario-convucase-upper-cases-the-data)
- [mimixbox df](#mimixbox-df) — 2 scenarios
  - [shows the column header](#scenario-shows-the-column-header)
  - [exits zero for the current directory](#scenario-exits-zero-for-the-current-directory)
- [mimixbox df GNU flags](#mimixbox-df-gnu-flags) — 10 scenarios
  - [--output prints the selected column headers in order](#scenario---output-prints-the-selected-column-headers-in-order)
  - [--output honors a reordered field list](#scenario---output-honors-a-reordered-field-list)
  - [--output rejects an unknown field](#scenario---output-rejects-an-unknown-field)
  - [--total emits a row labeled total](#scenario---total-emits-a-row-labeled-total)
  - [--total works with the classic layout too](#scenario---total-works-with-the-classic-layout-too)
  - [--type accepts a type filter and exits cleanly](#scenario---type-accepts-a-type-filter-and-exits-cleanly)
  - [--type is repeatable](#scenario---type-is-repeatable)
  - [--block-size labels the block-size in the classic header](#scenario---block-size-labels-the-block-size-in-the-classic-header)
  - [--block-size rejects an invalid size](#scenario---block-size-rejects-an-invalid-size)
  - [--all lists at least as many rows with -a as without](#scenario---all-lists-at-least-as-many-rows-with--a-as-without)
- [mimixbox dirname](#mimixbox-dirname) — 9 scenarios
  - [print /home/nao for an absolute file path](#scenario-print-homenao-for-an-absolute-file-path)
  - [print /home/nao for a filename without extension](#scenario-print-homenao-for-a-filename-without-extension)
  - [print /home/nao for a hidden file](#scenario-print-homenao-for-a-hidden-file)
  - [print error without operand](#scenario-print-error-without-operand-1)
  - [print / for the root directory](#scenario-print--for-the-root-directory)
  - [print . for an empty string](#scenario-print--for-an-empty-string)
  - [print /bin /home / with line feed for three arguments](#scenario-print-bin-home--with-line-feed-for-three-arguments)
  - [print NUL-separated dirnames for three arguments with -z](#scenario-print-nul-separated-dirnames-for-three-arguments-with--z)
  - [print /aaa/bbb/ccc built from an environment variable](#scenario-print-aaabbbccc-built-from-an-environment-variable)
- [mimixbox du](#mimixbox-du) — 2 scenarios
  - [-b reports the total apparent byte size](#scenario--b-reports-the-total-apparent-byte-size)
  - [-s reports the total in 1K blocks](#scenario--s-reports-the-total-in-1k-blocks)
- [mimixbox du GNU flags](#mimixbox-du-gnu-flags) — 7 scenarios
  - [omits directories deeper than --max-depth](#scenario-omits-directories-deeper-than---max-depth)
  - [prints only the operand total with --max-depth=0](#scenario-prints-only-the-operand-total-with---max-depth0)
  - [skips entries matching --exclude](#scenario-skips-entries-matching---exclude)
  - [skips glob-matching files under -a](#scenario-skips-glob-matching-files-under--a)
  - [reports exact bytes with --apparent-size](#scenario-reports-exact-bytes-with---apparent-size)
  - [reports block counts by default](#scenario-reports-block-counts-by-default)
  - [matches a plain run on a single filesystem with -x](#scenario-matches-a-plain-run-on-a-single-filesystem-with--x)
- [mimixbox echo](#mimixbox-echo) — 9 scenarios
  - [says Hello World!](#scenario-says-hello-world)
  - [says Hello World! (helper ignores the positional argument)](#scenario-says-hello-world-helper-ignores-the-positional-argument)
  - [expands an environment variable](#scenario-expands-an-environment-variable)
  - [pipes data through xargs](#scenario-pipes-data-through-xargs)
  - [says nothing with no arguments](#scenario-says-nothing-with-no-arguments)
  - [redirects data to a file and shows it](#scenario-redirects-data-to-a-file-and-shows-it)
  - [--help as the first argument prints usage instead of the literal text](#scenario---help-as-the-first-argument-prints-usage-instead-of-the-literal-text)
  - [--version as the first argument prints the version line](#scenario---version-as-the-first-argument-prints-the-version-line)
  - [--help that is not the first argument stays literal](#scenario---help-that-is-not-the-first-argument-stays-literal)
- [mimixbox env](#mimixbox-env) — 3 scenarios
  - [adds the assignment to the printed environment](#scenario-adds-the-assignment-to-the-printed-environment)
  - [-i prints only the given assignment](#scenario--i-prints-only-the-given-assignment)
  - [passes the variable to the run command](#scenario-passes-the-variable-to-the-run-command)
- [mimixbox env GNU flags](#mimixbox-env-gnu-flags) — 7 scenarios
  - [--chdir reports the chdir target via pwd (long flag with =)](#scenario---chdir-reports-the-chdir-target-via-pwd-long-flag-with-)
  - [--chdir reports the chdir target via pwd (-C short flag)](#scenario---chdir-reports-the-chdir-target-via-pwd--c-short-flag)
  - [--chdir fails when the directory does not exist](#scenario---chdir-fails-when-the-directory-does-not-exist)
  - [--split-string splits the command and its arguments (-S)](#scenario---split-string-splits-the-command-and-its-arguments--s)
  - [--split-string splits with the long flag and whitespace runs](#scenario---split-string-splits-with-the-long-flag-and-whitespace-runs)
  - [--ignore-signal accepts known names and still runs the command](#scenario---ignore-signal-accepts-known-names-and-still-runs-the-command)
  - [--ignore-signal rejects an unknown signal name](#scenario---ignore-signal-rejects-an-unknown-signal-name)
- [mimixbox expr](#mimixbox-expr) — 5 scenarios
  - [adds two numbers](#scenario-adds-two-numbers)
  - [multiplies two numbers](#scenario-multiplies-two-numbers)
  - [respects parentheses](#scenario-respects-parentheses)
  - [prints the string length](#scenario-prints-the-string-length)
  - [prints 0 and exits non-zero for a zero result](#scenario-prints-0-and-exits-non-zero-for-a-zero-result)
- [mimixbox factor](#mimixbox-factor) — 4 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-177)
  - [documents its purpose in --help](#scenario-documents-its-purpose-in---help-11)
  - [factors a small integer](#scenario-factors-a-small-integer)
  - [fails on a non-numeric operand](#scenario-fails-on-a-non-numeric-operand)
- [mimixbox false](#mimixbox-false) — 1 scenario
  - [prints nothing and exits 1](#scenario-prints-nothing-and-exits-1)
- [mimixbox free](#mimixbox-free) — 1 scenario
  - [prints the column header](#scenario-prints-the-column-header-2)
- [mimixbox fsync](#mimixbox-fsync) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-178)
- [mimixbox ghrdc](#mimixbox-ghrdc) — 2 scenarios
  - [prints usage with --help and exits 0](#scenario-prints-usage-with---help-and-exits-0-12)
  - [fails with a message when given no operand](#scenario-fails-with-a-message-when-given-no-operand-6)
- [mimixbox groups](#mimixbox-groups) — 2 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-179)
  - [prints the groups of a named user](#scenario-prints-the-groups-of-a-named-user)
- [mimixbox gzip](#mimixbox-gzip) — 1 scenario
  - [compresses and decompresses back to the original](#scenario-compresses-and-decompresses-back-to-the-original)
- [mimixbox shellutils --help helpers](#mimixbox-shellutils---help-helpers) — 5 scenarios
  - [fsync --help is structured](#scenario-fsync---help-is-structured)
  - [log-collect --help is structured](#scenario-log-collect---help-is-structured)
  - [sddf --help is structured](#scenario-sddf---help-is-structured)
  - [time --help is structured](#scenario-time---help-is-structured)
  - [usleep --help is structured](#scenario-usleep---help-is-structured)
- [mimixbox hermetic harness](#mimixbox-hermetic-harness) — 3 scenarios
  - [resolves cat to the MimixBox binary, not the host command](#scenario-resolves-cat-to-the-mimixbox-binary-not-the-host-command)
  - [resolves head to the MimixBox binary, not the host command](#scenario-resolves-head-to-the-mimixbox-binary-not-the-host-command)
  - [resolves base64 to the MimixBox binary, not the host command](#scenario-resolves-base64-to-the-mimixbox-binary-not-the-host-command)
- [mimixbox hostid](#mimixbox-hostid) — 2 scenarios
  - [prints 8 hexadecimal digits](#scenario-prints-8-hexadecimal-digits)
  - [prints the same value on repeated calls](#scenario-prints-the-same-value-on-repeated-calls)
- [mimixbox hostname](#mimixbox-hostname) — 1 scenario
  - [prints a non-empty host name](#scenario-prints-a-non-empty-host-name)
- [mimixbox id](#mimixbox-id) — 2 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-180)
  - [prints the current uid/gid line](#scenario-prints-the-current-uidgid-line)
- [mimixbox install](#mimixbox-install) — 4 scenarios
  - [copies the file content](#scenario-copies-the-file-content)
  - [sets the requested mode](#scenario-sets-the-requested-mode)
  - [creates directories with -d](#scenario-creates-directories-with--d)
  - [fails without a destination](#scenario-fails-without-a-destination)
- [mimixbox install_gnu](#mimixbox-install_gnu) — 7 scenarios
  - [makes a simple backup before overwriting with --backup=simple](#scenario-makes-a-simple-backup-before-overwriting-with---backupsimple)
  - [honors --suffix when backing up](#scenario-honors---suffix-when-backing-up)
  - [makes numbered backups with --backup=numbered](#scenario-makes-numbered-backups-with---backupnumbered)
  - [uses a simple backup with --backup=existing when none are numbered](#scenario-uses-a-simple-backup-with---backupexisting-when-none-are-numbered)
  - [attempts chown and fails as non-root with --owner/--group](#scenario-attempts-chown-and-fails-as-non-root-with---owner--group)
  - [rejects an invalid owner name](#scenario-rejects-an-invalid-owner-name)
  - [rejects an invalid --backup control](#scenario-rejects-an-invalid---backup-control)
- [mimixbox kill](#mimixbox-kill) — 1 scenario
  - [lists signal names with -l](#scenario-lists-signal-names-with--l)
- [mimixbox killall](#mimixbox-killall) — 1 scenario
  - [kills a process by name](#scenario-kills-a-process-by-name)
- [mimixbox leadtime](#mimixbox-leadtime) — 6 scenarios
  - [prints usage with --help and exits 0](#scenario-prints-usage-with---help-and-exits-0-13)
  - [fails when no subcommand is given](#scenario-fails-when-no-subcommand-is-given)
  - [fails on an unknown subcommand](#scenario-fails-on-an-unknown-subcommand)
  - [fails with a deterministic error when no token is set](#scenario-fails-with-a-deterministic-error-when-no-token-is-set)
  - [fails when --owner/--repo are missing](#scenario-fails-when---owner--repo-are-missing)
  - [rejects --json with --markdown](#scenario-rejects---json-with---markdown)
- [mimixbox top-level list/suggestion CLI](#mimixbox-top-level-listsuggestion-cli) — 4 scenarios
  - [--list --json emits a JSON array containing cat and ls on stdout](#scenario---list---json-emits-a-json-array-containing-cat-and-ls-on-stdout)
  - [--list --filter=cat includes cat and excludes ls](#scenario---list---filtercat-includes-cat-and-excludes-ls)
  - [--list --subsystem=textutils includes cat and excludes ls](#scenario---list---subsystemtextutils-includes-cat-and-excludes-ls)
  - [an unknown command suggests the nearest applet, error-first](#scenario-an-unknown-command-suggests-the-nearest-applet-error-first)
- [mimixbox log-collect](#mimixbox-log-collect-1) — 1 scenario
  - [copies log files into the output directory](#scenario-copies-log-files-into-the-output-directory)
- [mimixbox logname](#mimixbox-logname) — 1 scenario
  - [prints the login name from LOGNAME](#scenario-prints-the-login-name-from-logname)
- [mimixbox mbsh](#mimixbox-mbsh) — 12 scenarios
  - [runs an external command and shows a cwd-aware prompt](#scenario-runs-an-external-command-and-shows-a-cwd-aware-prompt)
  - [ignores comment lines](#scenario-ignores-comment-lines)
  - [expands $? to the last exit status](#scenario-expands--to-the-last-exit-status)
  - [lets a stdin-consuming command read the remaining script input](#scenario-lets-a-stdin-consuming-command-read-the-remaining-script-input)
  - [does not reparse command-consumed stdin as later commands](#scenario-does-not-reparse-command-consumed-stdin-as-later-commands)
  - [keeps double-quoted spaces in one argument](#scenario-keeps-double-quoted-spaces-in-one-argument)
  - [expands $HOME](#scenario-expands-home)
  - [passes a NAME=value prefix to the command environment](#scenario-passes-a-namevalue-prefix-to-the-command-environment)
  - [runs commands in sequence and redirects output](#scenario-runs-commands-in-sequence-and-redirects-output)
  - [pipes one command into another](#scenario-pipes-one-command-into-another)
  - [redirects input with <](#scenario-redirects-input-with-)
  - [changes directory with cd](#scenario-changes-directory-with-cd)
- [mimixbox top-level CLI](#mimixbox-top-level-cli) — 4 scenarios
  - [prints usage to stdout with --help and exits success](#scenario-prints-usage-to-stdout-with---help-and-exits-success)
  - [lists the applets with --list](#scenario-lists-the-applets-with---list)
  - [rejects an unknown option on stderr without polluting stdout](#scenario-rejects-an-unknown-option-on-stderr-without-polluting-stdout)
  - [installs and removes applet symlinks in a temp directory](#scenario-installs-and-removes-applet-symlinks-in-a-temp-directory)
- [mimixbox mknod](#mimixbox-mknod) — 2 scenarios
  - [creates a FIFO with type p](#scenario-creates-a-fifo-with-type-p)
  - [rejects an invalid device type](#scenario-rejects-an-invalid-device-type)
- [mimixbox nice](#mimixbox-nice) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-181)
- [mimixbox nohup](#mimixbox-nohup) — 1 scenario
  - [runs the command and passes output through](#scenario-runs-the-command-and-passes-output-through)
- [mimixbox nproc](#mimixbox-nproc) — 1 scenario
  - [prints a positive number](#scenario-prints-a-positive-number)
- [mimixbox od](#mimixbox-od) — 3 scenarios
  - [dumps characters with C escapes](#scenario-dumps-characters-with-c-escapes)
  - [dumps hex bytes with hex addresses](#scenario-dumps-hex-bytes-with-hex-addresses)
  - [suppresses the address column](#scenario-suppresses-the-address-column)
- [mimixbox path](#mimixbox-path) — 5 scenarios
  - [prints the base name with --basename](#scenario-prints-the-base-name-with---basename)
  - [prints the directory with --dirname](#scenario-prints-the-directory-with---dirname)
  - [prints the extension with --extension](#scenario-prints-the-extension-with---extension)
  - [prints the cleaned path with --canonical](#scenario-prints-the-cleaned-path-with---canonical)
  - [reports an error with no operand](#scenario-reports-an-error-with-no-operand)
- [mimixbox pidof](#mimixbox-pidof) — 2 scenarios
  - [finds the PID of a running process via MimixBox pidof](#scenario-finds-the-pid-of-a-running-process-via-mimixbox-pidof)
  - [resolves bare pidof to the MimixBox-installed symlink](#scenario-resolves-bare-pidof-to-the-mimixbox-installed-symlink)
- [mimixbox posixer](#mimixbox-posixer) — 1 scenario
  - [prints a table header](#scenario-prints-a-table-header)
- [mimixbox printenv](#mimixbox-printenv) — 1 scenario
  - [prints an environment variable](#scenario-prints-an-environment-variable)
- [mimixbox printf](#mimixbox-printf) — 1 scenario
  - [formats arguments](#scenario-formats-arguments)
- [mimixbox printf_meta](#mimixbox-printf_meta) — 3 scenarios
  - [prints help for a leading --help](#scenario-prints-help-for-a-leading---help)
  - [prints the version banner for a leading --version](#scenario-prints-the-version-banner-for-a-leading---version)
  - [treats a later --help as an ordinary operand](#scenario-treats-a-later---help-as-an-ordinary-operand)
- [mimixbox pwd](#mimixbox-pwd) — 1 scenario
  - [prints the working directory](#scenario-prints-the-working-directory)
- [mimixbox realpath](#mimixbox-realpath) — 3 scenarios
  - [resolves an existing file to its absolute path](#scenario-resolves-an-existing-file-to-its-absolute-path)
  - [prints the cleaned absolute path with -m on a missing path](#scenario-prints-the-cleaned-absolute-path-with--m-on-a-missing-path)
  - [reports an error with no operand](#scenario-reports-an-error-with-no-operand-1)
- [mimixbox realpath_gnu](#mimixbox-realpath_gnu) — 2 scenarios
  - [prints a path relative to --relative-to](#scenario-prints-a-path-relative-to---relative-to)
  - [resolves .. lexically with -L -m](#scenario-resolves--lexically-with--l--m)
- [mimixbox richhelp](#mimixbox-richhelp) — 6 scenarios
  - [cp --help has a description, examples, and exit status](#scenario-cp---help-has-a-description-examples-and-exit-status)
  - [tail --help documents follow mode with examples](#scenario-tail---help-documents-follow-mode-with-examples)
  - [wget --help has examples and compatibility notes](#scenario-wget---help-has-examples-and-compatibility-notes)
  - [mbsh --help describes the shell and its limits](#scenario-mbsh---help-describes-the-shell-and-its-limits)
  - [vi --help lists the supported keys](#scenario-vi---help-lists-the-supported-keys)
  - [find --help has examples and notes](#scenario-find---help-has-examples-and-notes)
- [mimixbox sddf](#mimixbox-sddf) — 2 scenarios
  - [prints usage with --help and exits 0](#scenario-prints-usage-with---help-and-exits-0-14)
  - [fails with a message when given no operand](#scenario-fails-with-a-message-when-given-no-operand-7)
- [mimixbox seq](#mimixbox-seq) — 6 scenarios
  - [counts from 1 to LAST](#scenario-counts-from-1-to-last)
  - [counts from FIRST to LAST](#scenario-counts-from-first-to-last)
  - [counts by INCREMENT](#scenario-counts-by-increment)
  - [joins the numbers with the separator](#scenario-joins-the-numbers-with-the-separator)
  - [pads numbers with leading zeros](#scenario-pads-numbers-with-leading-zeros)
  - [reports an error for an invalid operand](#scenario-reports-an-error-for-an-invalid-operand)
- [mimixbox sleep](#mimixbox-sleep) — 1 scenario
  - [sleeps then returns](#scenario-sleeps-then-returns)
- [mimixbox sort](#mimixbox-sort) — 4 scenarios
  - [sorts lines alphabetically](#scenario-sorts-lines-alphabetically)
  - [sorts by numeric value](#scenario-sorts-by-numeric-value)
  - [reverses the order](#scenario-reverses-the-order)
  - [drops duplicate lines](#scenario-drops-duplicate-lines)
- [mimixbox sort (GNU extensions)](#mimixbox-sort-gnu-extensions) — 8 scenarios
  - [-V orders version numbers by value](#scenario--v-orders-version-numbers-by-value)
  - [-g orders floating-point values including exponents](#scenario--g-orders-floating-point-values-including-exponents)
  - [-h orders human-readable sizes by magnitude](#scenario--h-orders-human-readable-sizes-by-magnitude)
  - [-s keeps input order for equal keys](#scenario--s-keeps-input-order-for-equal-keys)
  - [-z reads and writes NUL-delimited records](#scenario--z-reads-and-writes-nul-delimited-records)
  - [-m merges already-sorted input](#scenario--m-merges-already-sorted-input)
  - [--parallel is accepted without error](#scenario---parallel-is-accepted-without-error)
  - [--temporary-directory is accepted without error](#scenario---temporary-directory-is-accepted-without-error)
- [mimixbox speaker](#mimixbox-speaker) — 1 scenario
  - [errors when no text is given](#scenario-errors-when-no-text-is-given)
- [mimixbox sync](#mimixbox-sync) — 1 scenario
  - [flushes filesystem buffers](#scenario-flushes-filesystem-buffers)
- [mimixbox tee](#mimixbox-tee) — 3 scenarios
  - [echoes standard input to stdout](#scenario-echoes-standard-input-to-stdout)
  - [also writes the input to the file](#scenario-also-writes-the-input-to-the-file)
  - [appends to a file with -a keeping the existing content](#scenario-appends-to-a-file-with--a-keeping-the-existing-content)
- [mimixbox tee --output-error](#mimixbox-tee---output-error) — 3 scenarios
  - [copies input and succeeds with an explicit MODE](#scenario-copies-input-and-succeeds-with-an-explicit-mode)
  - [warn mode still writes the good file but exits nonzero](#scenario-warn-mode-still-writes-the-good-file-but-exits-nonzero)
  - [exit mode does not create the later good file and exits nonzero](#scenario-exit-mode-does-not-create-the-later-good-file-and-exits-nonzero)
- [mimixbox test](#mimixbox-test) — 5 scenarios
  - [string equality is true for equal strings](#scenario-string-equality-is-true-for-equal-strings)
  - [integer comparison is true when 2 > 1](#scenario-integer-comparison-is-true-when-2--1)
  - [integer comparison is false when 1 > 2](#scenario-integer-comparison-is-false-when-1--2)
  - [file existence is true for an existing file](#scenario-file-existence-is-true-for-an-existing-file)
  - [negation negates the expression](#scenario-negation-negates-the-expression)
- [mimixbox test (meta)](#mimixbox-test-meta) — 3 scenarios
  - [prints help for a sole --help](#scenario-prints-help-for-a-sole---help)
  - [prints the version banner for a sole --version](#scenario-prints-the-version-banner-for-a-sole---version)
  - [evaluates an expression when --help is not the sole argument](#scenario-evaluates-an-expression-when---help-is-not-the-sole-argument)
- [mimixbox time](#mimixbox-time) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-182)
- [mimixbox time / fsync](#mimixbox-time--fsync) — 4 scenarios
  - [time runs the command and passes its output through](#scenario-time-runs-the-command-and-passes-its-output-through)
  - [time reports the real elapsed time on stderr](#scenario-time-reports-the-real-elapsed-time-on-stderr)
  - [fsync succeeds on an existing file](#scenario-fsync-succeeds-on-an-existing-file)
  - [fsync fails on a missing file](#scenario-fsync-fails-on-a-missing-file)
- [mimixbox timeout](#mimixbox-timeout) — 2 scenarios
  - [runs the command to completion](#scenario-runs-the-command-to-completion)
  - [returns exit code 124 on timeout](#scenario-returns-exit-code-124-on-timeout)
- [mimixbox tree / nice](#mimixbox-tree--nice) — 3 scenarios
  - [tree counts directories and files in its summary](#scenario-tree-counts-directories-and-files-in-its-summary)
  - [tree exits successfully on a readable directory](#scenario-tree-exits-successfully-on-a-readable-directory)
  - [nice prints a numeric niceness](#scenario-nice-prints-a-numeric-niceness)
- [mimixbox true](#mimixbox-true) — 1 scenario
  - [prints nothing and exits 0](#scenario-prints-nothing-and-exits-0)
- [mimixbox tsort](#mimixbox-tsort) — 3 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-183)
  - [documents its purpose in --help](#scenario-documents-its-purpose-in---help-12)
  - [produces a topological order](#scenario-produces-a-topological-order)
- [mimixbox tty](#mimixbox-tty) — 1 scenario
  - [reports not a tty when stdin is a pipe](#scenario-reports-not-a-tty-when-stdin-is-a-pipe)
- [mimixbox uname](#mimixbox-uname) — 1 scenario
  - [prints the kernel name](#scenario-prints-the-kernel-name)
- [mimixbox uniq](#mimixbox-uniq) — 4 scenarios
  - [collapses repeated adjacent lines](#scenario-collapses-repeated-adjacent-lines)
  - [-c prefixes each line with its count](#scenario--c-prefixes-each-line-with-its-count)
  - [-d prints only repeated lines once](#scenario--d-prints-only-repeated-lines-once)
  - [-u prints only lines that never repeat](#scenario--u-prints-only-lines-that-never-repeat)
- [mimixbox users](#mimixbox-users) — 2 scenarios
  - [runs and exits successfully](#scenario-runs-and-exits-successfully)
  - [treats a missing utmp as nobody logged in](#scenario-treats-a-missing-utmp-as-nobody-logged-in)
- [mimixbox usleep](#mimixbox-usleep) — 2 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-184)
  - [rejects a non-numeric microsecond count](#scenario-rejects-a-non-numeric-microsecond-count)
- [mimixbox uuidgen](#mimixbox-uuidgen) — 1 scenario
  - [prints a well-formed UUID](#scenario-prints-a-well-formed-uuid)
- [mimixbox w](#mimixbox-w) — 2 scenarios
  - [prints a summary header with the load averages](#scenario-prints-a-summary-header-with-the-load-averages)
  - [prints the column header](#scenario-prints-the-column-header-3)
- [mimixbox watch](#mimixbox-watch) — 1 scenario
  - [runs the command and shows its output](#scenario-runs-the-command-and-shows-its-output)
- [mimixbox wget](#mimixbox-wget) — 3 scenarios
  - [prints usage with --help and exits 0](#scenario-prints-usage-with---help-and-exits-0-15)
  - [fails with a message when given no operand](#scenario-fails-with-a-message-when-given-no-operand-8)
  - [documents the added download options](#scenario-documents-the-added-download-options)
- [mimixbox which](#mimixbox-which) — 6 scenarios
  - [prints the MimixBox path](#scenario-prints-the-mimixbox-path)
  - [prints nothing for a binary that does not exist](#scenario-prints-nothing-for-a-binary-that-does-not-exist)
  - [prints paths of three binaries](#scenario-prints-paths-of-three-binaries)
  - [prints paths of two binaries and fails when one of three is missing](#scenario-prints-paths-of-two-binaries-and-fails-when-one-of-three-is-missing)
  - [prints nothing without an operand](#scenario-prints-nothing-without-an-operand)
  - [prints nothing when data comes from a pipe](#scenario-prints-nothing-when-data-comes-from-a-pipe)
- [mimixbox who](#mimixbox-who) — 3 scenarios
  - [prints nothing and succeeds on an empty utmp](#scenario-prints-nothing-and-succeeds-on-an-empty-utmp)
  - [-q reports zero users on an empty utmp](#scenario--q-reports-zero-users-on-an-empty-utmp)
  - [--help prints usage](#scenario---help-prints-usage)
- [mimixbox whoami](#mimixbox-whoami) — 2 scenarios
  - [prints the current user name](#scenario-prints-the-current-user-name)
  - [reports an error with an extra operand](#scenario-reports-an-error-with-an-extra-operand)
- [mimixbox yes](#mimixbox-yes) — 2 scenarios
  - [repeats y until the reader closes](#scenario-repeats-y-until-the-reader-closes)
  - [repeats the given string](#scenario-repeats-the-given-string)
- [mimixbox base32](#mimixbox-base32) — 2 scenarios
  - [encodes standard input](#scenario-encodes-standard-input-1)
  - [decodes standard input](#scenario-decodes-standard-input-1)
- [mimixbox cat](#mimixbox-cat) — 8 scenarios
  - [show shell family name](#scenario-show-shell-family-name)
  - [show shell family name with line numbers](#scenario-show-shell-family-name-with-line-numbers)
  - [show the piped path unchanged](#scenario-show-the-piped-path-unchanged)
  - [cat only the file operand, ignoring pipe data](#scenario-cat-only-the-file-operand-ignoring-pipe-data)
  - [concatenate two files](#scenario-concatenate-two-files)
  - [concatenate two files with line numbers](#scenario-concatenate-two-files-with-line-numbers)
  - [concatenate a heredoc and a file via redirect](#scenario-concatenate-a-heredoc-and-a-file-via-redirect)
  - [show error for a missing file](#scenario-show-error-for-a-missing-file)
- [mimixbox cat_showall](#mimixbox-cat_showall) — 4 scenarios
  - [-A and --show-all are aliases](#scenario--a-and---show-all-are-aliases)
  - [-v and --show-nonprinting are aliases](#scenario--v-and---show-nonprinting-are-aliases)
  - [--show-all renders tabs as ^I, non-printing bytes, and $ line ends](#scenario---show-all-renders-tabs-as-i-non-printing-bytes-and--line-ends)
  - [--show-nonprinting leaves TAB alone and renders ^X, ^?, and M- notation](#scenario---show-nonprinting-leaves-tab-alone-and-renders-x--and-m--notation)
- [mimixbox checksum](#mimixbox-checksum) — 3 scenarios
  - [sum prints the BSD checksum and block count](#scenario-sum-prints-the-bsd-checksum-and-block-count)
  - [crc32 prints the CRC-32 of stdin](#scenario-crc32-prints-the-crc-32-of-stdin)
  - [sha384sum prints the SHA-384 digest](#scenario-sha384sum-prints-the-sha-384-digest)
- [mimixbox cksum](#mimixbox-cksum) — 1 scenario
  - [prints the CRC checksum and byte count](#scenario-prints-the-crc-checksum-and-byte-count)
- [mimixbox comm](#mimixbox-comm) — 1 scenario
  - [print lines common to both files](#scenario-print-lines-common-to-both-files)
- [mimixbox comm_gnu](#mimixbox-comm_gnu) — 3 scenarios
  - [separate the columns with --output-delimiter](#scenario-separate-the-columns-with---output-delimiter)
  - [read and write NUL-terminated records with -z](#scenario-read-and-write-nul-terminated-records-with--z)
  - [report an unsorted input on stderr and fail with --check-order](#scenario-report-an-unsorted-input-on-stderr-and-fail-with---check-order)
- [mimixbox convert_mode](#mimixbox-convert_mode) — 2 scenarios
  - [dos2unix keeps the original mode](#scenario-dos2unix-keeps-the-original-mode)
  - [unix2dos keeps the original mode](#scenario-unix2dos-keeps-the-original-mode)
- [mimixbox crc32](#mimixbox-crc32) — 2 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-185)
  - [prints the CRC-32 of stdin](#scenario-prints-the-crc-32-of-stdin)
- [mimixbox dos2unix](#mimixbox-dos2unix) — 6 scenarios
  - [convert a CRLF file to LF and reclassify it](#scenario-convert-a-crlf-file-to-lf-and-reclassify-it)
  - [convert a CRLF file and exit success](#scenario-convert-a-crlf-file-and-exit-success)
  - [convert three CRLF files at once and reclassify each](#scenario-convert-three-crlf-files-at-once-and-reclassify-each)
  - [convert three CRLF files at once and exit success](#scenario-convert-three-crlf-files-at-once-and-exit-success)
  - [refuse a directory with a not-regular-file error](#scenario-refuse-a-directory-with-a-not-regular-file-error)
  - [convert the two files but fail on the directory operand](#scenario-convert-the-two-files-but-fail-on-the-directory-operand)
- [mimixbox expand](#mimixbox-expand) — 4 scenarios
  - [converts tabs to spaces (default tab stop 8)](#scenario-converts-tabs-to-spaces-default-tab-stop-8)
  - [converts tabs to the given width](#scenario-converts-tabs-to-the-given-width)
  - [converts tabs in the file](#scenario-converts-tabs-in-the-file)
  - [reports an error for a non-existent file](#scenario-reports-an-error-for-a-non-existent-file-1)
- [mimixbox fmt](#mimixbox-fmt) — 1 scenario
  - [reflows text to the given width](#scenario-reflows-text-to-the-given-width)
- [mimixbox fold](#mimixbox-fold) — 1 scenario
  - [wraps lines to the given width](#scenario-wraps-lines-to-the-given-width)
- [mimixbox head](#mimixbox-head) — 5 scenarios
  - [print the first 10 lines](#scenario-print-the-first-10-lines)
  - [print the first N lines](#scenario-print-the-first-n-lines)
  - [print the first N bytes](#scenario-print-the-first-n-bytes)
  - [print the first N lines of stdin](#scenario-print-the-first-n-lines-of-stdin)
  - [report an error for a non-existent file](#scenario-report-an-error-for-a-non-existent-file)
- [mimixbox head --zero-terminated](#mimixbox-head---zero-terminated) — 2 scenarios
  - [prints the first NUL-delimited record, preserving the embedded newline](#scenario-prints-the-first-nul-delimited-record-preserving-the-embedded-newline)
  - [prints two NUL-delimited records with embedded newlines preserved](#scenario-prints-two-nul-delimited-records-with-embedded-newlines-preserved)
- [mimixbox textutils --help helpers](#mimixbox-textutils---help-helpers) — 6 scenarios
  - [crc32 --help is structured](#scenario-crc32---help-is-structured)
  - [sha384sum --help is structured](#scenario-sha384sum---help-is-structured)
  - [sha3sum --help is structured](#scenario-sha3sum---help-is-structured)
  - [sum --help is structured](#scenario-sum---help-is-structured)
  - [uudecode --help is structured](#scenario-uudecode---help-is-structured)
  - [uuencode --help is structured](#scenario-uuencode---help-is-structured)
- [mimixbox man](#mimixbox-man) — 3 scenarios
  - [show a plain manual page](#scenario-show-a-plain-manual-page)
  - [decompress a gzipped manual page](#scenario-decompress-a-gzipped-manual-page)
  - [report a missing page with exit 16](#scenario-report-a-missing-page-with-exit-16)
- [mimixbox md5sum](#mimixbox-md5sum) — 7 scenarios
  - [get md5sum of one file](#scenario-get-md5sum-of-one-file)
  - [cannot get md5sum of one directory](#scenario-cannot-get-md5sum-of-one-directory)
  - [cannot get md5sum of not exist file](#scenario-cannot-get-md5sum-of-not-exist-file)
  - [get md5sum of three files](#scenario-get-md5sum-of-three-files)
  - [check md5sum with --check option](#scenario-check-md5sum-with---check-option)
  - [get md5sum for pipe data](#scenario-get-md5sum-for-pipe-data)
  - [get md5sum for pipe data and file at same time](#scenario-get-md5sum-for-pipe-data-and-file-at-same-time)
- [mimixbox nl](#mimixbox-nl) — 7 scenarios
  - [number each line of a file](#scenario-number-each-line-of-a-file)
  - [number the single line read from pipe data](#scenario-number-the-single-line-read-from-pipe-data)
  - [number only the file operand, ignoring pipe data](#scenario-number-only-the-file-operand-ignoring-pipe-data)
  - [number lines across two files](#scenario-number-lines-across-two-files)
  - [number heredoc then file via a redirect](#scenario-number-heredoc-then-file-via-a-redirect)
  - [number heredoc then file via a redirect with success status](#scenario-number-heredoc-then-file-via-a-redirect-with-success-status)
  - [report an error for a non-existent file](#scenario-report-an-error-for-a-non-existent-file-1)
- [mimixbox nl sections](#mimixbox-nl-sections) — 3 scenarios
  - [number every line in every section with -h a -b a -f a](#scenario-number-every-line-in-every-section-with--h-a--b-a--f-a)
  - [number header (a) and body (t) but not footer (n)](#scenario-number-header-a-and-body-t-but-not-footer-n)
  - [number every second blank line with -l 2](#scenario-number-every-second-blank-line-with--l-2)
- [mimixbox paste](#mimixbox-paste) — 1 scenario
  - [joins lines with a delimiter](#scenario-joins-lines-with-a-delimiter)
- [mimixbox rev](#mimixbox-rev) — 1 scenario
  - [reverses the characters of a line](#scenario-reverses-the-characters-of-a-line)
- [mimixbox sha1sum](#mimixbox-sha1sum) — 7 scenarios
  - [get sha1sum of one file](#scenario-get-sha1sum-of-one-file)
  - [cannot get sha1sum of one directory](#scenario-cannot-get-sha1sum-of-one-directory)
  - [cannot get sha1sum of not exist file](#scenario-cannot-get-sha1sum-of-not-exist-file)
  - [get sha1sum of three files](#scenario-get-sha1sum-of-three-files)
  - [check sha1sum with --check option](#scenario-check-sha1sum-with---check-option)
  - [get sha1sum for pipe data](#scenario-get-sha1sum-for-pipe-data)
  - [get sha1sum for pipe data and file at same time](#scenario-get-sha1sum-for-pipe-data-and-file-at-same-time)
- [mimixbox sha256sum](#mimixbox-sha256sum) — 7 scenarios
  - [get sha256sum of one file](#scenario-get-sha256sum-of-one-file)
  - [cannot get sha256sum of one directory](#scenario-cannot-get-sha256sum-of-one-directory)
  - [cannot get sha256sum of not exist file](#scenario-cannot-get-sha256sum-of-not-exist-file)
  - [get sha256sum of three files](#scenario-get-sha256sum-of-three-files)
  - [check sha256sum with --check option](#scenario-check-sha256sum-with---check-option)
  - [get sha256sum for pipe data](#scenario-get-sha256sum-for-pipe-data)
  - [get sha256sum for pipe data and file at same time](#scenario-get-sha256sum-for-pipe-data-and-file-at-same-time)
- [mimixbox sha384sum](#mimixbox-sha384sum) — 2 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-186)
  - [prints the SHA-384 digest of stdin](#scenario-prints-the-sha-384-digest-of-stdin)
- [mimixbox sha3sum](#mimixbox-sha3sum) — 2 scenarios
  - [defaults to SHA3-256](#scenario-defaults-to-sha3-256)
  - [selects SHA3-512 with -a](#scenario-selects-sha3-512-with--a)
- [mimixbox sha512sum](#mimixbox-sha512sum) — 7 scenarios
  - [get sha512sum of one file](#scenario-get-sha512sum-of-one-file)
  - [cannot get sha512sum of one directory](#scenario-cannot-get-sha512sum-of-one-directory)
  - [cannot get sha512sum of not exist file](#scenario-cannot-get-sha512sum-of-not-exist-file)
  - [get sha512sum of three files](#scenario-get-sha512sum-of-three-files)
  - [check sha512sum with --check option](#scenario-check-sha512sum-with---check-option)
  - [get sha512sum for pipe data](#scenario-get-sha512sum-for-pipe-data)
  - [get sha512sum for pipe data and file at same time](#scenario-get-sha512sum-for-pipe-data-and-file-at-same-time)
- [mimixbox shuf](#mimixbox-shuf) — 1 scenario
  - [shuffles a single-element range](#scenario-shuffles-a-single-element-range)
- [mimixbox split](#mimixbox-split) — 1 scenario
  - [split input into files of N lines](#scenario-split-input-into-files-of-n-lines)
- [mimixbox split (GNU flags)](#mimixbox-split-gnu-flags) — 4 scenarios
  - [use numeric suffixes with -d](#scenario-use-numeric-suffixes-with--d)
  - [write expected content to the first numeric piece](#scenario-write-expected-content-to-the-first-numeric-piece)
  - [append an additional suffix to each name](#scenario-append-an-additional-suffix-to-each-name)
  - [honor a custom suffix length with -a](#scenario-honor-a-custom-suffix-length-with--a)
- [mimixbox sqluv](#mimixbox-sqluv) — 6 scenarios
  - [print usage with --help and exit 0](#scenario-print-usage-with---help-and-exit-0)
  - [print the version and exit 0](#scenario-print-the-version-and-exit-0)
  - [fail with a message when given no operand](#scenario-fail-with-a-message-when-given-no-operand)
  - [query a CSV fixture in headless mode](#scenario-query-a-csv-fixture-in-headless-mode)
  - [query a SQLite-style table as JSON](#scenario-query-a-sqlite-style-table-as-json)
  - [fail deterministically on an unsupported S3 source](#scenario-fail-deterministically-on-an-unsupported-s3-source)
- [mimixbox sqluv (compressed input)](#mimixbox-sqluv-compressed-input) — 1 scenario
  - [query a gzip-compressed CSV fixture](#scenario-query-a-gzip-compressed-csv-fixture)
- [mimixbox sqluv (history file)](#mimixbox-sqluv-history-file) — 1 scenario
  - [write query history to the path given by --history-file](#scenario-write-query-history-to-the-path-given-by---history-file)
- [mimixbox sqluv (TUI smoke)](#mimixbox-sqluv-tui-smoke) — 1 scenario
  - [render the minimal viewer and exit cleanly on quit](#scenario-render-the-minimal-viewer-and-exit-cleanly-on-quit)
- [mimixbox strings](#mimixbox-strings) — 1 scenario
  - [prints printable sequences](#scenario-prints-printable-sequences)
- [mimixbox sum](#mimixbox-sum) — 3 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-187)
  - [documents its purpose in --help](#scenario-documents-its-purpose-in---help-13)
  - [prints a BSD checksum and block count for stdin](#scenario-prints-a-bsd-checksum-and-block-count-for-stdin)
- [mimixbox tac](#mimixbox-tac) — 3 scenarios
  - [print the lines in reverse order](#scenario-print-the-lines-in-reverse-order)
  - [reverse standard input](#scenario-reverse-standard-input)
  - [report an error for a non-existent file](#scenario-report-an-error-for-a-non-existent-file-2)
- [mimixbox tail](#mimixbox-tail) — 6 scenarios
  - [print the last 10 lines](#scenario-print-the-last-10-lines)
  - [print the last N lines](#scenario-print-the-last-n-lines)
  - [print the last N bytes](#scenario-print-the-last-n-bytes)
  - [print the last N lines of stdin](#scenario-print-the-last-n-lines-of-stdin)
  - [report an error for a non-existent file](#scenario-report-an-error-for-a-non-existent-file-3)
  - [print data appended while following](#scenario-print-data-appended-while-following)
- [mimixbox tail --pid](#mimixbox-tail---pid) — 1 scenario
  - [stop following once the watched process exits](#scenario-stop-following-once-the-watched-process-exits)
- [mimixbox tail --zero-terminated](#mimixbox-tail---zero-terminated) — 2 scenarios
  - [prints the last NUL-delimited record, preserving the embedded newline](#scenario-prints-the-last-nul-delimited-record-preserving-the-embedded-newline)
  - [prints two NUL-delimited records with embedded newlines preserved](#scenario-prints-two-nul-delimited-records-with-embedded-newlines-preserved-1)
- [mimixbox tr](#mimixbox-tr) — 1 scenario
  - [translates lowercase to uppercase](#scenario-translates-lowercase-to-uppercase)
- [mimixbox tr --truncate-set1](#mimixbox-tr---truncate-set1) — 2 scenarios
  - [truncates SET1 to SET2 length, leaving extra chars unchanged](#scenario-truncates-set1-to-set2-length-leaving-extra-chars-unchanged)
  - [accepts the -t short form](#scenario-accepts-the--t-short-form)
- [mimixbox unexpand](#mimixbox-unexpand) — 3 scenarios
  - [convert leading spaces to a tab](#scenario-convert-leading-spaces-to-a-tab)
  - [convert internal space runs to tabs with --all](#scenario-convert-internal-space-runs-to-tabs-with---all)
  - [report an error for a non-existent file](#scenario-report-an-error-for-a-non-existent-file-4)
- [mimixbox unix2dos](#mimixbox-unix2dos) — 6 scenarios
  - [convert an LF file to CRLF and reclassify it](#scenario-convert-an-lf-file-to-crlf-and-reclassify-it)
  - [convert an LF file and exit success](#scenario-convert-an-lf-file-and-exit-success)
  - [convert three LF files at once and reclassify each](#scenario-convert-three-lf-files-at-once-and-reclassify-each)
  - [convert three LF files at once and exit success](#scenario-convert-three-lf-files-at-once-and-exit-success)
  - [refuse a directory with a not-regular-file error](#scenario-refuse-a-directory-with-a-not-regular-file-error-1)
  - [convert the two files but fail on the directory operand](#scenario-convert-the-two-files-but-fail-on-the-directory-operand-1)
- [mimixbox uuencode/uudecode/usleep](#mimixbox-uuencodeuudecodeusleep) — 3 scenarios
  - [uuencode then uudecode round-trips (traditional)](#scenario-uuencode-then-uudecode-round-trips-traditional)
  - [uuencode -m then uudecode round-trips (base64)](#scenario-uuencode--m-then-uudecode-round-trips-base64)
  - [usleep waits and exits 0](#scenario-usleep-waits-and-exits-0)
- [mimixbox uudecode](#mimixbox-uudecode) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-188)
- [mimixbox uuencode](#mimixbox-uuencode) — 2 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-189)
  - [uuencodes stdin with a begin header](#scenario-uuencodes-stdin-with-a-begin-header)
- [mimixbox wc](#mimixbox-wc) — 11 scenarios
  - [count lines/words/bytes of one file](#scenario-count-lineswordsbytes-of-one-file)
  - [count only lines with --lines](#scenario-count-only-lines-with---lines)
  - [count only bytes with --bytes](#scenario-count-only-bytes-with---bytes)
  - [report the longest line with --max-line-length](#scenario-report-the-longest-line-with---max-line-length)
  - [count an empty file as all zeros](#scenario-count-an-empty-file-as-all-zeros)
  - [count three files and print a total](#scenario-count-three-files-and-print-a-total)
  - [count piped data](#scenario-count-piped-data)
  - [count only the file operand, ignoring pipe data](#scenario-count-only-the-file-operand-ignoring-pipe-data)
  - [report a directory as not a regular file](#scenario-report-a-directory-as-not-a-regular-file)
  - [count the file but zero the directory when given both](#scenario-count-the-file-but-zero-the-directory-when-given-both)
  - [count a single line piped in with --lines](#scenario-count-a-single-line-piped-in-with---lines)
- [mimixbox wc (GNU flags)](#mimixbox-wc-gnu-flags) — 6 scenarios
  - [print only the combined total with --total=only](#scenario-print-only-the-combined-total-with---totalonly)
  - [suppress the total line with --total=never](#scenario-suppress-the-total-line-with---totalnever)
  - [print a total even for one file with --total=always](#scenario-print-a-total-even-for-one-file-with---totalalways)
  - [read a NUL-separated name list with --files0-from](#scenario-read-a-nul-separated-name-list-with---files0-from)
  - [read the name list from standard input with --files0-from=-](#scenario-read-the-name-list-from-standard-input-with---files0-from-)
  - [combine --files0-from with --total=only](#scenario-combine---files0-from-with---totalonly)
- [mimixbox xxd](#mimixbox-xxd) — 2 scenarios
  - [prints a hex dump](#scenario-prints-a-hex-dump)
  - [reverses a hex dump](#scenario-reverses-a-hex-dump)
- [mimixbox blkdiscard](#mimixbox-blkdiscard) — 2 scenarios
  - [requires a device](#scenario-requires-a-device-1)
  - [describes itself with --help](#scenario-describes-itself-with---help-190)
- [mimixbox blkid](#mimixbox-blkid) — 3 scenarios
  - [identifies an ext filesystem](#scenario-identifies-an-ext-filesystem)
  - [identifies an xfs filesystem](#scenario-identifies-an-xfs-filesystem)
  - [exits 2 when nothing is identified](#scenario-exits-2-when-nothing-is-identified)
- [mimixbox blockdev](#mimixbox-blockdev) — 2 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-191)
  - [fails when no query flag is given](#scenario-fails-when-no-query-flag-is-given)
- [mimixbox chattr](#mimixbox-chattr) — 2 scenarios
  - [rejects a malformed mode](#scenario-rejects-a-malformed-mode)
  - [rejects an unknown attribute](#scenario-rejects-an-unknown-attribute)
- [mimixbox chrt](#mimixbox-chrt) — 2 scenarios
  - [prints a process scheduling policy](#scenario-prints-a-process-scheduling-policy)
  - [runs a command under a scheduling policy](#scenario-runs-a-command-under-a-scheduling-policy)
- [mimixbox dmesg](#mimixbox-dmesg) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-192)
- [mimixbox eject](#mimixbox-eject) — 2 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-193)
  - [fails on a missing device](#scenario-fails-on-a-missing-device)
- [mimixbox fallocate](#mimixbox-fallocate) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-194)
- [mimixbox fatattr](#mimixbox-fatattr) — 2 scenarios
  - [requires a file](#scenario-requires-a-file)
  - [rejects an unknown attribute](#scenario-rejects-an-unknown-attribute-1)
- [mimixbox fbset](#mimixbox-fbset) — 2 scenarios
  - [fails on a missing framebuffer](#scenario-fails-on-a-missing-framebuffer)
  - [describes itself with --help](#scenario-describes-itself-with---help-195)
- [mimixbox fdflush](#mimixbox-fdflush) — 2 scenarios
  - [requires a device](#scenario-requires-a-device-2)
  - [describes itself with --help](#scenario-describes-itself-with---help-196)
- [mimixbox fdformat](#mimixbox-fdformat) — 2 scenarios
  - [requires a device](#scenario-requires-a-device-3)
  - [describes itself with --help](#scenario-describes-itself-with---help-197)
- [mimixbox fdisk](#mimixbox-fdisk) — 2 scenarios
  - [lists an MBR Linux partition](#scenario-lists-an-mbr-linux-partition)
  - [rejects an image without an MBR signature](#scenario-rejects-an-image-without-an-mbr-signature)
- [mimixbox findfs](#mimixbox-findfs) — 2 scenarios
  - [fails for an unknown label](#scenario-fails-for-an-unknown-label)
  - [rejects a malformed tag](#scenario-rejects-a-malformed-tag)
- [mimixbox flock](#mimixbox-flock) — 2 scenarios
  - [runs a command while holding the lock](#scenario-runs-a-command-while-holding-the-lock)
  - [fails -n when the lock is already held](#scenario-fails--n-when-the-lock-is-already-held)
- [mimixbox freeramdisk](#mimixbox-freeramdisk) — 2 scenarios
  - [requires a device](#scenario-requires-a-device-4)
  - [describes itself with --help](#scenario-describes-itself-with---help-198)
- [mimixbox fsck](#mimixbox-fsck) — 2 scenarios
  - [detects a Minix filesystem](#scenario-detects-a-minix-filesystem)
  - [fails on an unrecognized image](#scenario-fails-on-an-unrecognized-image)
- [mimixbox fsck.minix](#mimixbox-fsckminix) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-199)
- [mimixbox fsck.minix](#mimixbox-fsckminix-1) — 2 scenarios
  - [validates a freshly made Minix filesystem](#scenario-validates-a-freshly-made-minix-filesystem)
  - [rejects a non-Minix image](#scenario-rejects-a-non-minix-image)
- [mimixbox fsfreeze](#mimixbox-fsfreeze) — 2 scenarios
  - [requires a freeze or unfreeze mode](#scenario-requires-a-freeze-or-unfreeze-mode)
  - [rejects both modes at once](#scenario-rejects-both-modes-at-once)
- [mimixbox fstrim](#mimixbox-fstrim) — 2 scenarios
  - [requires a mount point](#scenario-requires-a-mount-point)
  - [describes itself with --help](#scenario-describes-itself-with---help-200)
- [mimixbox getopt](#mimixbox-getopt) — 2 scenarios
  - [normalizes short and long options with quoted args](#scenario-normalizes-short-and-long-options-with-quoted-args)
  - [produces output a script can eval](#scenario-produces-output-a-script-can-eval)
- [mimixbox hd](#mimixbox-hd) — 2 scenarios
  - [describes itself with --help](#scenario-describes-itself-with---help-201)
  - [hexdumps stdin in canonical form](#scenario-hexdumps-stdin-in-canonical-form)
- [mimixbox util-linux --help helpers](#mimixbox-util-linux---help-helpers) — 14 scenarios
  - [fallocate --help is structured](#scenario-fallocate---help-is-structured)
  - [fsck.minix --help is structured](#scenario-fsckminix---help-is-structured)
  - [linux32 --help is structured](#scenario-linux32---help-is-structured)
  - [linux64 --help is structured](#scenario-linux64---help-is-structured)
  - [mkdosfs --help is structured](#scenario-mkdosfs---help-is-structured)
  - [mkfs.ext2 --help is structured](#scenario-mkfsext2---help-is-structured)
  - [mkfs.minix --help is structured](#scenario-mkfsminix---help-is-structured)
  - [mkfs.reiser --help is structured](#scenario-mkfsreiser---help-is-structured)
  - [mkfs.vfat --help is structured](#scenario-mkfsvfat---help-is-structured)
  - [scriptreplay --help is structured](#scenario-scriptreplay---help-is-structured)
  - [setsid --help is structured](#scenario-setsid---help-is-structured)
  - [sh --help is structured](#scenario-sh---help-is-structured)
  - [swapoff --help is structured](#scenario-swapoff---help-is-structured)
  - [swapon --help is structured](#scenario-swapon---help-is-structured)
- [mimixbox hexdump / hd](#mimixbox-hexdump--hd) — 3 scenarios
  - [hd shows the canonical hex+ASCII layout](#scenario-hd-shows-the-canonical-hexascii-layout)
  - [hexdump -C matches hd](#scenario-hexdump--c-matches-hd)
  - [hexdump default shows two-byte words](#scenario-hexdump-default-shows-two-byte-words)
- [mimixbox hwclock](#mimixbox-hwclock) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-202)
- [mimixbox ionice](#mimixbox-ionice) — 2 scenarios
  - [prints a process I/O class](#scenario-prints-a-process-io-class)
  - [runs a command at a given I/O class](#scenario-runs-a-command-at-a-given-io-class)
- [mimixbox ipcrm](#mimixbox-ipcrm) — 2 scenarios
  - [fails when nothing is requested](#scenario-fails-when-nothing-is-requested)
  - [fails to remove a non-existent id](#scenario-fails-to-remove-a-non-existent-id)
- [mimixbox ipcs](#mimixbox-ipcs) — 2 scenarios
  - [shows the IPC facility sections](#scenario-shows-the-ipc-facility-sections)
  - [limits to shared memory with -m](#scenario-limits-to-shared-memory-with--m)
- [mimixbox last](#mimixbox-last) — 2 scenarios
  - [treats an empty wtmp as no history and exits 0](#scenario-treats-an-empty-wtmp-as-no-history-and-exits-0)
  - [fails on a missing wtmp file](#scenario-fails-on-a-missing-wtmp-file)
- [mimixbox losetup](#mimixbox-losetup) — 2 scenarios
  - [lists active loop devices cleanly](#scenario-lists-active-loop-devices-cleanly)
  - [refuses to associate a loop device](#scenario-refuses-to-associate-a-loop-device)
- [mimixbox lsattr](#mimixbox-lsattr) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-203)
- [mimixbox lsblk](#mimixbox-lsblk) — 2 scenarios
  - [prints the column header](#scenario-prints-the-column-header-4)
  - [runs and exits successfully](#scenario-runs-and-exits-successfully-1)
- [mimixbox lspci](#mimixbox-lspci) — 1 scenario
  - [lists PCI devices and exits successfully](#scenario-lists-pci-devices-and-exits-successfully)
- [mimixbox lsusb](#mimixbox-lsusb) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-204)
- [mimixbox mdev](#mimixbox-mdev) — 2 scenarios
  - [requires scan mode](#scenario-requires-scan-mode)
  - [describes itself with --help](#scenario-describes-itself-with---help-205)
- [mimixbox mesg](#mimixbox-mesg) — 1 scenario
  - [reports an error when stdin is not a terminal](#scenario-reports-an-error-when-stdin-is-not-a-terminal)
- [mimixbox mke2fs / mkfs.ext2](#mimixbox-mke2fs--mkfsext2) — 2 scenarios
  - [writes the ext2 magic](#scenario-writes-the-ext2-magic)
  - [mkfs.ext2 refuses an oversized image](#scenario-mkfsext2-refuses-an-oversized-image)
- [mimixbox mkfs.minix](#mimixbox-mkfsminix-1) — 2 scenarios
  - [writes the Minix v1 magic](#scenario-writes-the-minix-v1-magic)
  - [refuses a too-small device](#scenario-refuses-a-too-small-device)
- [mimixbox mkfs.reiser](#mimixbox-mkfsreiser-1) — 2 scenarios
  - [refuses deterministically](#scenario-refuses-deterministically)
  - [explains that ReiserFS is deprecated](#scenario-explains-that-reiserfs-is-deprecated)
- [mimixbox mkfs.vfat / mkdosfs](#mimixbox-mkfsvfat--mkdosfs) — 2 scenarios
  - [writes the FAT16 type label](#scenario-writes-the-fat16-type-label)
  - [mkdosfs refuses a too-small image](#scenario-mkdosfs-refuses-a-too-small-image)
- [mimixbox mkswap](#mimixbox-mkswap) — 2 scenarios
  - [formats an image as swap](#scenario-formats-an-image-as-swap)
  - [writes the swap signature](#scenario-writes-the-swap-signature)
- [mimixbox mount](#mimixbox-mount) — 2 scenarios
  - [lists the root filesystem](#scenario-lists-the-root-filesystem)
  - [refuses to perform a mount](#scenario-refuses-to-perform-a-mount)
- [mimixbox nsenter](#mimixbox-nsenter) — 2 scenarios
  - [requires a target PID](#scenario-requires-a-target-pid)
  - [requires a namespace flag](#scenario-requires-a-namespace-flag)
- [mimixbox pivot_root](#mimixbox-pivot_root) — 2 scenarios
  - [requires two directories](#scenario-requires-two-directories)
  - [describes itself with --help](#scenario-describes-itself-with---help-206)
- [mimixbox rdate](#mimixbox-rdate) — 2 scenarios
  - [fails when no host is given](#scenario-fails-when-no-host-is-given)
  - [fails when the host has no time service](#scenario-fails-when-the-host-has-no-time-service)
- [mimixbox rdev](#mimixbox-rdev) — 1 scenario
  - [prints the root device with the / mountpoint](#scenario-prints-the-root-device-with-the--mountpoint)
- [mimixbox readprofile](#mimixbox-readprofile) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-207)
- [mimixbox renice](#mimixbox-renice) — 2 scenarios
  - [reports the priority change](#scenario-reports-the-priority-change)
  - [rejects a non-numeric PID](#scenario-rejects-a-non-numeric-pid)
- [mimixbox rtcwake](#mimixbox-rtcwake) — 2 scenarios
  - [rejects a suspend mode](#scenario-rejects-a-suspend-mode)
  - [requires a wake time](#scenario-requires-a-wake-time)
- [mimixbox script / scriptreplay](#mimixbox-script--scriptreplay) — 2 scenarios
  - [records command output to a typescript](#scenario-records-command-output-to-a-typescript)
  - [replays a recorded typescript](#scenario-replays-a-recorded-typescript)
- [mimixbox script / scriptreplay round-trip](#mimixbox-script--scriptreplay-round-trip) — 4 scenarios
  - [records the transcript framing](#scenario-records-the-transcript-framing)
  - [records the command payload in the transcript](#scenario-records-the-command-payload-in-the-transcript)
  - [writes a timing file of "delay bytes" records](#scenario-writes-a-timing-file-of-delay-bytes-records)
  - [replays the captured payload from the timing + transcript](#scenario-replays-the-captured-payload-from-the-timing--transcript)
- [mimixbox scriptreplay](#mimixbox-scriptreplay) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-208)
- [mimixbox setarch / linux32 / linux64](#mimixbox-setarch--linux32--linux64) — 4 scenarios
  - [linux32 makes uname report a 32-bit machine](#scenario-linux32-makes-uname-report-a-32-bit-machine)
  - [linux64 reports the native machine](#scenario-linux64-reports-the-native-machine)
  - [linux32 passes the command output through](#scenario-linux32-passes-the-command-output-through)
  - [setarch selects the personality from ARCH](#scenario-setarch-selects-the-personality-from-arch)
- [mimixbox setpriv](#mimixbox-setpriv) — 2 scenarios
  - [dumps the current privileges](#scenario-dumps-the-current-privileges)
  - [runs a command with --no-new-privs](#scenario-runs-a-command-with---no-new-privs)
- [mimixbox setsid](#mimixbox-setsid) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-209)
- [mimixbox setsid / fallocate](#mimixbox-setsid--fallocate) — 3 scenarios
  - [setsid runs a program in a new session](#scenario-setsid-runs-a-program-in-a-new-session)
  - [fallocate sizes a file to the requested length](#scenario-fallocate-sizes-a-file-to-the-requested-length)
  - [fallocate without -l fails](#scenario-fallocate-without--l-fails)
- [mimixbox swapon / swapoff](#mimixbox-swapon--swapoff) — 2 scenarios
  - [swapon -s prints the swaps header](#scenario-swapon--s-prints-the-swaps-header)
  - [swapoff requires a target](#scenario-swapoff-requires-a-target)
- [mimixbox swapoff](#mimixbox-swapoff) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-210)
- [mimixbox swapon](#mimixbox-swapon) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-211)
- [mimixbox switch_root](#mimixbox-switch_root) — 2 scenarios
  - [requires NEW_ROOT and INIT](#scenario-requires-new_root-and-init-1)
  - [rejects a non-directory NEW_ROOT](#scenario-rejects-a-non-directory-new_root-1)
- [mimixbox taskset](#mimixbox-taskset) — 3 scenarios
  - [prints a process affinity mask](#scenario-prints-a-process-affinity-mask)
  - [runs a command bound to a CPU](#scenario-runs-a-command-bound-to-a-cpu)
  - [rejects an invalid mask](#scenario-rejects-an-invalid-mask)
- [mimixbox tune2fs](#mimixbox-tune2fs) — 2 scenarios
  - [rejects a non-ext image](#scenario-rejects-a-non-ext-image)
  - [describes itself with --help](#scenario-describes-itself-with---help-212)
- [mimixbox uevent](#mimixbox-uevent) — 1 scenario
  - [describes itself with --help](#scenario-describes-itself-with---help-213)
- [mimixbox umount](#mimixbox-umount) — 2 scenarios
  - [fails for a target that is not mounted](#scenario-fails-for-a-target-that-is-not-mounted)
  - [requires a target](#scenario-requires-a-target)
- [mimixbox unshare](#mimixbox-unshare) — 2 scenarios
  - [requires a namespace flag](#scenario-requires-a-namespace-flag-1)
  - [describes itself with --help](#scenario-describes-itself-with---help-214)
- [mimixbox wall](#mimixbox-wall) — 1 scenario
  - [runs and exits successfully](#scenario-runs-and-exits-successfully-2)
## mimixbox ar
Source: `test/e2e/tools/mimixbox/archival/ar.atago.yaml`
### Scenario: lists members
#### Given
- Fixture file `a.txt` is created.
- Fixture file `b.txt` is created.
#### Inputs
_Fixture `a.txt`:_
```
alpha
```
_Fixture `b.txt`:_
```
beta
```
#### When
```shell
ar rc lib.a a.txt b.txt && ar t lib.a
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout equals an exact value
### Scenario: extracts a member
#### Given
- Fixture file `a.txt` is created.
#### Inputs
_Fixture `a.txt`:_
```
alpha
```
#### When
```shell
ar rc lib2.a a.txt && mkdir -p out && cd out && ar x ../lib2.a && cat a.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox bunzip2
Source: `test/e2e/tools/mimixbox/archival/bunzip2.atago.yaml`
### Scenario: decompresses a .bz2 file to stdout with -c
#### When
```shell
printf 'bunzip2 payload' | bzip2 -c > data.bz2 && bunzip2 -c data.bz2
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox bzcat
Source: `test/e2e/tools/mimixbox/archival/bzcat.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
bzcat --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: bzcat`
- stderr is empty
## mimixbox bzip2
Source: `test/e2e/tools/mimixbox/archival/bzip2.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
bzip2 --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: bzip2`
- stderr is empty
## mimixbox compress and uncompress
Source: `test/e2e/tools/mimixbox/archival/compress.atago.yaml`
### Scenario: round-trips a file through compress and uncompress
#### Given
- Fixture file `data.txt` is created.
#### Inputs
_Fixture `data.txt`:_
```
compress me compress me compress me
```
#### When
```shell
cp data.txt rt.txt && compress rt.txt && uncompress rt.txt.Z && cat rt.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox cpio
Source: `test/e2e/tools/mimixbox/archival/cpio.atago.yaml`
### Scenario: round-trips a file through -o and -i
#### Given
- Fixture file `in/file.txt` is created.
#### Inputs
_Fixture `in/file.txt`:_
```
payload
```
#### When
```shell
cd in
printf 'file.txt\n' | cpio -o > ../arch.cpio
cd ..
mkdir -p out && cd out
cpio -i < ../arch.cpio
cat file.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: lists archive contents with -i -t
#### Given
- Fixture file `in/file.txt` is created.
#### Inputs
_Fixture `in/file.txt`:_
```
payload
```
#### When
```shell
cd in
printf 'file.txt\n' | cpio -o > ../arch2.cpio
cd ..
cpio -i -t < arch2.cpio

```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox bzip2, lzop and Debian package applets
Source: `test/e2e/tools/mimixbox/archival/debcomp.atago.yaml`
### Scenario: bzip2 | bzip2 -dc round-trips data
#### When
```shell
printf 'roundtrip-bzip2\n' | bzip2 | bzip2 -dc
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: lzop | lzopcat round-trips data
#### When
```shell
printf 'roundtrip-lzop\n' | lzop | lzopcat
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: lzop | unlzop -c round-trips data
#### When
```shell
printf 'roundtrip-unlzop\n' | lzop | unlzop -c
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: dpkg-deb -c lists package contents
#### Given
- Fixture file `hello.deb` is created.
#### When
```shell
dpkg-deb -c hello.deb | grep -q 'usr/bin/hello' && printf 'has-hello\n'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: dpkg-deb -f prints a control field
#### Given
- Fixture file `hello.deb` is created.
#### When
```shell
dpkg-deb -f hello.deb Package
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: dpkg -x extracts the data tarball path-safely
#### Given
- Fixture file `hello.deb` is created.
#### When
```shell
dpkg -x hello.deb out
test -f out/usr/bin/hello && printf 'extracted\n'

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: dpkg rejects unsupported database operations
#### Given
- Fixture file `hello.deb` is created.
#### When
```shell
if dpkg -i hello.deb 2>/dev/null; then
  printf 'unexpected-success\n'
else
  printf 'rejected\n'
fi

```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox dpkg-deb
Source: `test/e2e/tools/mimixbox/archival/dpkg-deb.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
dpkg-deb --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: dpkg-deb`
- stderr is empty
## mimixbox dpkg
Source: `test/e2e/tools/mimixbox/archival/dpkg.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
dpkg --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: dpkg`
- stderr is empty
## mimixbox gunzip
Source: `test/e2e/tools/mimixbox/archival/gunzip.atago.yaml`
### Scenario: decompresses a .gz file to stdout with -c
#### When
```shell
printf 'gunzip payload' > data && gzip data && gunzip -c data.gz
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox archival commands expose a dedicated --help helper
Source: `test/e2e/tools/mimixbox/archival/help_helpers_archival.atago.yaml`
### Scenario: bzcat --help is structured
#### When
```shell
bzcat --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: bzip2 --help is structured
#### When
```shell
bzip2 --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: dpkg --help is structured
#### When
```shell
dpkg --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: dpkg-deb --help is structured
#### When
```shell
dpkg-deb --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: lzcat --help is structured
#### When
```shell
lzcat --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: lzma --help is structured
#### When
```shell
lzma --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: lzopcat --help is structured
#### When
```shell
lzopcat --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: pipe_progress --help is structured
#### When
```shell
pipe_progress --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: rpm2cpio --help is structured
#### When
```shell
rpm2cpio --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: uncompress --help is structured
#### When
```shell
uncompress --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: unlzma --help is structured
#### When
```shell
unlzma --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: unlzop --help is structured
#### When
```shell
unlzop --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: unxz --help is structured
#### When
```shell
unxz --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: unzip --help is structured
#### When
```shell
unzip --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: xz --help is structured
#### When
```shell
xz --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: xzcat --help is structured
#### When
```shell
xzcat --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: zcat --help is structured
#### When
```shell
zcat --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
## mimixbox rpm and rpm2cpio
Source: `test/e2e/tools/mimixbox/archival/rpm.atago.yaml`
### Scenario: queries the package identity with rpm -qp
#### Given
- Fixture file `sample.rpm` is created.
#### When
```shell
rpm -qp sample.rpm
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: lists package files with rpm -qpl
#### Given
- Fixture file `sample.rpm` is created.
#### When
```shell
rpm -qpl sample.rpm
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout equals an exact value
### Scenario: extracts the payload with rpm2cpio
#### Given
- Fixture file `sample.rpm` is created.
#### When
```shell
rpm2cpio sample.rpm
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox rpm2cpio
Source: `test/e2e/tools/mimixbox/archival/rpm2cpio.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
rpm2cpio --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: rpm2cpio`
- stderr is empty
## mimixbox tar
Source: `test/e2e/tools/mimixbox/archival/tar.atago.yaml`
### Scenario: creates and extracts an archive
#### Given
- Fixture file `src/a.txt` is created.
#### Inputs
_Fixture `src/a.txt`:_
```
alpha
```
#### When
```shell
tar -c -f out.tar -C "${workdir}" src
mkdir -p dest
tar -x -f out.tar -C "${workdir}/dest"
cat dest/src/a.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: lists archive contents
#### Given
- Fixture file `src/a.txt` is created.
#### Inputs
_Fixture `src/a.txt`:_
```
alpha
```
#### When
```shell
tar -c -f list.tar -C "${workdir}" src
tar -t -f list.tar

```
#### Then
- exit code is `0`
- stdout contains `src/a.txt`
## mimixbox uncompress
Source: `test/e2e/tools/mimixbox/archival/uncompress.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
uncompress --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: uncompress`
- stderr is empty
## mimixbox unlzma
Source: `test/e2e/tools/mimixbox/archival/unlzma.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
unlzma --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: unlzma`
- stderr is empty
## mimixbox unlzop
Source: `test/e2e/tools/mimixbox/archival/unlzop.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
unlzop --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: unlzop`
- stderr is empty
## mimixbox unxz
Source: `test/e2e/tools/mimixbox/archival/unxz.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
unxz --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: unxz`
- stderr is empty
## mimixbox unzip
Source: `test/e2e/tools/mimixbox/archival/unzip.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
unzip --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: unzip`
- stderr is empty
## mimixbox xz
Source: `test/e2e/tools/mimixbox/archival/xz.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
xz --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: xz`
- stderr is empty
## mimixbox xzcat
Source: `test/e2e/tools/mimixbox/archival/xzcat.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
xzcat --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: xzcat`
- stderr is empty
## mimixbox compression applets
Source: `test/e2e/tools/mimixbox/archival/xzcomp.atago.yaml`
### Scenario: xz | xzcat round-trips data
#### When
```shell
printf 'roundtrip-xz\n' | xz | xzcat
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: lzma | unlzma round-trips data
#### When
```shell
printf 'roundtrip-lzma\n' | lzma | unlzma
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: zcat decompresses a gzip file to stdout
#### When
```shell
printf 'gz-payload\n' | gzip > f.gz && zcat f.gz
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: pipe_progress passes stdin through to stdout
#### When
```shell
printf 'pass-through\n' | pipe_progress 2>/dev/null
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox zcat
Source: `test/e2e/tools/mimixbox/archival/zcat.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
zcat --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: zcat`
- stderr is empty
## mimixbox zip and unzip
Source: `test/e2e/tools/mimixbox/archival/zip.atago.yaml`
### Scenario: lists a zipped file via unzip -l
#### Given
- Fixture file `a.txt` is created.
#### Inputs
_Fixture `a.txt`:_
```
zipped
```
#### When
```shell
zip out.zip a.txt >/dev/null && unzip -l out.zip
```
#### Then
- exit code is `0`
- stdout contains `a.txt`
### Scenario: round-trips a file through zip and unzip
#### Given
- Fixture file `a.txt` is created.
#### Inputs
_Fixture `a.txt`:_
```
zipped
```
#### When
```shell
zip out2.zip a.txt >/dev/null
unzip -d dest out2.zip >/dev/null
cat dest/a.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox [
Source: `test/e2e/tools/mimixbox/compat/[.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
[ --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: [`
- stderr is empty
### Scenario: documents its purpose in --help
#### When
```shell
[ --help
```
#### Then
- exit code is `0`
- stdout contains `Evaluate EXPRESSION`
## mimixbox [[
Source: `test/e2e/tools/mimixbox/compat/[[.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
[[ --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: [[`
- stderr is empty
### Scenario: documents its purpose in --help
#### When
```shell
[[ --help
```
#### Then
- exit code is `0`
- stdout contains `Evaluate EXPRESSION`
## mimixbox ash
Source: `test/e2e/tools/mimixbox/compat/ash.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
ash --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: ash`
- stderr is empty
### Scenario: documents its purpose in --help
#### When
```shell
ash --help
```
#### Then
- exit code is `0`
- stdout contains `compatibility front-end over MimixBox`
## mimixbox bash
Source: `test/e2e/tools/mimixbox/compat/bash.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
bash --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: bash`
- stderr is empty
### Scenario: documents its purpose in --help
#### When
```shell
bash --help
```
#### Then
- exit code is `0`
- stdout contains `compatibility front-end over MimixBox`
## mimixbox busybox
Source: `test/e2e/tools/mimixbox/compat/busybox.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
busybox --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: busybox`
- stderr is empty
### Scenario: documents its purpose in --help
#### When
```shell
busybox --help
```
#### Then
- exit code is `0`
- stdout contains `multi-call front-end`
## mimixbox compat front-ends
Source: `test/e2e/tools/mimixbox/compat/compat.atago.yaml`
### Scenario: the [ alias returns true for an existing file
#### When
```shell
mimixbox '[' -f /etc/hosts ']' && echo yes || echo no
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: the [ alias returns false for a missing file
#### When
```shell
mimixbox '[' -f /no/such/mimixbox/file ']' && echo yes || echo no
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: busybox dispatches to an applet
#### Given
- Fixture file `f` is created.
#### Inputs
_Fixture `f`:_
```
hello
```
#### When
```shell
busybox cat f
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: busybox --list shows applets
#### When
```shell
busybox --list
```
#### Then
- exit code is `0`
- stdout contains `cat`
- stdout contains `busybox`
### Scenario: sh -c runs a command without a prompt
#### When
```shell
sh -c 'echo from-sh'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: bash reads a non-interactive script from stdin without a prompt
#### When
```shell
out=$(printf 'echo via-bash\n' | bash 2>/dev/null)
case "$out" in
    *"mbsh:"*) echo prompted ;;
    *via-bash*) echo ok ;;
    *) echo "$out" ;;
esac

```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox cttyhack
Source: `test/e2e/tools/mimixbox/compat/cttyhack.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
cttyhack --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: cttyhack`
- stderr is empty
## mimixbox compat commands expose a dedicated --help helper
Source: `test/e2e/tools/mimixbox/compat/help_helpers_compat.atago.yaml`
### Scenario: [ --help is structured
#### When
```shell
[ --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: [[ --help is structured
#### When
```shell
[[ --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: ash --help is structured
#### When
```shell
ash --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: bash --help is structured
#### When
```shell
bash --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: busybox --help is structured
#### When
```shell
busybox --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: cttyhack --help is structured
#### When
```shell
cttyhack --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: hush --help is structured
#### When
```shell
hush --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: unit --help is structured
#### When
```shell
unit --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
## mimixbox hush
Source: `test/e2e/tools/mimixbox/compat/hush.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
hush --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: hush`
- stderr is empty
### Scenario: documents its purpose in --help
#### When
```shell
hush --help
```
#### Then
- exit code is `0`
- stdout contains `compatibility front-end over MimixBox`
## mimixbox sh
Source: `test/e2e/tools/mimixbox/compat/sh.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
env sh --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: sh`
- stderr is empty
### Scenario: documents its purpose in --help
#### When
```shell
env sh --help
```
#### Then
- exit code is `0`
- stdout contains `compatibility front-end over MimixBox`
## mimixbox unit
Source: `test/e2e/tools/mimixbox/compat/unit.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
unit --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: unit`
- stderr is empty
### Scenario: documents its purpose in --help
#### When
```shell
unit --help
```
#### Then
- exit code is `0`
- stdout contains `does not embed`
## mimixbox adjtimex
Source: `test/e2e/tools/mimixbox/console-tools/adjtimex.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
adjtimex --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: adjtimex`
- stderr is empty
## mimixbox ascii
Source: `test/e2e/tools/mimixbox/console-tools/ascii.atago.yaml`
### Scenario: prints 128 entries
#### When
```shell
ascii | grep -c .
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: maps code 65 to A
#### When
```shell
ascii | grep '0x41' | grep -c 'A'
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox bbconfig
Source: `test/e2e/tools/mimixbox/console-tools/bbconfig.atago.yaml`
### Scenario: prints the version line
#### When
```shell
bbconfig | grep -c 'CONFIG_MIMIXBOX_VERSION='
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: lists itself among the applets
#### When
```shell
bbconfig --names | grep -c '^bbconfig$'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: rejects an unexpected argument
#### When
```shell
bbconfig extra
```
#### Then
- exit code is `1`
## mimixbox beep
Source: `test/e2e/tools/mimixbox/console-tools/beep.atago.yaml`
### Scenario: rejects a non-positive frequency
#### When
```shell
beep -f 0
```
#### Then
- exit code is `1`
### Scenario: rejects a zero repeat count
#### When
```shell
beep -r 0
```
#### Then
- exit code is `1`
## mimixbox chat
Source: `test/e2e/tools/mimixbox/console-tools/chat.atago.yaml`
### Scenario: sends the reply after the expected string
#### When
```shell
printf 'OK' | chat OK GO | grep -c 'GO'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: requires a script
#### When
```shell
chat
```
#### Then
- exit code is `1`
### Scenario: fails when an expected string never arrives
#### Inputs
_stdin for `chat`:_
```
nope
```
#### When
```shell
chat LOGIN: user
```
#### Then
- exit code is `1`
## mimixbox chvt
Source: `test/e2e/tools/mimixbox/console-tools/chvt.atago.yaml`
### Scenario: rejects a non-numeric VT
#### When
```shell
chvt notanumber
```
#### Then
- exit code is `1`
### Scenario: requires a VT number
#### When
```shell
chvt
```
#### Then
- exit code is `1`
## mimixbox clear
Source: `test/e2e/tools/mimixbox/console-tools/clear.atago.yaml`
### Scenario: prints usage with --help and exits 0
#### When
```shell
clear --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: clear`
### Scenario: exits 0 when clearing the screen
#### When
```shell
clear
```
#### Then
- exit code is `0`
## mimixbox conspy
Source: `test/e2e/tools/mimixbox/console-tools/conspy.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
conspy --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: conspy`
- stderr is empty
## mimixbox deallocvt
Source: `test/e2e/tools/mimixbox/console-tools/deallocvt.atago.yaml`
### Scenario: rejects a non-numeric VT
#### When
```shell
deallocvt notanumber
```
#### Then
- exit code is `1`
### Scenario: describes itself with --help
#### When
```shell
deallocvt --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: deallocvt`, `virtual terminal`
## mimixbox dumpkmap
Source: `test/e2e/tools/mimixbox/console-tools/dumpkmap.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
dumpkmap --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: dumpkmap`
- stderr is empty
## mimixbox fgconsole
Source: `test/e2e/tools/mimixbox/console-tools/fgconsole.atago.yaml`
### Scenario: fails without a virtual console
#### When
```shell
fgconsole
```
#### Then
- exit code is `1`
### Scenario: describes itself with --help
#### When
```shell
fgconsole --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: fgconsole`, `virtual terminal`
## mimixbox console-tools --help contract
Source: `test/e2e/tools/mimixbox/console-tools/help_helpers_console-tools.atago.yaml`
### Scenario: adjtimex --help is structured
#### When
```shell
adjtimex --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: conspy --help is structured
#### When
```shell
conspy --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: dumpkmap --help is structured
#### When
```shell
dumpkmap --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: less --help is structured
#### When
```shell
less --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: loadfont --help is structured
#### When
```shell
loadfont --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: loadkmap --help is structured
#### When
```shell
loadkmap --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: microcom --help is structured
#### When
```shell
microcom --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: more --help is structured
#### When
```shell
more --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: openvt --help is structured
#### When
```shell
openvt --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: rx --help is structured
#### When
```shell
rx --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: setfont --help is structured
#### When
```shell
setfont --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
## mimixbox inotifyd
Source: `test/e2e/tools/mimixbox/console-tools/inotifyd.atago.yaml`
### Scenario: runs the handler on a create event
#### When
```shell
TEST_DIR="${workdir}/inotifyd"; mkdir -p "$TEST_DIR"
d="$TEST_DIR/w"; mkdir -p "$d"
printf '#!/bin/sh\necho "$1 $3" >> %s/events\n' "$TEST_DIR" > "$TEST_DIR/h"
chmod +x "$TEST_DIR/h"
: > "$TEST_DIR/events"
inotifyd "$TEST_DIR/h" "$d:n" &
pid=$!
found=missing
for _ in $(seq 1 50); do
    rm -f "$d/created.txt"
    touch "$d/created.txt"
    if grep -q 'n created.txt' "$TEST_DIR/events" 2>/dev/null; then
        found=ok
        break
    fi
    sleep 0.1
done
kill "$pid" 2>/dev/null
wait "$pid" 2>/dev/null
echo "$found"

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: requires a handler and a file
#### When
```shell
inotifyd ./h
```
#### Then
- exit code is `1`
## mimixbox kbd_mode
Source: `test/e2e/tools/mimixbox/console-tools/kbd_mode.atago.yaml`
### Scenario: rejects conflicting mode options
#### When
```shell
kbd_mode -a -u
```
#### Then
- exit code is `1`
### Scenario: describes itself with --help
#### When
```shell
kbd_mode --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: kbd_mode`, `keyboard`
## mimixbox loadfont
Source: `test/e2e/tools/mimixbox/console-tools/loadfont.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
loadfont --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: loadfont`
- stderr is empty
## mimixbox loadkmap
Source: `test/e2e/tools/mimixbox/console-tools/loadkmap.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
loadkmap --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: loadkmap`
- stderr is empty
## mimixbox microcom
Source: `test/e2e/tools/mimixbox/console-tools/microcom.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
microcom --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: microcom`
- stderr is empty
## mimixbox openvt
Source: `test/e2e/tools/mimixbox/console-tools/openvt.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
openvt --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: openvt`
- stderr is empty
## mimixbox more / less
Source: `test/e2e/tools/mimixbox/console-tools/pager.atago.yaml`
### Scenario: more streams stdin through when stdout is not a terminal
#### When
```shell
printf 'a\nb\nc\n' | more
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
a
b
c
```
### Scenario: less streams stdin through when stdout is not a terminal
#### When
```shell
printf 'x\ny\nz\n' | less
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
x
y
z
```
### Scenario: more streams a file through
#### Given
- Fixture file `f` is created.
#### Inputs
_Fixture `f`:_
```
one
two
```
#### When
```shell
more f
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
one
two
```
## mimixbox reset
Source: `test/e2e/tools/mimixbox/console-tools/reset.atago.yaml`
### Scenario: prints usage with --help and exits 0
#### When
```shell
reset --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: reset`
## mimixbox resize
Source: `test/e2e/tools/mimixbox/console-tools/resize.atago.yaml`
### Scenario: shows the usage line for --help
#### When
```shell
resize --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: resize`
## mimixbox rfkill
Source: `test/e2e/tools/mimixbox/console-tools/rfkill.atago.yaml`
### Scenario: lists devices cleanly
#### When
```shell
rfkill list
```
#### Then
- exit code is `0`
### Scenario: rejects an unknown command
#### When
```shell
rfkill bogus
```
#### Then
- exit code is `1`
## mimixbox rx
Source: `test/e2e/tools/mimixbox/console-tools/rx.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
rx --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: rx`
- stderr is empty
## mimixbox setconsole
Source: `test/e2e/tools/mimixbox/console-tools/setconsole.atago.yaml`
### Scenario: fails on an inaccessible device
#### When
```shell
setconsole /dev/no_such_console
```
#### Then
- exit code is `1`
### Scenario: describes itself with --help
#### When
```shell
setconsole --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: setconsole`, `console`
## mimixbox setfont
Source: `test/e2e/tools/mimixbox/console-tools/setfont.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
setfont --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: setfont`
- stderr is empty
## mimixbox setkeycodes
Source: `test/e2e/tools/mimixbox/console-tools/setkeycodes.atago.yaml`
### Scenario: requires arguments in pairs
#### When
```shell
setkeycodes e060
```
#### Then
- exit code is `1`
### Scenario: rejects an invalid scancode
#### When
```shell
setkeycodes zz 1
```
#### Then
- exit code is `1`
## mimixbox setlogcons
Source: `test/e2e/tools/mimixbox/console-tools/setlogcons.atago.yaml`
### Scenario: rejects a non-numeric VT
#### When
```shell
setlogcons notanumber
```
#### Then
- exit code is `1`
### Scenario: describes itself with --help
#### When
```shell
setlogcons --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: setlogcons`, `kernel`
## mimixbox setserial
Source: `test/e2e/tools/mimixbox/console-tools/setserial.atago.yaml`
### Scenario: echoes the parsed request with -g
#### When
```shell
setserial -g /dev/ttyS0 baud_base 115200 | grep -c 'baud_base 115200'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: rejects an unknown parameter
#### When
```shell
setserial /dev/ttyS0 bogus 1
```
#### Then
- exit code is `1`
### Scenario: requires a device
#### When
```shell
setserial
```
#### Then
- exit code is `1`
## mimixbox showkey
Source: `test/e2e/tools/mimixbox/console-tools/showkey.atago.yaml`
### Scenario: rejects conflicting modes
#### When
```shell
showkey -a -s
```
#### Then
- exit code is `1`
### Scenario: fails deterministically without a console
#### When
```shell
showkey
```
#### Then
- exit code is `1`
## mimixbox stty
Source: `test/e2e/tools/mimixbox/console-tools/stty.atago.yaml`
### Scenario: reports when standard input is not a terminal
#### Inputs
_stdin for `stty`:_
```
x
```
#### When
```shell
stty
```
#### Then
- exit code is `1`
- stderr equals an exact value
## mimixbox ts
Source: `test/e2e/tools/mimixbox/console-tools/ts.atago.yaml`
### Scenario: prefixes each line with a timestamp
#### When
```shell
printf 'alpha\nbeta\n' | ts | grep -cE '^[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} (alpha|beta)$'
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox ttysize
Source: `test/e2e/tools/mimixbox/console-tools/ttysize.atago.yaml`
### Scenario: prints width and height
#### When
```shell
ttysize
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: prints just the width with w
#### When
```shell
ttysize w
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox coreutils slice
Source: `test/e2e/tools/mimixbox/coreutils/coreutils_slice.atago.yaml`
### Scenario: factor prints prime factors
#### When
```shell
factor 360
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: tsort topologically sorts
#### When
```shell
printf 'a b\nb c\n' | tsort
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
a
b
c
```
### Scenario: egrep uses extended regular expressions
#### When
```shell
printf 'foo\nbar\nbaz\n' | egrep 'ba(r|z)'
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
bar
baz
```
### Scenario: fgrep matches fixed strings literally
#### When
```shell
printf 'a.b\naxb\n' | fgrep 'a.b'
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox add-shell
Source: `test/e2e/tools/mimixbox/debianutils/add-shell.atago.yaml`
### Scenario: prints usage with --help and exits 0
#### When
```shell
add-shell --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: add-shell`
### Scenario: fails with a message when given no operand
#### When
```shell
add-shell
```
#### Then
- exit code is not `0`
- stderr contains `add-shell`
## mimixbox ischroot
Source: `test/e2e/tools/mimixbox/debianutils/ischroot.atago.yaml`
### Scenario: prints usage with --help and exits 0
#### When
```shell
ischroot --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: ischroot`
## mimixbox mktemp
Source: `test/e2e/tools/mimixbox/debianutils/mktemp.atago.yaml`
### Scenario: creates a regular file under the temp dir
#### When
```shell
mkdir -p ${workdir}/mktemp
f=$(mktemp -p ${workdir}/mktemp)
test -f "$f" && echo "created"

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: creates a directory
#### When
```shell
mkdir -p ${workdir}/mktemp
d=$(mktemp -d -p ${workdir}/mktemp)
test -d "$d" && echo "created"

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: mktemp -u only prints a name
#### When
```shell
mkdir -p ${workdir}/mktemp
f=$(mktemp -u -p ${workdir}/mktemp)
test ! -e "$f" && echo "not created"

```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox remove-shell
Source: `test/e2e/tools/mimixbox/debianutils/remove-shell.atago.yaml`
### Scenario: prints usage with --help and exits 0
#### When
```shell
remove-shell --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: remove-shell`
### Scenario: fails with a message when given no operand
#### When
```shell
remove-shell
```
#### Then
- exit code is not `0`
- stderr contains `remove-shell`
## mimixbox valid-shell
Source: `test/e2e/tools/mimixbox/debianutils/valid-shell.atago.yaml`
### Scenario: prints usage with --help and exits 0
#### When
```shell
valid-shell --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: valid-shell`
### Scenario: accepts a file listing existing shells
#### Given
- Fixture file `shells.txt` is created.
#### Inputs
_Fixture `shells.txt`:_
```
/bin/sh
/bin/bash
```
#### When
```shell
valid-shell shells.txt
```
#### Then
- exit code is `0`
- stdout contains `OK`
## mimixbox awk
Source: `test/e2e/tools/mimixbox/editors/awk.atago.yaml`
### Scenario: prints a field
#### When
```shell
printf 'one two three\n' | awk '{print $2}'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: honors -F
#### When
```shell
printf 'root:x:0\n' | awk -F: '{print $1}'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: selects a record with NR
#### When
```shell
printf 'a\nb\nc\n' | awk 'NR==2'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: counts records in END
#### When
```shell
printf 'a\nb\nc\n' | awk 'END{print NR}'
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox diff
Source: `test/e2e/tools/mimixbox/editors/diff.atago.yaml`
### Scenario: reports a change in normal format
#### Given
- Fixture file `a` is created.
- Fixture file `b` is created.
#### Inputs
_Fixture `a`:_
```
one
two
three
```
_Fixture `b`:_
```
one
2
three
```
#### When
```shell
diff a b
```
#### Then
- exit code is `1`
- stdout equals an exact value
### Scenario: is silent and succeeds for identical files
#### Given
- Fixture file `a` is created.
- Fixture file `c` is created.
#### Inputs
_Fixture `a`:_
```
one
two
three
```
_Fixture `c`:_
```
one
two
three
```
#### When
```shell
diff a c
```
#### Then
- exit code is `0`
- stdout is empty
### Scenario: reports briefly with -q
#### Given
- Fixture file `a` is created.
- Fixture file `b` is created.
#### Inputs
_Fixture `a`:_
```
one
two
three
```
_Fixture `b`:_
```
one
2
three
```
#### When
```shell
diff -q a b
```
#### Then
- exit code is `1`
- stdout contains `differ`
## mimixbox ed
Source: `test/e2e/tools/mimixbox/editors/ed.atago.yaml`
### Scenario: prints the buffer with size
#### Given
- Fixture file `buf.txt` is created.
#### Inputs
_Fixture `buf.txt`:_
```
one
two
three
```
_stdin for `ed`:_
```
1,$p
q
```
#### When
```shell
ed buf.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
14
one
two
three
```
### Scenario: appends a line and writes it
#### Given
- Fixture file `buf.txt` is created.
#### Inputs
_Fixture `buf.txt`:_
```
one
two
three
```
#### When
```shell
printf '2a\nINSERTED\n.\nw\nq\n' | ed buf.txt > /dev/null
cat buf.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
one
two
INSERTED
three
```
### Scenario: substitutes text on a line
#### Given
- Fixture file `buf.txt` is created.
#### Inputs
_Fixture `buf.txt`:_
```
one
two
three
```
#### When
```shell
printf '2s/two/TWO/\nw\nq\n' | ed buf.txt > /dev/null
cat buf.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
one
TWO
three
```
## mimixbox patch
Source: `test/e2e/tools/mimixbox/editors/patch.atago.yaml`
### Scenario: applies a unified diff
#### Given
- Fixture file `f.txt` is created.
- Fixture file `p.diff` is created.
#### Inputs
_Fixture `f.txt`:_
```
one
two
three
```
_Fixture `p.diff`:_
```
--- ${workdir}/f.txt
+++ ${workdir}/f.txt
@@ -1,3 +1,3 @@
 one
-two
+2
 three
```
#### When
```shell
patch -i p.diff
cat f.txt
```
#### Then
- after `patch -i p.diff`:
  - exit code is `0`
- after `cat f.txt`:
  - exit code is `0`
  - stdout equals an exact value
#### Expected output
_expected stdout:_
```
one
2
three
```
## mimixbox sed
Source: `test/e2e/tools/mimixbox/editors/sed.atago.yaml`
### Scenario: substitutes the first match
#### When
```shell
printf 'hello world\n' | sed 's/world/sed/'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: substitutes globally
#### When
```shell
printf 'a a a\n' | sed 's/a/b/g'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: deletes a line by number
#### When
```shell
printf '1\n2\n3\n' | sed '2d'
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout equals an exact value
### Scenario: prints a single line with -n
#### When
```shell
printf 'x\ny\nz\n' | sed -n '2p'
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox vi
Source: `test/e2e/tools/mimixbox/editors/vi.atago.yaml`
### Scenario: deletes a character and writes the file
#### Given
- Fixture file `a.txt` is created.
#### Inputs
_Fixture `a.txt`:_
```
hello
world
```
#### When
```shell
printf 'x:wq\r' | vi a.txt
cat a.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
ello
world
```
### Scenario: inserts text and writes the file
#### Given
- Fixture file `b.txt` is created.
#### Inputs
_Fixture `b.txt`:_
```
bar
```
#### When
```shell
printf 'ifoo\033:wq\r' | vi b.txt
cat b.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: creates a new file
#### When
```shell
printf 'icreated\033:wq\r' | vi new.txt
cat new.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: treats an arrow-key escape sequence as a motion, not an edit
#### Given
- Fixture file `a.txt` is created.
#### Inputs
_Fixture `a.txt`:_
```
hello
world
```
#### When
```shell
printf '\033[A!\033:wq\r' | vi a.txt
cat a.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
hello
world
```
### Scenario: duplicates a line with yy then p
#### Given
- Fixture file `yp.txt` is created.
#### Inputs
_Fixture `yp.txt`:_
```
one
two
```
#### When
```shell
printf 'yyp:wq\r' | vi yp.txt
cat yp.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
one
one
two
```
### Scenario: applies a count to an edit (2x)
#### Given
- Fixture file `cd.txt` is created.
#### Inputs
_Fixture `cd.txt`:_
```
abcdef
```
#### When
```shell
printf '2x:wq\r' | vi cd.txt
cat cd.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: undoes the last change with u
#### Given
- Fixture file `u.txt` is created.
#### Inputs
_Fixture `u.txt`:_
```
keepme
```
#### When
```shell
printf 'xu:wq\r' | vi u.txt
cat u.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: searches with /pattern and moves to the next match with n
#### Given
- Fixture file `sn.txt` is created.
#### Inputs
_Fixture `sn.txt`:_
```
x
foo
y
foo
z
```
#### When
```shell
printf '/foo\rndd:wq\r' | vi sn.txt
cat sn.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
x
foo
y
z
```
## mimixbox devmem
Source: `test/e2e/tools/mimixbox/embedded/devmem.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
devmem --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: devmem`
- stderr is empty
## mimixbox capability-gated applets
Source: `test/e2e/tools/mimixbox/embedded/gated_plan.atago.yaml`
### Scenario: netctl: brctl addbr prints the plan then fails with a capability-gated backend error
#### When
```shell
brctl addbr br0
```
#### Then
- exit code is not `0`
- stdout contains `brctl: planned action: brctl addbr br0`
- stderr contains `planned action [brctl addbr br0] requires privileged kernel network configuration not available in this environment (capability-gated backend)`
### Scenario: selinux: setenforce refuses to mutate SELinux state and exits non-zero
#### When
```shell
setenforce Permissive
```
#### Then
- exit code is not `0`
- stderr contains `setenforce: refusing to mutate SELinux state: requires CAP_MAC_ADMIN and a loaded policy; this operation is intentionally not implemented in the hermetic build`
### Scenario: modutils: modprobe validates the module then fails on the CAP_SYS_MODULE gate
#### When
```shell
modprobe foo
```
#### Then
- exit code is not `0`
- stderr contains `modprobe: load of foo validated successfully, but inserting/removing kernel modules requires CAP_SYS_MODULE; this privileged step is intentionally not implemented in the hermetic build`
## mimixbox getfattr
Source: `test/e2e/tools/mimixbox/embedded/getfattr.atago.yaml`
### Scenario: dumps a user attribute set by setfattr (or skips without xattr support)
#### Given
- Fixture file `file.txt` is created.
#### When
```shell
if ! setfattr -n user.demo -v hello file.txt 2>/dev/null; then
    echo "user.demo (skipped: filesystem has no xattr support)"
else
    getfattr -d file.txt | grep 'user.demo'
fi

```
#### Then
- exit code is `0`
- stdout contains `user.demo`
### Scenario: fails when no file operand is given
#### When
```shell
getfattr
```
#### Then
- exit code is not `0`
- stderr contains `file operand`
### Scenario: prints usage for --help
#### When
```shell
getfattr --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: getfattr`
### Scenario: prints the version line for --version
#### When
```shell
getfattr --version
```
#### Then
- exit code is `0`
- stdout contains `getfattr (mimixbox)`
## mimixbox compression --help contract
Source: `test/e2e/tools/mimixbox/embedded/help_compression.atago.yaml`
### Scenario: xz --help exposes the documented sections
#### When
```shell
xz --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: xz/`
- stdout contains `Examples:`, `Exit status:`, `  xz `
### Scenario: unxz --help exposes the documented sections
#### When
```shell
unxz --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: unxz/`
- stdout contains `Examples:`, `Exit status:`, `  unxz `
### Scenario: xzcat --help exposes the documented sections
#### When
```shell
xzcat --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: xzcat/`
- stdout contains `Examples:`, `Exit status:`, `  xzcat `
### Scenario: lzma --help exposes the documented sections
#### When
```shell
lzma --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: lzma/`
- stdout contains `Examples:`, `Exit status:`, `  lzma `
### Scenario: unlzma --help exposes the documented sections
#### When
```shell
unlzma --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: unlzma/`
- stdout contains `Examples:`, `Exit status:`, `  unlzma `
### Scenario: lzcat --help exposes the documented sections
#### When
```shell
lzcat --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: lzcat/`
- stdout contains `Examples:`, `Exit status:`, `  lzcat `
### Scenario: lzop --help exposes the documented sections
#### When
```shell
lzop --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: lzop/`
- stdout contains `Examples:`, `Exit status:`, `  lzop `
### Scenario: unlzop --help exposes the documented sections
#### When
```shell
unlzop --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: unlzop/`
- stdout contains `Examples:`, `Exit status:`, `  unlzop `
### Scenario: lzopcat --help exposes the documented sections
#### When
```shell
lzopcat --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: lzopcat/`
- stdout contains `Examples:`, `Exit status:`, `  lzopcat `
### Scenario: zcat --help exposes the documented sections
#### When
```shell
zcat --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: zcat/`
- stdout contains `Examples:`, `Exit status:`, `  zcat `
### Scenario: bzcat --help exposes the documented sections
#### When
```shell
bzcat --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: bzcat/`
- stdout contains `Examples:`, `Exit status:`, `  bzcat `
### Scenario: unit --help exposes the documented sections
#### When
```shell
unit --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: unit/`
- stdout contains `Examples:`, `Exit status:`, `  unit `
## mimixbox --help exit-status contract
Source: `test/e2e/tools/mimixbox/embedded/help_exit_status.atago.yaml`
### Scenario: ash --help exposes the documented sections
#### When
```shell
ash --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: ash/`
- stdout contains `Exit status:`
### Scenario: bash --help exposes the documented sections
#### When
```shell
bash --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: bash/`
- stdout contains `Exit status:`
### Scenario: bc --help exposes the documented sections
#### When
```shell
bc --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: bc/`
- stdout contains `Exit status:`
### Scenario: busybox --help exposes the documented sections
#### When
```shell
busybox --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: busybox/`
- stdout contains `Exit status:`
### Scenario: cttyhack --help exposes the documented sections
#### When
```shell
cttyhack --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: cttyhack/`
- stdout contains `Exit status:`
### Scenario: dc --help exposes the documented sections
#### When
```shell
dc --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: dc/`
- stdout contains `Exit status:`
### Scenario: ed --help exposes the documented sections
#### When
```shell
ed --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: ed/`
- stdout contains `Exit status:`
### Scenario: hd --help exposes the documented sections
#### When
```shell
hd --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: hd/`
- stdout contains `Exit status:`
### Scenario: hexdump --help exposes the documented sections
#### When
```shell
hexdump --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: hexdump/`
- stdout contains `Exit status:`
### Scenario: hush --help exposes the documented sections
#### When
```shell
hush --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: hush/`
- stdout contains `Exit status:`
### Scenario: iostat --help exposes the documented sections
#### When
```shell
iostat --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: iostat/`
- stdout contains `Exit status:`
### Scenario: ipcs --help exposes the documented sections
#### When
```shell
ipcs --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: ipcs/`
- stdout contains `Exit status:`
### Scenario: last --help exposes the documented sections
#### When
```shell
last --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: last/`
- stdout contains `Exit status:`
### Scenario: less --help exposes the documented sections
#### When
```shell
less --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: less/`
- stdout contains `Exit status:`
### Scenario: lsblk --help exposes the documented sections
#### When
```shell
lsblk --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: lsblk/`
- stdout contains `Exit status:`
### Scenario: lspci --help exposes the documented sections
#### When
```shell
lspci --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: lspci/`
- stdout contains `Exit status:`
### Scenario: lsusb --help exposes the documented sections
#### When
```shell
lsusb --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: lsusb/`
- stdout contains `Exit status:`
### Scenario: mbsh --help exposes the documented sections
#### When
```shell
mbsh --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: mbsh/`
- stdout contains `Exit status:`
### Scenario: minips --help exposes the documented sections
#### When
```shell
minips --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: minips/`
- stdout contains `Exit status:`
### Scenario: more --help exposes the documented sections
#### When
```shell
more --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: more/`
- stdout contains `Exit status:`
### Scenario: mpstat --help exposes the documented sections
#### When
```shell
mpstat --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: mpstat/`
- stdout contains `Exit status:`
### Scenario: nmeter --help exposes the documented sections
#### When
```shell
nmeter --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: nmeter/`
- stdout contains `Exit status:`
### Scenario: pipe_progress --help exposes the documented sections
#### When
```shell
pipe_progress --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: pipe_progress/`
- stdout contains `Exit status:`
### Scenario: powertop --help exposes the documented sections
#### When
```shell
powertop --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: powertop/`
- stdout contains `Exit status:`
### Scenario: ps --help exposes the documented sections
#### When
```shell
ps --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: ps/`
- stdout contains `Exit status:`
### Scenario: pstree --help exposes the documented sections
#### When
```shell
pstree --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: pstree/`
- stdout contains `Exit status:`
### Scenario: sh --help exposes the documented sections
#### When
```shell
sh --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: sh/`
- stdout contains `Exit status:`
### Scenario: smemcap --help exposes the documented sections
#### When
```shell
smemcap --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: smemcap/`
- stdout contains `Exit status:`
### Scenario: top --help exposes the documented sections
#### When
```shell
top --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: top/`
- stdout contains `Exit status:`
### Scenario: uptime --help exposes the documented sections
#### When
```shell
uptime --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: uptime/`
- stdout contains `Exit status:`
### Scenario: users --help exposes the documented sections
#### When
```shell
users --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: users/`
- stdout contains `Exit status:`
### Scenario: uudecode --help exposes the documented sections
#### When
```shell
uudecode --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: uudecode/`
- stdout contains `Exit status:`
### Scenario: uuencode --help exposes the documented sections
#### When
```shell
uuencode --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: uuencode/`
- stdout contains `Exit status:`
### Scenario: vi --help exposes the documented sections
#### When
```shell
vi --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: vi/`
- stdout contains `Exit status:`
### Scenario: vmstat --help exposes the documented sections
#### When
```shell
vmstat --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: vmstat/`
- stdout contains `Exit status:`
### Scenario: w --help exposes the documented sections
#### When
```shell
w --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: w/`
- stdout contains `Exit status:`
### Scenario: wall --help exposes the documented sections
#### When
```shell
wall --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: wall/`
- stdout contains `Exit status:`
## mimixbox embedded --help helpers
Source: `test/e2e/tools/mimixbox/embedded/help_helpers_embedded.atago.yaml`
### Scenario: devmem --help is structured
#### When
```shell
devmem --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: i2cdetect --help is structured
#### When
```shell
i2cdetect --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: i2cdump --help is structured
#### When
```shell
i2cdump --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: i2cget --help is structured
#### When
```shell
i2cget --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: i2cset --help is structured
#### When
```shell
i2cset --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: partprobe --help is structured
#### When
```shell
partprobe --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: raidautorun --help is structured
#### When
```shell
raidautorun --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: readahead --help is structured
#### When
```shell
readahead --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: resume --help is structured
#### When
```shell
resume --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: seedrng --help is structured
#### When
```shell
seedrng --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: volname --help is structured
#### When
```shell
volname --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: watchdog --help is structured
#### When
```shell
watchdog --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
## mimixbox --help Notes contract
Source: `test/e2e/tools/mimixbox/embedded/help_notes.atago.yaml`
### Scenario: acpid --help exposes the documented sections
#### When
```shell
acpid --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: acpid/`
- stdout contains `Notes:`
### Scenario: brctl --help exposes the documented sections
#### When
```shell
brctl --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: brctl/`
- stdout contains `Notes:`
### Scenario: crond --help exposes the documented sections
#### When
```shell
crond --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: crond/`
- stdout contains `Notes:`
### Scenario: ifenslave --help exposes the documented sections
#### When
```shell
ifenslave --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: ifenslave/`
- stdout contains `Notes:`
### Scenario: mkfs.reiser --help exposes the documented sections
#### When
```shell
mkfs.reiser --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: mkfs\.reiser/`
- stdout contains `Notes:`
### Scenario: nbd-client --help exposes the documented sections
#### When
```shell
nbd-client --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: nbd\-client/`
- stdout contains `Notes:`
### Scenario: ssl_server --help exposes the documented sections
#### When
```shell
ssl_server --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: ssl_server/`
- stdout contains `Notes:`
### Scenario: tunctl --help exposes the documented sections
#### When
```shell
tunctl --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: tunctl/`
- stdout contains `Notes:`
### Scenario: vconfig --help exposes the documented sections
#### When
```shell
vconfig --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: vconfig/`
- stdout contains `Notes:`
### Scenario: zcip --help exposes the documented sections
#### When
```shell
zcip --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: zcip/`
- stdout contains `Notes:`
### Scenario: [ --help exposes the documented sections
#### When
```shell
env [ --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: \[/`
- stdout contains `Notes:`
### Scenario: [[ --help exposes the documented sections
#### When
```shell
env [[ --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: \[\[/`
- stdout contains `Notes:`
## mimixbox structured --help sections
Source: `test/e2e/tools/mimixbox/embedded/help_structured_sections.atago.yaml`
### Scenario: ln --help exposes the documented sections
#### When
```shell
ln --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: ln/`
- stdout contains `Examples:`, `Exit status:`, `  ln `
### Scenario: log-collect --help exposes the documented sections
#### When
```shell
log-collect --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: log\-collect/`
- stdout contains `Examples:`, `Exit status:`, `  log-collect `
### Scenario: logname --help exposes the documented sections
#### When
```shell
logname --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: logname/`
- stdout contains `Examples:`, `Exit status:`, `  logname `
### Scenario: md5sum --help exposes the documented sections
#### When
```shell
md5sum --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: md5sum/`
- stdout contains `Examples:`, `Exit status:`, `  md5sum `
### Scenario: mkdir --help exposes the documented sections
#### When
```shell
mkdir --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: mkdir/`
- stdout contains `Examples:`, `Exit status:`, `  mkdir `
### Scenario: mkfifo --help exposes the documented sections
#### When
```shell
mkfifo --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: mkfifo/`
- stdout contains `Examples:`, `Exit status:`, `  mkfifo `
### Scenario: mknod --help exposes the documented sections
#### When
```shell
mknod --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: mknod/`
- stdout contains `Examples:`, `Exit status:`, `  mknod `
### Scenario: mktemp --help exposes the documented sections
#### When
```shell
mktemp --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: mktemp/`
- stdout contains `Examples:`, `Exit status:`, `  mktemp `
### Scenario: mountpoint --help exposes the documented sections
#### When
```shell
mountpoint --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: mountpoint/`
- stdout contains `Examples:`, `Exit status:`, `  mountpoint `
### Scenario: mv --help exposes the documented sections
#### When
```shell
mv --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: mv/`
- stdout contains `Examples:`, `Exit status:`, `  mv `
### Scenario: nc --help exposes the documented sections
#### When
```shell
nc --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: nc/`
- stdout contains `Examples:`, `Exit status:`, `  nc `
### Scenario: netcat --help exposes the documented sections
#### When
```shell
netcat --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: netcat/`
- stdout contains `Examples:`, `Exit status:`, `  netcat `
### Scenario: nl --help exposes the documented sections
#### When
```shell
nl --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: nl/`
- stdout contains `Examples:`, `Exit status:`, `  nl `
### Scenario: nohup --help exposes the documented sections
#### When
```shell
nohup --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: nohup/`
- stdout contains `Examples:`, `Exit status:`, `  nohup `
### Scenario: nproc --help exposes the documented sections
#### When
```shell
nproc --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: nproc/`
- stdout contains `Examples:`, `Exit status:`, `  nproc `
### Scenario: nyancat --help exposes the documented sections
#### When
```shell
nyancat --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: nyancat/`
- stdout contains `Examples:`, `Exit status:`, `  nyancat `
### Scenario: od --help exposes the documented sections
#### When
```shell
od --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: od/`
- stdout contains `Examples:`, `Exit status:`, `  od `
### Scenario: paste --help exposes the documented sections
#### When
```shell
paste --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: paste/`
- stdout contains `Examples:`, `Exit status:`, `  paste `
### Scenario: patch --help exposes the documented sections
#### When
```shell
patch --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: patch/`
- stdout contains `Examples:`, `Exit status:`, `  patch `
### Scenario: path --help exposes the documented sections
#### When
```shell
path --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: path/`
- stdout contains `Examples:`, `Exit status:`, `  path `
### Scenario: pidof --help exposes the documented sections
#### When
```shell
pidof --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: pidof/`
- stdout contains `Examples:`, `Exit status:`, `  pidof `
### Scenario: ping --help exposes the documented sections
#### When
```shell
ping --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: ping/`
- stdout contains `Examples:`, `Exit status:`, `  ping `
### Scenario: posixer --help exposes the documented sections
#### When
```shell
posixer --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: posixer/`
- stdout contains `Examples:`, `Exit status:`, `  posixer `
### Scenario: poweroff --help exposes the documented sections
#### When
```shell
poweroff --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: poweroff/`
- stdout contains `Examples:`, `Exit status:`, `  poweroff `
### Scenario: printenv --help exposes the documented sections
#### When
```shell
printenv --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: printenv/`
- stdout contains `Examples:`, `Exit status:`, `  printenv `
### Scenario: pwcrack --help exposes the documented sections
#### When
```shell
pwcrack --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: pwcrack/`
- stdout contains `Examples:`, `Exit status:`, `  pwcrack `
### Scenario: pwgen --help exposes the documented sections
#### When
```shell
pwgen --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: pwgen/`
- stdout contains `Examples:`, `Exit status:`, `  pwgen `
### Scenario: pwscore --help exposes the documented sections
#### When
```shell
pwscore --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: pwscore/`
- stdout contains `Examples:`, `Exit status:`, `  pwscore `
### Scenario: readlink --help exposes the documented sections
#### When
```shell
readlink --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: readlink/`
- stdout contains `Examples:`, `Exit status:`, `  readlink `
### Scenario: realpath --help exposes the documented sections
#### When
```shell
realpath --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: realpath/`
- stdout contains `Examples:`, `Exit status:`, `  realpath `
### Scenario: reboot --help exposes the documented sections
#### When
```shell
reboot --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: reboot/`
- stdout contains `Examples:`, `Exit status:`, `  reboot `
### Scenario: remove-shell --help exposes the documented sections
#### When
```shell
remove-shell --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: remove\-shell/`
- stdout contains `Examples:`, `Exit status:`, `  remove-shell `
### Scenario: reset --help exposes the documented sections
#### When
```shell
reset --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: reset/`
- stdout contains `Examples:`, `Exit status:`, `  reset `
### Scenario: resize --help exposes the documented sections
#### When
```shell
resize --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: resize/`
- stdout contains `Examples:`, `Exit status:`, `  resize `
### Scenario: rev --help exposes the documented sections
#### When
```shell
rev --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: rev/`
- stdout contains `Examples:`, `Exit status:`, `  rev `
### Scenario: rm --help exposes the documented sections
#### When
```shell
rm --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: rm/`
- stdout contains `Examples:`, `Exit status:`, `  rm `
### Scenario: rmdir --help exposes the documented sections
#### When
```shell
rmdir --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: rmdir/`
- stdout contains `Examples:`, `Exit status:`, `  rmdir `
### Scenario: rpm --help exposes the documented sections
#### When
```shell
rpm --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: rpm/`
- stdout contains `Examples:`, `Exit status:`, `  rpm `
### Scenario: rpm2cpio --help exposes the documented sections
#### When
```shell
rpm2cpio --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: rpm2cpio/`
- stdout contains `Examples:`, `Exit status:`, `  rpm2cpio `
### Scenario: sddf --help exposes the documented sections
#### When
```shell
sddf --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: sddf/`
- stdout contains `Examples:`, `Exit status:`, `  sddf `
### Scenario: sed --help exposes the documented sections
#### When
```shell
sed --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: sed/`
- stdout contains `Examples:`, `Exit status:`, `  sed `
### Scenario: seq --help exposes the documented sections
#### When
```shell
seq --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: seq/`
- stdout contains `Examples:`, `Exit status:`, `  seq `
### Scenario: serial --help exposes the documented sections
#### When
```shell
serial --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: serial/`
- stdout contains `Examples:`, `Exit status:`, `  serial `
### Scenario: sha1sum --help exposes the documented sections
#### When
```shell
sha1sum --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: sha1sum/`
- stdout contains `Examples:`, `Exit status:`, `  sha1sum `
### Scenario: sha256sum --help exposes the documented sections
#### When
```shell
sha256sum --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: sha256sum/`
- stdout contains `Examples:`, `Exit status:`, `  sha256sum `
### Scenario: sha384sum --help exposes the documented sections
#### When
```shell
sha384sum --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: sha384sum/`
- stdout contains `Examples:`, `Exit status:`, `  sha384sum `
### Scenario: sha3sum --help exposes the documented sections
#### When
```shell
sha3sum --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: sha3sum/`
- stdout contains `Examples:`, `Exit status:`, `  sha3sum `
### Scenario: sha512sum --help exposes the documented sections
#### When
```shell
sha512sum --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: sha512sum/`
- stdout contains `Examples:`, `Exit status:`, `  sha512sum `
### Scenario: shred --help exposes the documented sections
#### When
```shell
shred --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: shred/`
- stdout contains `Examples:`, `Exit status:`, `  shred `
### Scenario: shuf --help exposes the documented sections
#### When
```shell
shuf --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: shuf/`
- stdout contains `Examples:`, `Exit status:`, `  shuf `
### Scenario: sl --help exposes the documented sections
#### When
```shell
sl --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: sl/`
- stdout contains `Examples:`, `Exit status:`, `  sl `
### Scenario: sleep --help exposes the documented sections
#### When
```shell
sleep --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: sleep/`
- stdout contains `Examples:`, `Exit status:`, `  sleep `
### Scenario: sort --help exposes the documented sections
#### When
```shell
sort --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: sort/`
- stdout contains `Examples:`, `Exit status:`, `  sort `
### Scenario: speaker --help exposes the documented sections
#### When
```shell
speaker --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: speaker/`
- stdout contains `Examples:`, `Exit status:`, `  speaker `
### Scenario: split --help exposes the documented sections
#### When
```shell
split --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: split/`
- stdout contains `Examples:`, `Exit status:`, `  split `
### Scenario: stat --help exposes the documented sections
#### When
```shell
stat --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: stat/`
- stdout contains `Examples:`, `Exit status:`, `  stat `
### Scenario: strings --help exposes the documented sections
#### When
```shell
strings --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: strings/`
- stdout contains `Examples:`, `Exit status:`, `  strings `
### Scenario: sync --help exposes the documented sections
#### When
```shell
sync --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: sync/`
- stdout contains `Examples:`, `Exit status:`, `  sync `
### Scenario: tac --help exposes the documented sections
#### When
```shell
tac --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: tac/`
- stdout contains `Examples:`, `Exit status:`, `  tac `
### Scenario: tar --help exposes the documented sections
#### When
```shell
tar --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: tar/`
- stdout contains `Examples:`, `Exit status:`, `  tar `
### Scenario: tee --help exposes the documented sections
#### When
```shell
tee --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: tee/`
- stdout contains `Examples:`, `Exit status:`, `  tee `
### Scenario: timeout --help exposes the documented sections
#### When
```shell
timeout --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: timeout/`
- stdout contains `Examples:`, `Exit status:`, `  timeout `
### Scenario: touch --help exposes the documented sections
#### When
```shell
touch --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: touch/`
- stdout contains `Examples:`, `Exit status:`, `  touch `
### Scenario: tr --help exposes the documented sections
#### When
```shell
tr --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: tr/`
- stdout contains `Examples:`, `Exit status:`, `  tr `
### Scenario: truncate --help exposes the documented sections
#### When
```shell
truncate --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: truncate/`
- stdout contains `Examples:`, `Exit status:`, `  truncate `
### Scenario: tty --help exposes the documented sections
#### When
```shell
tty --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: tty/`
- stdout contains `Examples:`, `Exit status:`, `  tty `
### Scenario: uname --help exposes the documented sections
#### When
```shell
uname --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: uname/`
- stdout contains `Examples:`, `Exit status:`, `  uname `
### Scenario: uncompress --help exposes the documented sections
#### When
```shell
uncompress --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: uncompress/`
- stdout contains `Examples:`, `Exit status:`, `  uncompress `
### Scenario: unexpand --help exposes the documented sections
#### When
```shell
unexpand --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: unexpand/`
- stdout contains `Examples:`, `Exit status:`, `  unexpand `
### Scenario: uniq --help exposes the documented sections
#### When
```shell
uniq --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: uniq/`
- stdout contains `Examples:`, `Exit status:`, `  uniq `
### Scenario: unix2dos --help exposes the documented sections
#### When
```shell
unix2dos --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: unix2dos/`
- stdout contains `Examples:`, `Exit status:`, `  unix2dos `
### Scenario: unlink --help exposes the documented sections
#### When
```shell
unlink --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: unlink/`
- stdout contains `Examples:`, `Exit status:`, `  unlink `
### Scenario: unshadow --help exposes the documented sections
#### When
```shell
unshadow --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: unshadow/`
- stdout contains `Examples:`, `Exit status:`, `  unshadow `
### Scenario: unzip --help exposes the documented sections
#### When
```shell
unzip --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: unzip/`
- stdout contains `Examples:`, `Exit status:`, `  unzip `
### Scenario: uuidgen --help exposes the documented sections
#### When
```shell
uuidgen --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: uuidgen/`
- stdout contains `Examples:`, `Exit status:`, `  uuidgen `
### Scenario: valid-shell --help exposes the documented sections
#### When
```shell
valid-shell --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: valid\-shell/`
- stdout contains `Examples:`, `Exit status:`, `  valid-shell `
### Scenario: watch --help exposes the documented sections
#### When
```shell
watch --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: watch/`
- stdout contains `Examples:`, `Exit status:`, `  watch `
### Scenario: wc --help exposes the documented sections
#### When
```shell
wc --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: wc/`
- stdout contains `Examples:`, `Exit status:`, `  wc `
### Scenario: which --help exposes the documented sections
#### When
```shell
which --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: which/`
- stdout contains `Examples:`, `Exit status:`, `  which `
### Scenario: who --help exposes the documented sections
#### When
```shell
who --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: who/`
- stdout contains `Examples:`, `Exit status:`, `  who `
### Scenario: whoami --help exposes the documented sections
#### When
```shell
whoami --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: whoami/`
- stdout contains `Examples:`, `Exit status:`, `  whoami `
### Scenario: whris --help exposes the documented sections
#### When
```shell
whris --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: whris/`
- stdout contains `Examples:`, `Exit status:`, `  whris `
### Scenario: xargs --help exposes the documented sections
#### When
```shell
xargs --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: xargs/`
- stdout contains `Examples:`, `Exit status:`, `  xargs `
### Scenario: xxd --help exposes the documented sections
#### When
```shell
xxd --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: xxd/`
- stdout contains `Examples:`, `Exit status:`, `  xxd `
### Scenario: yes --help exposes the documented sections
#### When
```shell
yes --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: yes/`
- stdout contains `Examples:`, `Exit status:`, `  yes `
### Scenario: zip --help exposes the documented sections
#### When
```shell
zip --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: zip/`
- stdout contains `Examples:`, `Exit status:`, `  zip `
### Scenario: zip-pwcrack --help exposes the documented sections
#### When
```shell
zip-pwcrack --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: zip\-pwcrack/`
- stdout contains `Examples:`, `Exit status:`, `  zip-pwcrack `
### Scenario: true --help exposes the documented sections
#### When
```shell
env true --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: true/`
- stdout contains `Examples:`, `Exit status:`, `  true `
### Scenario: test --help exposes the documented sections
#### When
```shell
env test --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: test/`
- stdout contains `Examples:`, `Exit status:`, `  test `
### Scenario: printf --help exposes the documented sections
#### When
```shell
env printf --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: printf/`
- stdout contains `Examples:`, `Exit status:`, `  printf `
### Scenario: pwd --help exposes the documented sections
#### When
```shell
env pwd --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: pwd/`
- stdout contains `Examples:`, `Exit status:`, `  pwd `
## mimixbox structured --help sections (2)
Source: `test/e2e/tools/mimixbox/embedded/help_structured_sections_2.atago.yaml`
### Scenario: add-shell --help exposes the documented sections
#### When
```shell
add-shell --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: add\-shell/`
- stdout contains `Examples:`, `Exit status:`, `  add-shell `
### Scenario: ar --help exposes the documented sections
#### When
```shell
ar --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: ar/`
- stdout contains `Examples:`, `Exit status:`, `  ar `
### Scenario: arch --help exposes the documented sections
#### When
```shell
arch --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: arch/`
- stdout contains `Examples:`, `Exit status:`, `  arch `
### Scenario: awk --help exposes the documented sections
#### When
```shell
awk --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: awk/`
- stdout contains `Examples:`, `Exit status:`, `  awk `
### Scenario: banner --help exposes the documented sections
#### When
```shell
banner --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: banner/`
- stdout contains `Examples:`, `Exit status:`, `  banner `
### Scenario: base32 --help exposes the documented sections
#### When
```shell
base32 --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: base32/`
- stdout contains `Examples:`, `Exit status:`, `  base32 `
### Scenario: base64 --help exposes the documented sections
#### When
```shell
base64 --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: base64/`
- stdout contains `Examples:`, `Exit status:`, `  base64 `
### Scenario: basename --help exposes the documented sections
#### When
```shell
basename --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: basename/`
- stdout contains `Examples:`, `Exit status:`, `  basename `
### Scenario: bunzip2 --help exposes the documented sections
#### When
```shell
bunzip2 --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: bunzip2/`
- stdout contains `Examples:`, `Exit status:`, `  bunzip2 `
### Scenario: cal --help exposes the documented sections
#### When
```shell
cal --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: cal/`
- stdout contains `Examples:`, `Exit status:`, `  cal `
### Scenario: cat --help exposes the documented sections
#### When
```shell
cat --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: cat/`
- stdout contains `Examples:`, `Exit status:`, `  cat `
### Scenario: chgrp --help exposes the documented sections
#### When
```shell
chgrp --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: chgrp/`
- stdout contains `Examples:`, `Exit status:`, `  chgrp `
### Scenario: chmod --help exposes the documented sections
#### When
```shell
chmod --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: chmod/`
- stdout contains `Examples:`, `Exit status:`, `  chmod `
### Scenario: chown --help exposes the documented sections
#### When
```shell
chown --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: chown/`
- stdout contains `Examples:`, `Exit status:`, `  chown `
### Scenario: cksum --help exposes the documented sections
#### When
```shell
cksum --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: cksum/`
- stdout contains `Examples:`, `Exit status:`, `  cksum `
### Scenario: clear --help exposes the documented sections
#### When
```shell
clear --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: clear/`
- stdout contains `Examples:`, `Exit status:`, `  clear `
### Scenario: cmatrix --help exposes the documented sections
#### When
```shell
cmatrix --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: cmatrix/`
- stdout contains `Examples:`, `Exit status:`, `  cmatrix `
### Scenario: cmp --help exposes the documented sections
#### When
```shell
cmp --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: cmp/`
- stdout contains `Examples:`, `Exit status:`, `  cmp `
### Scenario: comm --help exposes the documented sections
#### When
```shell
comm --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: comm/`
- stdout contains `Examples:`, `Exit status:`, `  comm `
### Scenario: compress --help exposes the documented sections
#### When
```shell
compress --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: compress/`
- stdout contains `Examples:`, `Exit status:`, `  compress `
### Scenario: cowsay --help exposes the documented sections
#### When
```shell
cowsay --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: cowsay/`
- stdout contains `Examples:`, `Exit status:`, `  cowsay `
### Scenario: cowthink --help exposes the documented sections
#### When
```shell
cowthink --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: cowthink/`
- stdout contains `Examples:`, `Exit status:`, `  cowthink `
### Scenario: cpio --help exposes the documented sections
#### When
```shell
cpio --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: cpio/`
- stdout contains `Examples:`, `Exit status:`, `  cpio `
### Scenario: cut --help exposes the documented sections
#### When
```shell
cut --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: cut/`
- stdout contains `Examples:`, `Exit status:`, `  cut `
### Scenario: date --help exposes the documented sections
#### When
```shell
date --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: date/`
- stdout contains `Examples:`, `Exit status:`, `  date `
### Scenario: dd --help exposes the documented sections
#### When
```shell
dd --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: dd/`
- stdout contains `Examples:`, `Exit status:`, `  dd `
### Scenario: df --help exposes the documented sections
#### When
```shell
df --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: df/`
- stdout contains `Examples:`, `Exit status:`, `  df `
### Scenario: diff --help exposes the documented sections
#### When
```shell
diff --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: diff/`
- stdout contains `Examples:`, `Exit status:`, `  diff `
### Scenario: dirname --help exposes the documented sections
#### When
```shell
dirname --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: dirname/`
- stdout contains `Examples:`, `Exit status:`, `  dirname `
### Scenario: dos2unix --help exposes the documented sections
#### When
```shell
dos2unix --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: dos2unix/`
- stdout contains `Examples:`, `Exit status:`, `  dos2unix `
### Scenario: du --help exposes the documented sections
#### When
```shell
du --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: du/`
- stdout contains `Examples:`, `Exit status:`, `  du `
### Scenario: egrep --help exposes the documented sections
#### When
```shell
egrep --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: egrep/`
- stdout contains `Examples:`, `Exit status:`, `  egrep `
### Scenario: env --help exposes the documented sections
#### When
```shell
env --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: env/`
- stdout contains `Examples:`, `Exit status:`, `  env `
### Scenario: expand --help exposes the documented sections
#### When
```shell
expand --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: expand/`
- stdout contains `Examples:`, `Exit status:`, `  expand `
### Scenario: expr --help exposes the documented sections
#### When
```shell
expr --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: expr/`
- stdout contains `Examples:`, `Exit status:`, `  expr `
### Scenario: fakemovie --help exposes the documented sections
#### When
```shell
fakemovie --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: fakemovie/`
- stdout contains `Examples:`, `Exit status:`, `  fakemovie `
### Scenario: fgrep --help exposes the documented sections
#### When
```shell
fgrep --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: fgrep/`
- stdout contains `Examples:`, `Exit status:`, `  fgrep `
### Scenario: fmt --help exposes the documented sections
#### When
```shell
fmt --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: fmt/`
- stdout contains `Examples:`, `Exit status:`, `  fmt `
### Scenario: fold --help exposes the documented sections
#### When
```shell
fold --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: fold/`
- stdout contains `Examples:`, `Exit status:`, `  fold `
### Scenario: fortune --help exposes the documented sections
#### When
```shell
fortune --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: fortune/`
- stdout contains `Examples:`, `Exit status:`, `  fortune `
### Scenario: free --help exposes the documented sections
#### When
```shell
free --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: free/`
- stdout contains `Examples:`, `Exit status:`, `  free `
### Scenario: ghrdc --help exposes the documented sections
#### When
```shell
ghrdc --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: ghrdc/`
- stdout contains `Examples:`, `Exit status:`, `  ghrdc `
### Scenario: grep --help exposes the documented sections
#### When
```shell
grep --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: grep/`
- stdout contains `Examples:`, `Exit status:`, `  grep `
### Scenario: groups --help exposes the documented sections
#### When
```shell
groups --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: groups/`
- stdout contains `Examples:`, `Exit status:`, `  groups `
### Scenario: gunzip --help exposes the documented sections
#### When
```shell
gunzip --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: gunzip/`
- stdout contains `Examples:`, `Exit status:`, `  gunzip `
### Scenario: gzip --help exposes the documented sections
#### When
```shell
gzip --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: gzip/`
- stdout contains `Examples:`, `Exit status:`, `  gzip `
### Scenario: halt --help exposes the documented sections
#### When
```shell
halt --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: halt/`
- stdout contains `Examples:`, `Exit status:`, `  halt `
### Scenario: head --help exposes the documented sections
#### When
```shell
head --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: head/`
- stdout contains `Examples:`, `Exit status:`, `  head `
### Scenario: hostid --help exposes the documented sections
#### When
```shell
hostid --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: hostid/`
- stdout contains `Examples:`, `Exit status:`, `  hostid `
### Scenario: hostname --help exposes the documented sections
#### When
```shell
hostname --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: hostname/`
- stdout contains `Examples:`, `Exit status:`, `  hostname `
### Scenario: http-status-code --help exposes the documented sections
#### When
```shell
http-status-code --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: http\-status\-code/`
- stdout contains `Examples:`, `Exit status:`, `  http-status-code `
### Scenario: id --help exposes the documented sections
#### When
```shell
id --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: id/`
- stdout contains `Examples:`, `Exit status:`, `  id `
### Scenario: install --help exposes the documented sections
#### When
```shell
install --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: install/`
- stdout contains `Examples:`, `Exit status:`, `  install `
### Scenario: ischroot --help exposes the documented sections
#### When
```shell
ischroot --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: ischroot/`
- stdout contains `Examples:`, `Exit status:`, `  ischroot `
### Scenario: killall --help exposes the documented sections
#### When
```shell
killall --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: killall/`
- stdout contains `Examples:`, `Exit status:`, `  killall `
### Scenario: lifegame --help exposes the documented sections
#### When
```shell
lifegame --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: lifegame/`
- stdout contains `Examples:`, `Exit status:`, `  lifegame `
### Scenario: link --help exposes the documented sections
#### When
```shell
link --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: link/`
- stdout contains `Examples:`, `Exit status:`, `  link `
### Scenario: echo --help exposes the documented sections
#### When
```shell
env echo --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: echo/`
- stdout contains `Examples:`, `Exit status:`, `  echo `
### Scenario: false --help exposes the documented sections
#### When
```shell
env false --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: false/`
- stdout contains `Examples:`, `Exit status:`, `  false `
### Scenario: kill --help exposes the documented sections
#### When
```shell
env kill --help
```
#### Then
- exit code is `0`
- stdout matches `/^Usage: kill/`
- stdout contains `Examples:`, `Exit status:`, `  kill `
## mimixbox i2cdetect
Source: `test/e2e/tools/mimixbox/embedded/i2cdetect.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
i2cdetect --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: i2cdetect`
- stderr is empty
## mimixbox i2cdump
Source: `test/e2e/tools/mimixbox/embedded/i2cdump.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
i2cdump --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: i2cdump`
- stderr is empty
## mimixbox i2cget
Source: `test/e2e/tools/mimixbox/embedded/i2cget.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
i2cget --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: i2cget`
- stderr is empty
## mimixbox i2cset
Source: `test/e2e/tools/mimixbox/embedded/i2cset.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
i2cset --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: i2cset`
- stderr is empty
## mimixbox ifup
Source: `test/e2e/tools/mimixbox/embedded/ifup.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
ifup --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: ifup`
- stderr is empty
## mimixbox insmod
Source: `test/e2e/tools/mimixbox/embedded/insmod.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
insmod --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: insmod`
- stderr is empty
## mimixbox ip
Source: `test/e2e/tools/mimixbox/embedded/ip.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
ip --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: ip`
- stderr is empty
## mimixbox ipaddr
Source: `test/e2e/tools/mimixbox/embedded/ipaddr.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
ipaddr --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: ipaddr`
- stderr is empty
## mimixbox iplink
Source: `test/e2e/tools/mimixbox/embedded/iplink.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
iplink --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: iplink`
- stderr is empty
## mimixbox ipneigh
Source: `test/e2e/tools/mimixbox/embedded/ipneigh.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
ipneigh --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: ipneigh`
- stderr is empty
## mimixbox iproute
Source: `test/e2e/tools/mimixbox/embedded/iproute.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
iproute --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: iproute`
- stderr is empty
## mimixbox iprule
Source: `test/e2e/tools/mimixbox/embedded/iprule.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
iprule --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: iprule`
- stderr is empty
## mimixbox iptunnel
Source: `test/e2e/tools/mimixbox/embedded/iptunnel.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
iptunnel --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: iptunnel`
- stderr is empty
## mimixbox less
Source: `test/e2e/tools/mimixbox/embedded/less.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
less --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: less`
- stderr is empty
## mimixbox linux32
Source: `test/e2e/tools/mimixbox/embedded/linux32.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
linux32 --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: linux32`
- stderr is empty
## mimixbox linux64
Source: `test/e2e/tools/mimixbox/embedded/linux64.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
linux64 --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: linux64`
- stderr is empty
## mimixbox linuxrc
Source: `test/e2e/tools/mimixbox/embedded/linuxrc.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
linuxrc --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: linuxrc`
- stderr is empty
## mimixbox load_policy
Source: `test/e2e/tools/mimixbox/embedded/load_policy.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
load_policy --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: load_policy`
- stderr is empty
## mimixbox log-collect
Source: `test/e2e/tools/mimixbox/embedded/log-collect.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
log-collect --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: log-collect`
- stderr is empty
## mimixbox lpd
Source: `test/e2e/tools/mimixbox/embedded/lpd.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
lpd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: lpd`
- stderr is empty
## mimixbox lpq
Source: `test/e2e/tools/mimixbox/embedded/lpq.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
lpq --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: lpq`
- stderr is empty
## mimixbox lpr
Source: `test/e2e/tools/mimixbox/embedded/lpr.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
lpr --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: lpr`
- stderr is empty
## mimixbox lpr_roundtrip
Source: `test/e2e/tools/mimixbox/embedded/lpr_roundtrip.atago.yaml`
### Scenario: queues, lists, drains, and empties the spool
#### When
```shell
printf 'page content\n' > document.txt

# Queue a file and a stdin job.
lpr -S spool document.txt || exit 1
printf 'from stdin\n' | lpr -S spool || exit 1

# lpq must list both jobs in id order.
lpq -S spool | grep -q 'document.txt' || exit 1
lpq -S spool | grep -q '(stdin)' || exit 1

# Drain to an output directory; both jobs are reported printed.
lpd -S spool -o printed | grep -q 'printed job 1' || exit 1
lpd_out=$(lpd -S spool -o printed) # second drain run is a no-op
[ -z "${lpd_out}" ] || exit 1

# The printed file content survives the round-trip.
cat printed/0001-document.txt

# The queue is now empty.
lpq -S spool

```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
page content
no entries
```
## mimixbox lsscsi
Source: `test/e2e/tools/mimixbox/embedded/lsscsi.atago.yaml`
### Scenario: lists SCSI devices from sysfs without error (empty is allowed)
#### When
```shell
if [ -d /sys/bus/scsi/devices ]; then
    lsscsi >/dev/null && echo ok
else
    echo ok
fi

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: prints usage for --help
#### When
```shell
lsscsi --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: lsscsi`
### Scenario: prints the version line for --version
#### When
```shell
lsscsi --version
```
#### Then
- exit code is `0`
- stdout contains `lsscsi (mimixbox)`
## mimixbox lzcat
Source: `test/e2e/tools/mimixbox/embedded/lzcat.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
lzcat --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: lzcat`
- stderr is empty
## mimixbox lzma
Source: `test/e2e/tools/mimixbox/embedded/lzma.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
lzma --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: lzma`
- stderr is empty
## mimixbox lzop
Source: `test/e2e/tools/mimixbox/embedded/lzop.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
lzop --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: lzop`
- stderr is empty
## mimixbox lzopcat
Source: `test/e2e/tools/mimixbox/embedded/lzopcat.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
lzopcat --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: lzopcat`
- stderr is empty
## mimixbox makedevs
Source: `test/e2e/tools/mimixbox/embedded/makedevs.atago.yaml`
### Scenario: creates the directory and file tree from a device table
#### Given
- Fixture file `table.txt` is created.
#### Inputs
_Fixture `table.txt`:_
```
# device table
/dev d 755 0 0 0 0 0 0 0
/etc/hostname f 644 0 0 0 0 0 0 0
```
#### When
```shell
makedevs -d table.txt rootfs
if [ -d rootfs/dev ] && [ -f rootfs/etc/hostname ]; then
    echo 1
else
    echo 0
fi

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: fails without the -d table option
#### When
```shell
makedevs ./rootfs
```
#### Then
- exit code is not `0`
- stderr contains `usage: makedevs`
### Scenario: prints usage for --help
#### When
```shell
makedevs --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: makedevs`
### Scenario: prints the version line for --version
#### When
```shell
makedevs --version
```
#### Then
- exit code is `0`
- stdout contains `makedevs (mimixbox)`
## mimixbox matchpathcon
Source: `test/e2e/tools/mimixbox/embedded/matchpathcon.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
matchpathcon --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: matchpathcon`
- stderr is empty
## mimixbox mkdosfs
Source: `test/e2e/tools/mimixbox/embedded/mkdosfs.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
mkdosfs --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: mkdosfs`
- stderr is empty
## mimixbox mkfs.ext2
Source: `test/e2e/tools/mimixbox/embedded/mkfs.ext2.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
mkfs.ext2 --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: mkfs.ext2`
- stderr is empty
## mimixbox mkfs.minix
Source: `test/e2e/tools/mimixbox/embedded/mkfs.minix.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
mkfs.minix --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: mkfs.minix`
- stderr is empty
## mimixbox mkfs.reiser
Source: `test/e2e/tools/mimixbox/embedded/mkfs.reiser.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
mkfs.reiser --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: mkfs.reiser`
- stderr is empty
## mimixbox mkfs.vfat
Source: `test/e2e/tools/mimixbox/embedded/mkfs.vfat.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
mkfs.vfat --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: mkfs.vfat`
- stderr is empty
## mimixbox modprobe
Source: `test/e2e/tools/mimixbox/embedded/modprobe.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
modprobe --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: modprobe`
- stderr is empty
## mimixbox more
Source: `test/e2e/tools/mimixbox/embedded/more.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
more --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: more`
- stderr is empty
## mimixbox nameif
Source: `test/e2e/tools/mimixbox/embedded/nameif.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
nameif --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: nameif`
- stderr is empty
## mimixbox nbd-client
Source: `test/e2e/tools/mimixbox/embedded/nbd-client.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
nbd-client --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: nbd-client`
- stderr is empty
## mimixbox partprobe
Source: `test/e2e/tools/mimixbox/embedded/partprobe.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
partprobe --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: partprobe`
- stderr is empty
## mimixbox ping6
Source: `test/e2e/tools/mimixbox/embedded/ping6.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
ping6 --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: ping6`
- stderr is empty
## mimixbox pipe_progress
Source: `test/e2e/tools/mimixbox/embedded/pipe_progress.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
pipe_progress --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: pipe_progress`
- stderr is empty
## mimixbox pkill
Source: `test/e2e/tools/mimixbox/embedded/pkill.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
pkill --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: pkill`
- stderr is empty
## mimixbox poweroff
Source: `test/e2e/tools/mimixbox/embedded/poweroff.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
poweroff --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: poweroff`
- stderr is empty
## mimixbox preexisting_tmp_root
Source: `test/e2e/tools/mimixbox/embedded/preexisting_tmp_root.atago.yaml`
### Scenario: allocates a usable per-run root that is not /tmp/mimixbox
#### When
```shell
[ -n "${workdir}" ] || exit 1
[ "${workdir}" != "/tmp/mimixbox" ] || exit 1
mkdir -p "${workdir}/it478" || exit 1
touch "${workdir}/it478/probe" || exit 1
printf 'ok'

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: leaves a pre-existing /tmp/mimixbox file untouched (harness-specific)
#### When
```shell
[ ! -d /tmp/mimixbox ] && printf 'not-a-dir'

```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox raidautorun
Source: `test/e2e/tools/mimixbox/embedded/raidautorun.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
raidautorun --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: raidautorun`
- stderr is empty
## mimixbox readahead
Source: `test/e2e/tools/mimixbox/embedded/readahead.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
readahead --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: readahead`
- stderr is empty
## mimixbox reboot
Source: `test/e2e/tools/mimixbox/embedded/reboot.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
reboot --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: reboot`
- stderr is empty
## mimixbox restorecon
Source: `test/e2e/tools/mimixbox/embedded/restorecon.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
restorecon --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: restorecon`
- stderr is empty
## mimixbox resume
Source: `test/e2e/tools/mimixbox/embedded/resume.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
resume --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: resume`
- stderr is empty
## mimixbox seedrng
Source: `test/e2e/tools/mimixbox/embedded/seedrng.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
seedrng --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: seedrng`
- stderr is empty
### Scenario: documents its purpose in --help
#### When
```shell
seedrng --help
```
#### Then
- exit code is `0`
- stdout contains `boot seed`
## mimixbox setfattr
Source: `test/e2e/tools/mimixbox/embedded/setfattr.atago.yaml`
### Scenario: sets an attribute that getfattr can read back (or skips without xattr support)
#### Given
- Fixture file `file.txt` is created.
#### When
```shell
if ! setfattr -n user.k -v v file.txt 2>/dev/null; then
    echo 'user.k="v" (skipped: filesystem has no xattr support)'
else
    getfattr -d file.txt | grep 'user.k'
fi

```
#### Then
- exit code is `0`
- stdout contains `user.k`
### Scenario: rejects mutually exclusive -n and -x
#### Given
- Fixture file `file.txt` is created.
#### When
```shell
setfattr -n user.k -x user.k file.txt
```
#### Then
- exit code is not `0`
- stderr contains `mutually exclusive`
### Scenario: prints usage for --help
#### When
```shell
setfattr --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: setfattr`
### Scenario: prints the version line for --version
#### When
```shell
setfattr --version
```
#### Then
- exit code is `0`
- stdout contains `setfattr (mimixbox)`
## mimixbox volname
Source: `test/e2e/tools/mimixbox/embedded/volname.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
volname --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: volname`
- stderr is empty
## mimixbox watchdog
Source: `test/e2e/tools/mimixbox/embedded/watchdog.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
watchdog --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: watchdog`
- stderr is empty
## mimixbox chgrp
Source: `test/e2e/tools/mimixbox/fileutils/chgrp.atago.yaml`
### Scenario: prints usage with --help and exits 0
#### When
```shell
chgrp --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: chgrp`
### Scenario: fails with a message when given no operand
#### When
```shell
chgrp
```
#### Then
- exit code is not `0`
- stderr contains `chgrp`
## mimixbox chown
Source: `test/e2e/tools/mimixbox/fileutils/chown.atago.yaml`
### Scenario: prints usage with --help and exits 0
#### When
```shell
chown --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: chown`
### Scenario: fails with a message when given no operand
#### When
```shell
chown
```
#### Then
- exit code is not `0`
- stderr contains `chown`
## mimixbox cp
Source: `test/e2e/tools/mimixbox/fileutils/cp.atago.yaml`
### Scenario: copy one file
#### When
```shell
mkdir -p cp/inner cp2 && touch cp/1.txt cp/2.txt cp/3.txt cp/inner/inner.txt
cp ${workdir}/cp/1.txt ${workdir}/cp/cp.txt && ls ${workdir}/cp/cp.txt
```
#### Then
- after `cp ${workdir}/cp/1.txt ${workdir}/cp/cp.txt && ls ${workdir}/cp/cp.txt`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: copy directory recursively
#### When
```shell
mkdir -p cp/inner cp2 && touch cp/1.txt cp/2.txt cp/3.txt cp/inner/inner.txt
cp -r ${workdir}/cp ${workdir}/cp2 && ls ${workdir}/cp2 && ls ${workdir}/cp2/cp
```
#### Then
- after `cp -r ${workdir}/cp ${workdir}/cp2 && ls ${workdir}/cp2 && ls ${workdir}/cp2/cp`:
  - exit code is `0`
  - stdout equals an exact value
#### Expected output
_expected stdout:_
```
cp
1.txt
2.txt
3.txt
inner
```
### Scenario: can not copy when src and dest are the same
#### When
```shell
mkdir -p cp/inner cp2 && touch cp/1.txt cp/2.txt cp/3.txt cp/inner/inner.txt
cp -r ${workdir}/cp ${workdir}/cp; ls ${workdir}/cp
```
#### Then
- after `cp -r ${workdir}/cp ${workdir}/cp; ls ${workdir}/cp`:
  - stdout equals an exact value
  - stderr equals an exact value
#### Expected output
_expected stdout:_
```
1.txt
2.txt
3.txt
inner
```
### Scenario: status failure when src and dest are the same
#### When
```shell
mkdir -p cp/inner cp2 && touch cp/1.txt cp/2.txt cp/3.txt cp/inner/inner.txt
cp -r ${workdir}/cp ${workdir}/cp
```
#### Then
- after `cp -r ${workdir}/cp ${workdir}/cp`:
  - exit code is not `0`
  - stderr equals an exact value
### Scenario: copy three files at the same time
#### When
```shell
mkdir -p cp/inner cp2 && touch cp/1.txt cp/2.txt cp/3.txt cp/inner/inner.txt
cp ${workdir}/cp/1.txt ${workdir}/cp/2.txt ${workdir}/cp/3.txt ${workdir}/cp2 && ls ${workdir}/cp2
```
#### Then
- after `cp ${workdir}/cp/1.txt ${workdir}/cp/2.txt ${workdir}/cp/3.txt ${workdir}/cp2 && ls ${workdir}/cp2`:
  - exit code is `0`
  - stdout equals an exact value
#### Expected output
_expected stdout:_
```
1.txt
2.txt
3.txt
```
### Scenario: can not copy a directory without the recursive option
#### When
```shell
mkdir -p cp/inner cp2 && touch cp/1.txt cp/2.txt cp/3.txt cp/inner/inner.txt
cp ${workdir}/cp ${workdir}/cp2
```
#### Then
- after `cp ${workdir}/cp ${workdir}/cp2`:
  - exit code is not `0`
  - stderr equals an exact value
#### Expected output
_expected stderr:_
```
cp: --recursive is not specified: omitting directory: ${workdir}/cp
```
### Scenario: can not copy a directory to root without authority
_skipped on linux_
#### When
```shell
cp -r ${workdir}/cp /
```
#### Then
- exit code is not `0`
## mimixbox cp GNU flags
Source: `test/e2e/tools/mimixbox/fileutils/cp_gnu.atago.yaml`
### Scenario: copies into the target directory (-t equals destination-last form)
#### When
```shell
printf 'A\n' > a.txt && printf 'B\n' > b.txt && mkdir dst_t dst_plain && cp --target-directory dst_t a.txt b.txt && cp a.txt b.txt dst_plain && [ "$(ls dst_t)" = "$(ls dst_plain)" ] && cat dst_t/a.txt dst_t/b.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
A
B
```
### Scenario: rejects a directory destination with --no-target-directory (-T)
#### When
```shell
printf 'A\n' > a.txt && mkdir b && cp -T a.txt b
```
#### Then
- exit code is not `0`
- stderr contains `cannot overwrite directory`
### Scenario: recreates the source path prefix with --parents
#### When
```shell
mkdir -p src/a dst && printf 'deep\n' > src/a/b.txt && cp --parents src/a/b.txt dst/ && cat dst/src/a/b.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: makes a backup before overwriting with --backup
#### When
```shell
printf 'new\n' > src.txt && printf 'old\n' > dst.txt && cp --backup=simple src.txt dst.txt && cat dst.txt dst.txt~

```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
new
old
```
### Scenario: skips the copy when the destination is newer (-u)
#### When
```shell
printf 'srcdata\n' > src.txt && printf 'dstdata\n' > dst.txt && touch -d '2020-01-01T00:00:00' src.txt && touch -d '2021-01-01T00:00:00' dst.txt && cp -u src.txt dst.txt && cat dst.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: skips a newer file inside the tree with -ru (#940)
#### When
```shell
mkdir -p src dst/src && printf 'from-src\n' > src/f.txt && printf 'newer-dst\n' > dst/src/f.txt && touch -d '2020-01-01T00:00:00' src/f.txt && touch -d '2021-01-01T00:00:00' dst/src/f.txt && cp -ru src dst && cat dst/src/f.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: backs up a file inside the tree with -r --backup (#940)
#### When
```shell
mkdir -p src dst/src && printf 'new\n' > src/f.txt && printf 'old\n' > dst/src/f.txt && cp -r --backup=simple src dst && cat dst/src/f.txt dst/src/f.txt~

```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
new
old
```
## mimixbox cp symlink handling
Source: `test/e2e/tools/mimixbox/fileutils/cp_symlink.atago.yaml`
### Scenario: cp -P copies the symlink as a link
#### When
```shell
echo data > real.txt
ln -s real.txt link
cp -P link copy
[ -L copy ] && echo link || echo notlink

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: cp -L copies the link target as a regular file
#### When
```shell
echo data > real.txt
ln -s real.txt link
cp -L link copy
if [ -L copy ]; then echo link; elif [ -f copy ]; then echo regular; else echo missing; fi

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: cp -d preserves a symlink inside a copied tree
#### When
```shell
mkdir -p src
echo data > src/real.txt
ln -s real.txt src/lnk
cp -d -r src dst
[ -L dst/lnk ] && echo link || echo notlink

```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox fileutils help helpers
Source: `test/e2e/tools/mimixbox/fileutils/help_helpers_fileutils.atago.yaml`
### Scenario: chgrp --help is structured
#### When
```shell
chgrp --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: chown --help is structured
#### When
```shell
chown --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
## mimixbox link
Source: `test/e2e/tools/mimixbox/fileutils/link.atago.yaml`
### Scenario: creates a hard link sharing contents
#### Given
- Fixture file `link_src` is created.
#### Inputs
_Fixture `link_src`:_
```
data
```
#### When
```shell
link ${workdir}/link_src ${workdir}/link_dst && cat ${workdir}/link_dst
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox ln
Source: `test/e2e/tools/mimixbox/fileutils/ln.atago.yaml`
### Scenario: ln creates a hard link to the same content
#### Given
- Fixture file `ln/target.txt` is created.
#### Inputs
_Fixture `ln/target.txt`:_
```
content
```
#### When
```shell
ln ${workdir}/ln/target.txt ${workdir}/ln/hardlink.txt && cat ${workdir}/ln/hardlink.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: ln -s creates a symbolic link
#### Given
- Fixture file `ln/target.txt` is created.
#### Inputs
_Fixture `ln/target.txt`:_
```
content
```
#### When
```shell
ln -s ${workdir}/ln/target.txt ${workdir}/ln/symlink.txt && test -L ${workdir}/ln/symlink.txt && echo "is symlink"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: ln with no operand reports an error
#### When
```shell
ln
```
#### Then
- exit code is not `0`
- stderr equals an exact value
## mimixbox ln GNU flags
Source: `test/e2e/tools/mimixbox/fileutils/ln_gnu.atago.yaml`
### Scenario: ln -s --relative stores the target relative to the link location
#### When
```shell
mkdir -p ln_gnu/a ln_gnu/b ln_gnu/dst
printf 'content\n' > ln_gnu/a/target.txt

ln -s --relative ${workdir}/ln_gnu/a/target.txt ${workdir}/ln_gnu/b/link.txt && readlink ${workdir}/ln_gnu/b/link.txt
```
#### Then
- after `ln -s --relative ${workdir}/ln_gnu/a/target.txt ${workdir}/ln_gnu/b/link.txt && readlink ${workdir}/ln_gnu/b/link.txt`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: ln --target-directory links each operand into the directory
#### When
```shell
mkdir -p ln_gnu/a ln_gnu/b ln_gnu/dst
printf 'A\n' > ln_gnu/a.txt
printf 'B\n' > ln_gnu/b.txt

ln --target-directory ${workdir}/ln_gnu/dst ${workdir}/ln_gnu/a.txt ${workdir}/ln_gnu/b.txt && ls ${workdir}/ln_gnu/dst
```
#### Then
- after `ln --target-directory ${workdir}/ln_gnu/dst ${workdir}/ln_gnu/a.txt ${workdir}/ln_gnu/b.txt && ls ${workdir}/ln_gnu/dst`:
  - exit code is `0`
  - stdout equals an exact value
#### Expected output
_expected stdout:_
```
a.txt
b.txt
```
## mimixbox ls
Source: `test/e2e/tools/mimixbox/fileutils/ls.atago.yaml`
### Scenario: lists entries sorted, hiding dotfiles
#### When
```shell
mkdir -p ls/sub && touch ls/a.txt ls/b.txt ls/.hidden
ls ${workdir}/ls
```
#### Then
- after `ls ${workdir}/ls`:
  - exit code is `0`
  - stdout equals an exact value
#### Expected output
_expected stdout:_
```
a.txt
b.txt
sub
```
### Scenario: includes dotfiles with -a
#### When
```shell
mkdir -p ls/sub && touch ls/a.txt ls/b.txt ls/.hidden
ls -a ${workdir}/ls
```
#### Then
- after `ls -a ${workdir}/ls`:
  - exit code is `0`
  - stdout contains `.hidden`
### Scenario: marks directories with -F
#### When
```shell
mkdir -p ls/sub && touch ls/a.txt ls/b.txt ls/.hidden
ls -F ${workdir}/ls
```
#### Then
- after `ls -F ${workdir}/ls`:
  - exit code is `0`
  - stdout contains `sub/`
## mimixbox ls GNU flags
Source: `test/e2e/tools/mimixbox/fileutils/ls_gnu.atago.yaml`
### Scenario: colors directories with --color=always
#### When
```shell
mkdir -p ls_gnu/adir
printf 'xxxxxxxxxx' > ls_gnu/small.txt
dd if=/dev/zero of=ls_gnu/big.txt bs=1 count=5000 2>/dev/null
: > ls_gnu/a.log
: > ls_gnu/tmpfile
printf '#!/bin/sh\n' > ls_gnu/run.sh
chmod 0755 ls_gnu/run.sh
ln -s small.txt ls_gnu/link

ls --color=always ${workdir}/ls_gnu
```
#### Then
- after `ls --color=always ${workdir}/ls_gnu`:
  - exit code is `0`
  - stdout contains `adir`
  - stdout matches `/\x1b\[01;34m/`
### Scenario: emits no escapes with --color=never
#### When
```shell
mkdir -p ls_gnu/adir
printf 'xxxxxxxxxx' > ls_gnu/small.txt
dd if=/dev/zero of=ls_gnu/big.txt bs=1 count=5000 2>/dev/null
: > ls_gnu/a.log
: > ls_gnu/tmpfile
printf '#!/bin/sh\n' > ls_gnu/run.sh
chmod 0755 ls_gnu/run.sh
ln -s small.txt ls_gnu/link

ls --color=never ${workdir}/ls_gnu
```
#### Then
- after `ls --color=never ${workdir}/ls_gnu`:
  - exit code is `0`
  - stdout does not contain ``
### Scenario: appends / * @ with -F
#### When
```shell
mkdir -p ls_gnu/adir
printf 'xxxxxxxxxx' > ls_gnu/small.txt
dd if=/dev/zero of=ls_gnu/big.txt bs=1 count=5000 2>/dev/null
: > ls_gnu/a.log
: > ls_gnu/tmpfile
printf '#!/bin/sh\n' > ls_gnu/run.sh
chmod 0755 ls_gnu/run.sh
ln -s small.txt ls_gnu/link

ls -F ${workdir}/ls_gnu
```
#### Then
- after `ls -F ${workdir}/ls_gnu`:
  - exit code is `0`
  - stdout contains `adir/`, `run.sh*`, `link@`
### Scenario: omits * for executables with --file-type
#### When
```shell
mkdir -p ls_gnu/adir
printf 'xxxxxxxxxx' > ls_gnu/small.txt
dd if=/dev/zero of=ls_gnu/big.txt bs=1 count=5000 2>/dev/null
: > ls_gnu/a.log
: > ls_gnu/tmpfile
printf '#!/bin/sh\n' > ls_gnu/run.sh
chmod 0755 ls_gnu/run.sh
ln -s small.txt ls_gnu/link

ls --file-type ${workdir}/ls_gnu
```
#### Then
- after `ls --file-type ${workdir}/ls_gnu`:
  - exit code is `0`
  - stdout contains `adir/`
  - stdout does not contain `run.sh*`
### Scenario: marks only dirs with --indicator-style=slash
#### When
```shell
mkdir -p ls_gnu/adir
printf 'xxxxxxxxxx' > ls_gnu/small.txt
dd if=/dev/zero of=ls_gnu/big.txt bs=1 count=5000 2>/dev/null
: > ls_gnu/a.log
: > ls_gnu/tmpfile
printf '#!/bin/sh\n' > ls_gnu/run.sh
chmod 0755 ls_gnu/run.sh
ln -s small.txt ls_gnu/link

ls --indicator-style=slash ${workdir}/ls_gnu
```
#### Then
- after `ls --indicator-style=slash ${workdir}/ls_gnu`:
  - exit code is `0`
  - stdout contains `adir/`
  - stdout does not contain `link@`
### Scenario: lists largest first with --sort=size
#### When
```shell
mkdir -p ls_gnu/adir
printf 'xxxxxxxxxx' > ls_gnu/small.txt
dd if=/dev/zero of=ls_gnu/big.txt bs=1 count=5000 2>/dev/null
: > ls_gnu/a.log
: > ls_gnu/tmpfile
printf '#!/bin/sh\n' > ls_gnu/run.sh
chmod 0755 ls_gnu/run.sh
ln -s small.txt ls_gnu/link

ls --sort=size ${workdir}/ls_gnu
```
#### Then
- after `ls --sort=size ${workdir}/ls_gnu`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: lists directories first with --group-directories-first
#### When
```shell
mkdir -p ls_gnu/adir
printf 'xxxxxxxxxx' > ls_gnu/small.txt
dd if=/dev/zero of=ls_gnu/big.txt bs=1 count=5000 2>/dev/null
: > ls_gnu/a.log
: > ls_gnu/tmpfile
printf '#!/bin/sh\n' > ls_gnu/run.sh
chmod 0755 ls_gnu/run.sh
ln -s small.txt ls_gnu/link

ls --group-directories-first ${workdir}/ls_gnu
```
#### Then
- after `ls --group-directories-first ${workdir}/ls_gnu`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: drops matches with --ignore
#### When
```shell
mkdir -p ls_gnu/adir
printf 'xxxxxxxxxx' > ls_gnu/small.txt
dd if=/dev/zero of=ls_gnu/big.txt bs=1 count=5000 2>/dev/null
: > ls_gnu/a.log
: > ls_gnu/tmpfile
printf '#!/bin/sh\n' > ls_gnu/run.sh
chmod 0755 ls_gnu/run.sh
ln -s small.txt ls_gnu/link

ls --ignore=*.log ${workdir}/ls_gnu
```
#### Then
- after `ls --ignore=*.log ${workdir}/ls_gnu`:
  - exit code is `0`
  - stdout does not contain `a.log`
### Scenario: drops matches with --hide
#### When
```shell
mkdir -p ls_gnu/adir
printf 'xxxxxxxxxx' > ls_gnu/small.txt
dd if=/dev/zero of=ls_gnu/big.txt bs=1 count=5000 2>/dev/null
: > ls_gnu/a.log
: > ls_gnu/tmpfile
printf '#!/bin/sh\n' > ls_gnu/run.sh
chmod 0755 ls_gnu/run.sh
ln -s small.txt ls_gnu/link

ls --hide=tmp* ${workdir}/ls_gnu
```
#### Then
- after `ls --hide=tmp* ${workdir}/ls_gnu`:
  - exit code is `0`
  - stdout does not contain `tmpfile`
### Scenario: keeps hidden matches when -a is given
#### When
```shell
mkdir -p ls_gnu/adir
printf 'xxxxxxxxxx' > ls_gnu/small.txt
dd if=/dev/zero of=ls_gnu/big.txt bs=1 count=5000 2>/dev/null
: > ls_gnu/a.log
: > ls_gnu/tmpfile
printf '#!/bin/sh\n' > ls_gnu/run.sh
chmod 0755 ls_gnu/run.sh
ln -s small.txt ls_gnu/link

ls -a --hide=tmp* ${workdir}/ls_gnu
```
#### Then
- after `ls -a --hide=tmp* ${workdir}/ls_gnu`:
  - exit code is `0`
  - stdout contains `tmpfile`
### Scenario: prints an inode number with -i
#### When
```shell
mkdir -p ls_gnu/adir
printf 'xxxxxxxxxx' > ls_gnu/small.txt

ls -i ${workdir}/ls_gnu/small.txt
```
#### Then
- after `ls -i ${workdir}/ls_gnu/small.txt`:
  - exit code is `0`
  - stdout matches `/[0-9]+ +.*small.txt/`
### Scenario: scales sizes to 1024-byte blocks with -k
#### When
```shell
mkdir -p ls_gnu/adir
dd if=/dev/zero of=ls_gnu/big.txt bs=1 count=5000 2>/dev/null

ls -l -k ${workdir}/ls_gnu/big.txt
```
#### Then
- after `ls -l -k ${workdir}/ls_gnu/big.txt`:
  - exit code is `0`
  - stdout contains ` 5 `
## mimixbox mkdir
Source: `test/e2e/tools/mimixbox/fileutils/mkdir.atago.yaml`
### Scenario: make a single directory
#### When
```shell
mkdir -p mkdir && mkdir ${workdir}/mkdir/single && ls ${workdir}/mkdir
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: make three directories
#### When
```shell
mkdir -p mkdir && mkdir ${workdir}/mkdir/1 ${workdir}/mkdir/2 ${workdir}/mkdir/3 && ls ${workdir}/mkdir
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
1
2
3
```
### Scenario: make a parent/child directory with -p
#### When
```shell
mkdir -p mkdir && mkdir -p ${workdir}/mkdir/parents/child && ls ${workdir}/mkdir/parents/
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: make a directory from a pipe
#### When
```shell
mkdir -p mkdir && echo "${workdir}/mkdir/pipe" | xargs mkdir && ls ${workdir}/mkdir
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: print error without operand
#### When
```shell
mkdir
```
#### Then
- exit code is not `0`
- stderr equals an exact value
### Scenario: print error with --parents and no operand
#### When
```shell
mkdir -p
```
#### Then
- exit code is not `0`
- stderr equals an exact value
### Scenario: make 1 and 3 but fail to make 2 at an unwritable path
#### When
```shell
mkdir -p mkdir && mkdir ${workdir}/mkdir/1 /mkdir/2 ${workdir}/mkdir/3; ls ${workdir}/mkdir/
```
#### Then
- stdout equals an exact value
- stderr equals an exact value
#### Expected output
_expected stdout:_
```
1
3
```
## mimixbox mkfifo
Source: `test/e2e/tools/mimixbox/fileutils/mkfifo.atago.yaml`
### Scenario: make one named pipe with mode prw-r--r--
#### When
```shell
mkdir -p mkfifo && chmod 775 mkfifo && mkfifo ${workdir}/mkfifo/1 && ls -al ${workdir}/mkfifo/1 | cut -f 1 -d " "
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: make three named pipes with mode prw-r--r--
#### When
```shell
mkdir -p mkfifo && chmod 775 mkfifo
mkfifo ${workdir}/mkfifo/1 ${workdir}/mkfifo/2 ${workdir}/mkfifo/3
ls -al ${workdir}/mkfifo/1 | cut -f 1 -d " "
ls -al ${workdir}/mkfifo/2 | cut -f 1 -d " "
ls -al ${workdir}/mkfifo/3 | cut -f 1 -d " "

```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
prw-r--r--
prw-r--r--
prw-r--r--
```
### Scenario: print error for a non-existent path
#### When
```shell
mkfifo /no_exist_path/fifo
```
#### Then
- exit code is not `0`
- stderr equals an exact value
### Scenario: print error when the same name already exists
#### When
```shell
mkdir -p mkfifo && chmod 775 mkfifo && mkfifo ${workdir}/mkfifo/1
mkfifo ${workdir}/mkfifo/1
```
#### Then
- after `mkfifo ${workdir}/mkfifo/1`:
  - exit code is not `0`
  - stderr equals an exact value
### Scenario: make two pipes and report the one that failed
#### When
```shell
mkdir -p mkfifo && chmod 775 mkfifo
mkfifo ${workdir}/mkfifo/1 /no_exist_path/fifo ${workdir}/mkfifo/3
ls ${workdir}/mkfifo

```
#### Then
- stdout equals an exact value
- stderr equals an exact value
#### Expected output
_expected stdout:_
```
1
3
```
## mimixbox mountpoint
Source: `test/e2e/tools/mimixbox/fileutils/mountpoint.atago.yaml`
### Scenario: reports that / is a mountpoint
#### When
```shell
mountpoint /
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox mv
Source: `test/e2e/tools/mimixbox/fileutils/mv.atago.yaml`
### Scenario: rename a file
#### When
```shell
mkdir -p mv/inner mv2 mv3 mv4 && touch mv/1.txt mv/2.txt mv/3.txt mv/inner/inner.txt
mv ${workdir}/mv/1.txt ${workdir}/mv/rename.txt && ls ${workdir}/mv/rename.txt
```
#### Then
- after `mv ${workdir}/mv/1.txt ${workdir}/mv/rename.txt && ls ${workdir}/mv/rename.txt`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: move a file into an inner directory
#### When
```shell
mkdir -p mv/inner mv2 mv3 mv4 && touch mv/1.txt mv/2.txt mv/3.txt mv/inner/inner.txt
mv ${workdir}/mv/1.txt ${workdir}/mv/inner && ls ${workdir}/mv/inner
```
#### Then
- after `mv ${workdir}/mv/1.txt ${workdir}/mv/inner && ls ${workdir}/mv/inner`:
  - exit code is `0`
  - stdout equals an exact value
#### Expected output
_expected stdout:_
```
1.txt
inner.txt
```
### Scenario: move three files into an inner directory
#### When
```shell
mkdir -p mv/inner mv2 mv3 mv4 && touch mv/1.txt mv/2.txt mv/3.txt mv/inner/inner.txt
mv ${workdir}/mv/1.txt ${workdir}/mv/2.txt ${workdir}/mv/3.txt ${workdir}/mv/inner && ls ${workdir}/mv/inner
```
#### Then
- after `mv ${workdir}/mv/1.txt ${workdir}/mv/2.txt ${workdir}/mv/3.txt ${workdir}/mv/inner && ls ${workdir}/mv/inner`:
  - exit code is `0`
  - stdout equals an exact value
#### Expected output
_expected stdout:_
```
1.txt
2.txt
3.txt
inner.txt
```
### Scenario: move three files where one does not exist
#### When
```shell
mkdir -p mv/inner mv2 mv3 mv4 && touch mv/1.txt mv/2.txt mv/3.txt mv/inner/inner.txt
mv ${workdir}/mv/1.txt ${workdir}/mv/no_exist_file ${workdir}/mv/3.txt ${workdir}/mv/inner; ls ${workdir}/mv/inner
```
#### Then
- after `mv ${workdir}/mv/1.txt ${workdir}/mv/no_exist_file ${workdir}/mv/3.txt ${workdir}/mv/inner; ls ${workdir}/mv/inner`:
  - stdout equals an exact value
  - stderr equals an exact value
#### Expected output
_expected stdout:_
```
1.txt
3.txt
inner.txt
```
### Scenario: move a directory into a directory
#### When
```shell
mkdir -p mv/inner mv2 mv3 mv4 && touch mv/1.txt mv/2.txt mv/3.txt mv/inner/inner.txt
mv ${workdir}/mv2 ${workdir}/mv && ls ${workdir}/mv
```
#### Then
- after `mv ${workdir}/mv2 ${workdir}/mv && ls ${workdir}/mv`:
  - exit code is `0`
  - stdout equals an exact value
#### Expected output
_expected stdout:_
```
1.txt
2.txt
3.txt
inner
mv2
```
### Scenario: move three directories
#### When
```shell
mkdir -p mv/inner mv2 mv3 mv4 && touch mv/1.txt mv/2.txt mv/3.txt mv/inner/inner.txt
mv ${workdir}/mv2 ${workdir}/mv3 ${workdir}/mv4 ${workdir}/mv && ls ${workdir}/mv
```
#### Then
- after `mv ${workdir}/mv2 ${workdir}/mv3 ${workdir}/mv4 ${workdir}/mv && ls ${workdir}/mv`:
  - exit code is `0`
  - stdout equals an exact value
#### Expected output
_expected stdout:_
```
1.txt
2.txt
3.txt
inner
mv2
mv3
mv4
```
### Scenario: move three directories where one does not exist
#### When
```shell
mkdir -p mv/inner mv2 mv3 mv4 && touch mv/1.txt mv/2.txt mv/3.txt mv/inner/inner.txt
mv ${workdir}/mv2 ${workdir}/mv/no_exist_dir ${workdir}/mv4 ${workdir}/mv/inner; ls ${workdir}/mv/inner
```
#### Then
- after `mv ${workdir}/mv2 ${workdir}/mv/no_exist_dir ${workdir}/mv4 ${workdir}/mv/inner; ls ${workdir}/mv/inner`:
  - stdout equals an exact value
  - stderr equals an exact value
#### Expected output
_expected stdout:_
```
inner.txt
mv2
mv4
```
### Scenario: moving a file onto itself fails
#### When
```shell
mkdir -p mv/inner mv2 mv3 mv4 && touch mv/1.txt mv/2.txt mv/3.txt mv/inner/inner.txt
mv ${workdir}/mv/1.txt ${workdir}/mv/1.txt
```
#### Then
- after `mv ${workdir}/mv/1.txt ${workdir}/mv/1.txt`:
  - exit code is not `0`
  - stderr equals an exact value
#### Expected output
_expected stderr:_
```
mv: source '${workdir}/mv/1.txt' and destination '${workdir}/mv/1.txt' is same
```
### Scenario: overwrite a file with the same destination name
#### When
```shell
mkdir -p mv/inner mv2 mv3 mv4 && touch mv/1.txt mv/2.txt mv/3.txt mv/inner/inner.txt
touch ${workdir}/mv/inner.txt && mv ${workdir}/mv/inner.txt ${workdir}/mv/inner/inner.txt && ls ${workdir}/mv && ls ${workdir}/mv/inner
```
#### Then
- after `touch ${workdir}/mv/inner.txt && mv ${workdir}/mv/inner.txt ${workdir}/mv/inner/inner.txt && ls ${workdir}/mv && ls ${workdir}/mv/inner`:
  - exit code is `0`
  - stdout equals an exact value
#### Expected output
_expected stdout:_
```
1.txt
2.txt
3.txt
inner
inner.txt
```
### Scenario: overwrite with the backup option keeps a tilde copy
#### When
```shell
mkdir -p mv/inner mv2 mv3 mv4 && touch mv/1.txt mv/2.txt mv/3.txt mv/inner/inner.txt
touch ${workdir}/mv/inner.txt && mv -b ${workdir}/mv/inner.txt ${workdir}/mv/inner && ls ${workdir}/mv/inner
```
#### Then
- after `touch ${workdir}/mv/inner.txt && mv -b ${workdir}/mv/inner.txt ${workdir}/mv/inner && ls ${workdir}/mv/inner`:
  - exit code is `0`
  - stdout equals an exact value
#### Expected output
_expected stdout:_
```
inner.txt
inner.txt~
```
## mimixbox mv GNU flags
Source: `test/e2e/tools/mimixbox/fileutils/mv_gnu.atago.yaml`
### Scenario: mv --target-directory moves each source into the directory
#### When
```shell
mkdir -p mv_gnu/dst
printf 'A\n' > mv_gnu/a.txt
printf 'B\n' > mv_gnu/b.txt

mv --target-directory ${workdir}/mv_gnu/dst ${workdir}/mv_gnu/a.txt ${workdir}/mv_gnu/b.txt && ls ${workdir}/mv_gnu/dst
```
#### Then
- after `mv --target-directory ${workdir}/mv_gnu/dst ${workdir}/mv_gnu/a.txt ${workdir}/mv_gnu/b.txt && ls ${workdir}/mv_gnu/dst`:
  - exit code is `0`
  - stdout equals an exact value
#### Expected output
_expected stdout:_
```
a.txt
b.txt
```
### Scenario: mv --update preserves a newer destination
#### When
```shell
mkdir -p mv_gnu
printf 'source\n' > mv_gnu/src.txt
sleep 1
printf 'newer-dest\n' > mv_gnu/dest.txt
mv --update ${workdir}/mv_gnu/src.txt ${workdir}/mv_gnu/dest.txt
cat ${workdir}/mv_gnu/dest.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox readlink
Source: `test/e2e/tools/mimixbox/fileutils/readlink.atago.yaml`
### Scenario: prints the symlink target
#### When
```shell
printf 'x' > ${workdir}/rl_target && ln -sf ${workdir}/rl_target ${workdir}/rl_link
readlink ${workdir}/rl_link
```
#### Then
- after `readlink ${workdir}/rl_link`:
  - exit code is `0`
  - stdout equals an exact value
## mimixbox readlink_gnu
Source: `test/e2e/tools/mimixbox/fileutils/readlink_gnu.atago.yaml`
### Scenario: fails when -e is given a missing path
#### Given
- Fixture file `target` is created.
- Fixture file `link` is created.
#### When
```shell
readlink -e ${workdir}/does-not-exist
```
#### Then
- exit code is not `0`
- stdout is empty
### Scenario: succeeds when -e is given an existing symlink
#### Given
- Fixture file `target` is created.
- Fixture file `link` is created.
#### When
```shell
readlink -e ${workdir}/link
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: succeeds with -m on a missing path
#### Given
- Fixture file `target` is created.
- Fixture file `link` is created.
#### When
```shell
readlink -m ${workdir}/a/b/c
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: terminates output with NUL under -z
#### Given
- Fixture file `target` is created.
- Fixture file `link` is created.
#### When
```shell
readlink -z ${workdir}/link
```
#### Then
- exit code is `0`
- stdout matches `/target\x00$/`
## mimixbox rm
Source: `test/e2e/tools/mimixbox/fileutils/rm.atago.yaml`
### Scenario: remove one file
#### When
```shell
mkdir -p rm/inner && touch rm/1.txt rm/2.txt rm/3.txt rm/inner/inner.txt
rm ${workdir}/rm/1.txt && ls ${workdir}/rm
```
#### Then
- after `rm ${workdir}/rm/1.txt && ls ${workdir}/rm`:
  - exit code is `0`
  - stdout equals an exact value
#### Expected output
_expected stdout:_
```
2.txt
3.txt
inner
```
### Scenario: remove files using a wildcard
#### When
```shell
mkdir -p rm/inner && touch rm/1.txt rm/2.txt rm/3.txt rm/inner/inner.txt
rm ${workdir}/rm/*.txt && ls ${workdir}/rm
```
#### Then
- after `rm ${workdir}/rm/*.txt && ls ${workdir}/rm`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: remove three files at the same time
#### When
```shell
mkdir -p rm/inner && touch rm/1.txt rm/2.txt rm/3.txt rm/inner/inner.txt
rm ${workdir}/rm/1.txt ${workdir}/rm/2.txt ${workdir}/rm/3.txt && ls ${workdir}/rm
```
#### Then
- after `rm ${workdir}/rm/1.txt ${workdir}/rm/2.txt ${workdir}/rm/3.txt && ls ${workdir}/rm`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: remove two files and report the missing one
#### When
```shell
mkdir -p rm/inner && touch rm/1.txt rm/2.txt rm/3.txt rm/inner/inner.txt
rm ${workdir}/rm/1.txt ${workdir}/rm/no_exist_file.txt ${workdir}/rm/3.txt; ls ${workdir}/rm
```
#### Then
- after `rm ${workdir}/rm/1.txt ${workdir}/rm/no_exist_file.txt ${workdir}/rm/3.txt; ls ${workdir}/rm`:
  - stdout equals an exact value
  - stderr equals an exact value
#### Expected output
_expected stdout:_
```
2.txt
inner
```
_expected stderr:_
```
rm: can't remove ${workdir}/rm/no_exist_file.txt: No such file or directory exists
```
### Scenario: can not remove a directory without the recursive option
#### When
```shell
mkdir -p rm/inner && touch rm/1.txt rm/2.txt rm/3.txt rm/inner/inner.txt
rm ${workdir}/rm; ls ${workdir}/rm
```
#### Then
- after `rm ${workdir}/rm; ls ${workdir}/rm`:
  - stdout equals an exact value
  - stderr equals an exact value
#### Expected output
_expected stdout:_
```
1.txt
2.txt
3.txt
inner
```
### Scenario: remove a directory with the recursive option
#### When
```shell
mkdir -p rm/inner && touch rm/1.txt rm/2.txt rm/3.txt rm/inner/inner.txt
rm -rf ${workdir}/rm && ls ${workdir}
```
#### Then
- after `rm -rf ${workdir}/rm && ls ${workdir}`:
  - exit code is `0`
## mimixbox rm GNU flags
Source: `test/e2e/tools/mimixbox/fileutils/rm_gnu.atago.yaml`
### Scenario: refuses to recurse on / by default (--preserve-root)
#### When
```shell
rm -r /
```
#### Then
- exit code is not `0`
- stderr contains `it is dangerous to operate recursively on '/'`, `use --no-preserve-root to override this failsafe`
### Scenario: removes an ordinary directory recursively (guard does not interfere)
#### When
```shell
mkdir -p tree/sub && : > tree/sub/leaf.txt && rm -r tree && [ ! -e tree ] && printf 'gone'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: removes a single-filesystem tree with --one-file-system
#### When
```shell
mkdir -p ofs/a/b && : > ofs/a/b/leaf.txt && : > ofs/top.txt && rm -r --one-file-system ofs && [ ! -e ofs ] && printf 'gone'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: removes a tree when --no-preserve-root is given (safe target)
#### When
```shell
mkdir -p victim/sub && : > victim/sub/f && rm -r --no-preserve-root victim && [ ! -e victim ] && printf 'gone'
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox rmdir
Source: `test/e2e/tools/mimixbox/fileutils/rmdir.atago.yaml`
### Scenario: removes an empty directory
#### When
```shell
mkdir -p ${workdir}/rmdir/empty && rmdir ${workdir}/rmdir/empty && test ! -d ${workdir}/rmdir/empty && echo "removed"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: fails on a non-empty directory
#### When
```shell
mkdir -p ${workdir}/rmdir/full && touch ${workdir}/rmdir/full/file.txt
rmdir ${workdir}/rmdir/full
```
#### Then
- after `rmdir ${workdir}/rmdir/full`:
  - exit code is not `0`
  - stderr equals an exact value
#### Expected output
_expected stderr:_
```
rmdir: failed to remove '${workdir}/rmdir/full': Directory not empty
```
### Scenario: rmdir -p removes nested empty directories
#### When
```shell
mkdir -p ${workdir}/rmdir/a/b/c
cd ${workdir}/rmdir
rmdir -p a/b/c
test ! -d ${workdir}/rmdir/a && echo "removed"

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: rmdir with no operand reports an error
#### When
```shell
rmdir
```
#### Then
- exit code is not `0`
- stderr equals an exact value
## mimixbox serial
Source: `test/e2e/tools/mimixbox/fileutils/serial.atago.yaml`
### Scenario: adds a serial-number prefix to each file
#### When
```shell
mkdir -p serial && touch serial/apple.txt serial/banana.txt serial/cherry.txt
serial ${workdir}/serial > /dev/null && ls ${workdir}/serial
```
#### Then
- after `serial ${workdir}/serial > /dev/null && ls ${workdir}/serial`:
  - exit code is `0`
  - stdout equals an exact value
#### Expected output
_expected stdout:_
```
0_apple.txt
1_banana.txt
2_cherry.txt
```
### Scenario: --dry-run does not rename anything
#### When
```shell
mkdir -p serial && touch serial/apple.txt serial/banana.txt serial/cherry.txt
serial -d ${workdir}/serial > /dev/null && ls ${workdir}/serial
```
#### Then
- after `serial -d ${workdir}/serial > /dev/null && ls ${workdir}/serial`:
  - exit code is `0`
  - stdout equals an exact value
#### Expected output
_expected stdout:_
```
apple.txt
banana.txt
cherry.txt
```
### Scenario: serial with no operand reports an error
#### When
```shell
serial
```
#### Then
- exit code is not `0`
- stderr equals an exact value
## mimixbox shred
Source: `test/e2e/tools/mimixbox/fileutils/shred.atago.yaml`
### Scenario: overwrites and removes the file
#### Given
- Fixture file `shred_file` is created.
#### Inputs
_Fixture `shred_file`:_
```
secret
```
#### When
```shell
shred -u ${workdir}/shred_file && test ! -e ${workdir}/shred_file && echo gone
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox stat
Source: `test/e2e/tools/mimixbox/fileutils/stat.atago.yaml`
### Scenario: prints the size with a custom format
#### Given
- Fixture file `stat_file` is created.
#### Inputs
_Fixture `stat_file`:_
```
hello
```
#### When
```shell
stat -c %s ${workdir}/stat_file
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox stat GNU flags
Source: `test/e2e/tools/mimixbox/fileutils/stat_gnu.atago.yaml`
### Scenario: prints name and size via --printf with no trailing newline
#### Given
- Fixture file `stat_file` is created.
#### Inputs
_Fixture `stat_file`:_
```
hello
```
#### When
```shell
stat --printf '%n=%s' ${workdir}/stat_file
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: interprets backslash escapes in --printf
#### Given
- Fixture file `stat_file` is created.
#### Inputs
_Fixture `stat_file`:_
```
hello
```
#### When
```shell
stat --printf '%n %s\n' ${workdir}/stat_file
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: appends a trailing newline for --format
#### Given
- Fixture file `stat_file` is created.
#### Inputs
_Fixture `stat_file`:_
```
hello
```
#### When
```shell
stat --format '%s' ${workdir}/stat_file | wc -l | tr -d ' '
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: prints a single space-separated terse line
#### Given
- Fixture file `stat_file` is created.
#### Inputs
_Fixture `stat_file`:_
```
hello
```
#### When
```shell
stat --terse ${workdir}/stat_file | awk '{print NF}'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: reports the size as the second terse field
#### Given
- Fixture file `stat_file` is created.
#### Inputs
_Fixture `stat_file`:_
```
hello
```
#### When
```shell
stat --terse ${workdir}/stat_file | awk '{print $2}'
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox touch
Source: `test/e2e/tools/mimixbox/fileutils/touch.atago.yaml`
### Scenario: make one file
#### When
```shell
mkdir -p touch && touch ${workdir}/touch/touch.txt && ls ${workdir}/touch/touch.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: make three files at the same time
#### When
```shell
mkdir -p touch
touch ${workdir}/touch/1.txt ${workdir}/touch/2.txt ${workdir}/touch/3.txt
ls ${workdir}/touch/1.txt
ls ${workdir}/touch/2.txt
ls ${workdir}/touch/3.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
${workdir}/touch/1.txt
${workdir}/touch/2.txt
${workdir}/touch/3.txt
```
### Scenario: make two files and fail to make one at an unwritable path
#### When
```shell
mkdir -p touch
touch ${workdir}/touch/1.txt /touch/2.txt ${workdir}/touch/3.txt
ls ${workdir}/touch/1.txt
ls ${workdir}/touch/3.txt

```
#### Then
- stdout equals an exact value
- stderr equals an exact value
#### Expected output
_expected stdout:_
```
${workdir}/touch/1.txt
${workdir}/touch/3.txt
```
## mimixbox touch GNU flags
Source: `test/e2e/tools/mimixbox/fileutils/touch_gnu.atago.yaml`
### Scenario: copies the reference file mtime (--reference)
#### When
```shell
: > ref && : > dst && touch -d '2001-02-03 04:05:06' ref && touch --reference=ref dst && [ "$(/usr/bin/stat -c '%Y' ref)" = "$(/usr/bin/stat -c '%Y' dst)" ] && printf 'match'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: sets a known time with --date
#### When
```shell
: > f && touch -d '2020-06-15 12:34:56' f && [ "$(/bin/date -d '2020-06-15 12:34:56' '+%s')" = "$(/usr/bin/stat -c '%Y' f)" ] && printf 'match'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: accepts --time=atime without error
#### When
```shell
touch --time=atime g
```
#### Then
- exit code is `0`
- file `g` exists
#### Generated artifacts
- `g`
### Scenario: rejects an invalid --time word
#### When
```shell
touch --time=bogus h
```
#### Then
- exit code is not `0`
- stderr contains `invalid argument`
### Scenario: changes the symlink itself with --no-dereference (-h)
#### When
```shell
: > target && touch -d '2005-05-05 05:05:05' target && ln -s target link && touch -h -d '2030-01-01 00:00:00' link && [ "$(/bin/date -d '2005-05-05 05:05:05' '+%s')" = "$(/usr/bin/stat -c '%Y' target)" ] && printf 'target-unchanged'
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox truncate
Source: `test/e2e/tools/mimixbox/fileutils/truncate.atago.yaml`
### Scenario: sets the file to the given size
#### When
```shell
truncate -s 7 ${workdir}/tr_file && wc -c < ${workdir}/tr_file | tr -d ' '
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox unlink
Source: `test/e2e/tools/mimixbox/fileutils/unlink.atago.yaml`
### Scenario: removes a single file
#### Given
- Fixture file `unlink.txt` is created.
#### Inputs
_Fixture `unlink.txt`:_
```
x
```
#### When
```shell
unlink ${workdir}/unlink.txt && test ! -e ${workdir}/unlink.txt && echo gone
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox alias parity
Source: `test/e2e/tools/mimixbox/findutils/alias_parity.atago.yaml`
### Scenario: egrep matches the same lines as grep -E over the same file
#### Given
- Fixture file `fixture.txt` is created.
#### Inputs
_Fixture `fixture.txt`:_
```
apple
banana
a.b
axb
```
#### When
```shell
egrep 'a(p|x)' fixture.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
apple
axb
```
### Scenario: fgrep matches the same lines as grep -F over the same file
#### Given
- Fixture file `fixture.txt` is created.
#### Inputs
_Fixture `fixture.txt`:_
```
apple
banana
a.b
axb
```
#### When
```shell
fgrep 'a.b' fixture.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: netcat answers --help with exit 0 and an Examples section
#### When
```shell
netcat --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: netcat`, `Examples:`
- stderr is empty
### Scenario: nc answers --help with exit 0 and an Examples section
#### When
```shell
nc --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: nc`, `Examples:`
- stderr is empty
## mimixbox egrep
Source: `test/e2e/tools/mimixbox/findutils/egrep.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
egrep --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: egrep`
- stderr is empty
## mimixbox fgrep
Source: `test/e2e/tools/mimixbox/findutils/fgrep.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
fgrep --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: fgrep`
- stderr is empty
## mimixbox find
Source: `test/e2e/tools/mimixbox/findutils/find.atago.yaml`
### Scenario: finds a file by -name
#### When
```shell
mkdir -p find/sub && touch find/a.txt find/sub/b.txt
find ${workdir}/find -name a.txt
```
#### Then
- after `find ${workdir}/find -name a.txt`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: lists directories with -type d
#### When
```shell
mkdir -p find/sub && touch find/a.txt find/sub/b.txt
find ${workdir}/find -type d | wc -l | tr -d ' '
```
#### Then
- after `find ${workdir}/find -type d | wc -l | tr -d ' '`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: rejects an unknown predicate
#### When
```shell
mkdir -p find/sub && touch find/a.txt find/sub/b.txt
find ${workdir}/find -bogus
```
#### Then
- after `find ${workdir}/find -bogus`:
  - exit code is not `0`
  - stderr contains `unknown predicate`
### Scenario: prints usage for --help
#### When
```shell
find --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: find`
### Scenario: lists an Options block with --help and --version in --help
#### When
```shell
find --help
```
#### Then
- exit code is `0`
- stdout contains `Options:`, `--help`, `--version`
### Scenario: documents the supported subset tokens in --help
#### When
```shell
find --help
```
#### Then
- exit code is `0`
- stdout contains `-name`, `-type`, `-print0`, `-maxdepth`, `-mindepth`
### Scenario: prints the version line for --version
#### When
```shell
find --version
```
#### Then
- exit code is `0`
- stdout contains `find (mimixbox)`
## mimixbox grep
Source: `test/e2e/tools/mimixbox/findutils/grep.atago.yaml`
### Scenario: matches lines from stdin
#### When
```shell
printf 'one\ntwo\nthree\n' | grep two
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: matches lines from a file
#### Given
- Fixture file `fruits.txt` is created.
#### Inputs
_Fixture `fruits.txt`:_
```
apple
banana
cherry
```
#### When
```shell
grep an ${workdir}/fruits.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: counts matching lines with -c
#### Given
- Fixture file `fruits.txt` is created.
#### Inputs
_Fixture `fruits.txt`:_
```
apple
banana
cherry
```
#### When
```shell
grep -c a ${workdir}/fruits.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: exits 1 when nothing matches
#### When
```shell
printf 'x\n' | grep zzz
```
#### Then
- exit code is `1`
## mimixbox grep GNU flags
Source: `test/e2e/tools/mimixbox/findutils/grep_gnu.atago.yaml`
### Scenario: prints trailing context with -A1
#### Given
- Fixture file `grep_gnu/ctx.txt` is created.
#### Inputs
_Fixture `grep_gnu/ctx.txt`:_
```
1
2
MATCH
b
c
```
#### When
```shell
grep -A1 MATCH ${workdir}/grep_gnu/ctx.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
MATCH
b
```
### Scenario: prints leading context with -B1
#### Given
- Fixture file `grep_gnu/ctx2.txt` is created.
#### Inputs
_Fixture `grep_gnu/ctx2.txt`:_
```
a
MATCH
b
```
#### When
```shell
grep -B1 MATCH ${workdir}/grep_gnu/ctx2.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
a
MATCH
```
### Scenario: prints surrounding context with -C1
#### Given
- Fixture file `grep_gnu/ctx2.txt` is created.
#### Inputs
_Fixture `grep_gnu/ctx2.txt`:_
```
a
MATCH
b
```
#### When
```shell
grep -C1 MATCH ${workdir}/grep_gnu/ctx2.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
a
MATCH
b
```
### Scenario: separates non-contiguous groups with --
#### Given
- Fixture file `grep_gnu/groups.txt` is created.
#### Inputs
_Fixture `grep_gnu/groups.txt`:_
```
MATCH
b
c
d
e
f
MATCH
```
#### When
```shell
grep -A1 MATCH ${workdir}/grep_gnu/groups.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: searches only included files with --include
#### When
```shell
mkdir -p grep_gnu
printf 'needle\n' > grep_gnu/keep.go
printf 'needle\n' > grep_gnu/skip.txt

grep -r --include=*.go needle ${workdir}/grep_gnu
```
#### Then
- after `grep -r --include=*.go needle ${workdir}/grep_gnu`:
  - exit code is `0`
  - stdout contains `keep.go`
  - stdout does not contain `skip.txt`
### Scenario: skips excluded files with --exclude
#### When
```shell
mkdir -p grep_gnu
printf 'needle\n' > grep_gnu/keep.go
printf 'needle\n' > grep_gnu/app.log

grep -r --exclude=*.log needle ${workdir}/grep_gnu
```
#### Then
- after `grep -r --exclude=*.log needle ${workdir}/grep_gnu`:
  - exit code is `0`
  - stdout does not contain `app.log`
### Scenario: skips excluded directories with --exclude-dir
#### When
```shell
mkdir -p grep_gnu/src grep_gnu/vendor
printf 'needle\n' > grep_gnu/src/a.txt
printf 'needle\n' > grep_gnu/vendor/b.txt

grep -r --exclude-dir=vendor needle ${workdir}/grep_gnu
```
#### Then
- after `grep -r --exclude-dir=vendor needle ${workdir}/grep_gnu`:
  - exit code is `0`
  - stdout does not contain `vendor`
### Scenario: highlights matches with --color=always
#### When
```shell
printf 'hello world\n' | grep --color=always world
```
#### Then
- exit code is `0`
- stdout contains `world`
### Scenario: prints byte offsets with -b
#### Given
- Fixture file `grep_gnu/off.txt` is created.
#### Inputs
_Fixture `grep_gnu/off.txt`:_
```
aaa
bbb
ccc
```
#### When
```shell
grep -b bbb ${workdir}/grep_gnu/off.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: prints files without a match with -L
#### When
```shell
mkdir -p grep_gnu
printf 'needle\n' > grep_gnu/hit.txt
printf 'nothing\n' > grep_gnu/miss.txt

grep -L needle ${workdir}/grep_gnu/hit.txt ${workdir}/grep_gnu/miss.txt
```
#### Then
- after `grep -L needle ${workdir}/grep_gnu/hit.txt ${workdir}/grep_gnu/miss.txt`:
  - exit code is `0`
  - stdout contains `miss.txt`
  - stdout does not contain `hit.txt`
## mimixbox findutils help helpers
Source: `test/e2e/tools/mimixbox/findutils/help_helpers_findutils.atago.yaml`
### Scenario: egrep --help is structured
#### When
```shell
egrep --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: fgrep --help is structured
#### When
```shell
fgrep --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
## mimixbox xargs
Source: `test/e2e/tools/mimixbox/findutils/xargs.atago.yaml`
### Scenario: appends stdin items to the command
#### When
```shell
printf 'a b c\n' | xargs echo
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: splits into groups with -n
#### When
```shell
printf '1 2 3 4\n' | xargs -n 2 echo | wc -l | tr -d ' '
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: substitutes with -I
#### When
```shell
printf 'world\n' | xargs -I {} echo hello {}
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox xargs GNU flags
Source: `test/e2e/tools/mimixbox/findutils/xargs_gnu.atago.yaml`
### Scenario: runs once per input line with -L 1
#### When
```shell
printf 'a b\nc d\ne f\n' | xargs -L 1 echo | wc -l | tr -d ' '
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: groups two input lines per invocation with -L 2
#### When
```shell
printf 'a\nb\nc\n' | xargs -L 2 echo | wc -l | tr -d ' '
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: splits a long input into multiple invocations with -s
#### When
```shell
printf '1 2 3 4 5 6 7 8\n' | xargs -s 8 echo | wc -l | tr -d ' '
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: keeps every item across the -s split invocations
#### When
```shell
printf '1 2 3 4 5 6 7 8\n' | xargs -s 8 echo | tr ' ' '\n' | grep -c .
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: runs all batches concurrently with -P 4
#### When
```shell
printf 'a\nb\nc\nd\n' | xargs -P 4 -n 1 echo | sort | tr '\n' ' ' | sed 's/ $//'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: runs all batches with -P 0
#### When
```shell
printf 'a\nb\nc\nd\n' | xargs -P 0 -n 1 echo | sort | tr '\n' ' ' | sed 's/ $//'
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox lifegame CLI contract
Source: `test/e2e/tools/mimixbox/games/lifegame.atago.yaml`
### Scenario: prints usage with --help and exits 0
#### When
```shell
lifegame --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: lifegame`
## mimixbox banner
Source: `test/e2e/tools/mimixbox/jokeutils/banner.atago.yaml`
### Scenario: prints five rows of art
#### When
```shell
banner HI | wc -l | tr -d ' '
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox cmatrix
Source: `test/e2e/tools/mimixbox/jokeutils/cmatrix.atago.yaml`
### Scenario: exits gracefully without a terminal
#### When
```shell
cmatrix
echo "rc:$?"

```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox cowsay
Source: `test/e2e/tools/mimixbox/jokeutils/cowsay.atago.yaml`
### Scenario: prints usage with --help and exits 0
#### When
```shell
cowsay --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: cowsay`
### Scenario: renders the message in the speech bubble
#### When
```shell
cowsay hello
```
#### Then
- exit code is `0`
- stdout contains `hello`
## mimixbox cowthink
Source: `test/e2e/tools/mimixbox/jokeutils/cowthink.atago.yaml`
### Scenario: draws the thought-bubble connector
#### When
```shell
cowthink hi | head -n 4 | tail -n 1
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox fakemovie
Source: `test/e2e/tools/mimixbox/jokeutils/fakemovie.atago.yaml`
### Scenario: prints usage with --help and exits 0
#### When
```shell
fakemovie --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: fakemovie`
### Scenario: fails with a message when given no operand
#### When
```shell
fakemovie
```
#### Then
- exit code is not `0`
- stderr contains `fakemovie`
## mimixbox fortune
Source: `test/e2e/tools/mimixbox/jokeutils/fortune.atago.yaml`
### Scenario: prints a single adage line
#### When
```shell
fortune | wc -l | tr -d ' '
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox jokeutils --help contract
Source: `test/e2e/tools/mimixbox/jokeutils/help_helpers.atago.yaml`
### Scenario: sl --help is structured
#### When
```shell
sl --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
## mimixbox nyancat
Source: `test/e2e/tools/mimixbox/jokeutils/nyancat.atago.yaml`
### Scenario: exits gracefully without a terminal
#### When
```shell
nyancat
echo "rc:$?"

```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox sl
Source: `test/e2e/tools/mimixbox/jokeutils/sl.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
sl --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: sl`
- stderr is empty
## mimixbox acpid
Source: `test/e2e/tools/mimixbox/loginutils/acpid.atago.yaml`
### Scenario: requires foreground mode
#### When
```shell
acpid
```
#### Then
- exit code is not `0`
### Scenario: describes itself with --help
#### When
```shell
acpid --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: acpid`, `ACPI`
## mimixbox addgroup
Source: `test/e2e/tools/mimixbox/loginutils/addgroup.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
addgroup --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: addgroup`
- stderr is empty
## mimixbox adduser
Source: `test/e2e/tools/mimixbox/loginutils/adduser.atago.yaml`
### Scenario: requires a user name
#### When
```shell
adduser
```
#### Then
- exit code is not `0`
### Scenario: describes itself with --help
#### When
```shell
adduser --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: adduser`, `account`
## mimixbox bootchartd
Source: `test/e2e/tools/mimixbox/loginutils/bootchartd.atago.yaml`
### Scenario: records a proc_stat sample
#### When
```shell
bootchartd -o "${workdir}/bc" >/dev/null && grep -c '^cpu ' "${workdir}/bc/proc_stat.log"
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox chpasswd
Source: `test/e2e/tools/mimixbox/loginutils/chpasswd.atago.yaml`
### Scenario: rejects an unknown method
#### Inputs
_stdin for `chpasswd`:_
```
alice:secret
```
#### When
```shell
chpasswd -c bogus
```
#### Then
- exit code is not `0`
### Scenario: describes itself with --help
#### When
```shell
chpasswd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: chpasswd`, `password`
## mimixbox chsh
Source: `test/e2e/tools/mimixbox/loginutils/chsh.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
chsh --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: chsh`, `login shell`
### Scenario: lists shells and exits successfully
#### When
```shell
chsh -l
```
#### Then
- exit code is `0`
### Scenario: rejects an unknown user
#### When
```shell
chsh -s /bin/sh mimixbox-no-such-user-xyz
```
#### Then
- exit code is not `0`
### Scenario: rejects a relative shell path
#### When
```shell
chsh -s relative/shell root
```
#### Then
- exit code is not `0`
## mimixbox crond
Source: `test/e2e/tools/mimixbox/loginutils/crond.atago.yaml`
### Scenario: requires foreground mode
#### When
```shell
crond
```
#### Then
- exit code is not `0`
### Scenario: describes itself with --help
#### When
```shell
crond --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: crond`, `cron`
## mimixbox crontab
Source: `test/e2e/tools/mimixbox/loginutils/crontab.atago.yaml`
### Scenario: reports that interactive edit is unsupported
#### When
```shell
crontab -e
```
#### Then
- exit code is not `0`
### Scenario: describes itself with --help
#### When
```shell
crontab --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: crontab`, `crontab`
## mimixbox cryptpw
Source: `test/e2e/tools/mimixbox/loginutils/cryptpw.atago.yaml`
### Scenario: hashes a stdin password with sha-512
#### Inputs
_stdin for `cryptpw`:_
```
secret
```
#### When
```shell
cryptpw -S abcdefgh
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
$6$abcdefgh$ltjgWl6579NluT/Vi1nwEvcil.G5Nbc4NiXZaNGStk8PSwGfQv72N2CKPPrVACtLtip/cZ/1GM/O6IND4WQhG.
```
### Scenario: supports the md5 method
#### Inputs
_stdin for `cryptpw`:_
```
secret
```
#### When
```shell
cryptpw -m md5 -S abcdefgh
```
#### Then
- exit code is `0`
- stdout contains `$1$abcdefgh$`
## mimixbox delgroup
Source: `test/e2e/tools/mimixbox/loginutils/delgroup.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
delgroup --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: delgroup`
- stderr is empty
## mimixbox deluser
Source: `test/e2e/tools/mimixbox/loginutils/deluser.atago.yaml`
### Scenario: requires a user name
#### When
```shell
deluser
```
#### Then
- exit code is not `0`
### Scenario: describes itself with --help
#### When
```shell
deluser --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: deluser`, `user`
## mimixbox getty
Source: `test/e2e/tools/mimixbox/loginutils/getty.atago.yaml`
### Scenario: prints the login prompt
#### When
```shell
printf '\n' | getty tty1 2>/dev/null | grep -c 'login: '

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: requires a TTY argument
#### Inputs
_stdin for `getty`:_
```
alice
```
#### When
```shell
getty
```
#### Then
- exit code is not `0`
## mimixbox addgroup / delgroup
Source: `test/e2e/tools/mimixbox/loginutils/group.atago.yaml`
### Scenario: addgroup requires a group name
#### When
```shell
addgroup
```
#### Then
- exit code is not `0`
### Scenario: delgroup requires a group name
#### When
```shell
delgroup
```
#### Then
- exit code is not `0`
## mimixbox loginutils --help helpers
Source: `test/e2e/tools/mimixbox/loginutils/help_helpers_loginutils.atago.yaml`
### Scenario: addgroup --help is structured
#### When
```shell
addgroup --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: delgroup --help is structured
#### When
```shell
delgroup --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: linuxrc --help is structured
#### When
```shell
linuxrc --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: run-init --help is structured
#### When
```shell
run-init --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: run-parts --help is structured
#### When
```shell
run-parts --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: start-stop-daemon --help is structured
#### When
```shell
start-stop-daemon --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
## mimixbox init
Source: `test/e2e/tools/mimixbox/loginutils/init.atago.yaml`
### Scenario: runs the inittab sysinit and wait actions
#### Given
- Fixture file `inittab` is created.
#### Inputs
_Fixture `inittab`:_
```
si::sysinit:echo SYSINIT
l::wait:echo WAIT
```
#### When
```shell
init -t inittab
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout equals an exact value
### Scenario: fails on a missing inittab
#### When
```shell
init -t no_such_inittab
```
#### Then
- exit code is not `0`
## mimixbox login
Source: `test/e2e/tools/mimixbox/loginutils/login.atago.yaml`
### Scenario: fails for an unknown user
#### Inputs
_stdin for `login`:_
```
nope
```
#### When
```shell
login nonexistent_user_xyz
```
#### Then
- exit code is not `0`
### Scenario: describes itself with --help
#### When
```shell
login --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: login`, `login shell`
## mimixbox mkpasswd
Source: `test/e2e/tools/mimixbox/loginutils/mkpasswd.atago.yaml`
### Scenario: hashes with sha-512 and a fixed salt
#### When
```shell
mkpasswd -m sha-512 -S abcdefgh secret
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
$6$abcdefgh$ltjgWl6579NluT/Vi1nwEvcil.G5Nbc4NiXZaNGStk8PSwGfQv72N2CKPPrVACtLtip/cZ/1GM/O6IND4WQhG.
```
### Scenario: reads the password from stdin
#### Inputs
_stdin for `mkpasswd`:_
```
frompipe
```
#### When
```shell
mkpasswd -m md5 -S abcdefgh
```
#### Then
- exit code is `0`
- stdout contains `$1$abcdefgh$`
## mimixbox nologin
Source: `test/e2e/tools/mimixbox/loginutils/nologin.atago.yaml`
### Scenario: prints a refusal and exits non-zero
#### When
```shell
nologin
```
#### Then
- exit code is not `0`
- stdout contains `not available`
### Scenario: never runs a passed command
#### When
```shell
count=$(nologin -c "echo pwned" 2>/dev/null | grep -c pwned); echo "$count"
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox passwd
Source: `test/e2e/tools/mimixbox/loginutils/passwd.atago.yaml`
### Scenario: rejects conflicting flags
#### When
```shell
passwd -l -u alice
```
#### Then
- exit code is not `0`
### Scenario: describes itself with --help
#### When
```shell
passwd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: passwd`, `password`
## mimixbox run-init
Source: `test/e2e/tools/mimixbox/loginutils/run-init.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
run-init --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: run-init`
- stderr is empty
## mimixbox run-parts
Source: `test/e2e/tools/mimixbox/loginutils/run-parts.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
run-parts --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: run-parts`
- stderr is empty
## mimixbox run-init
Source: `test/e2e/tools/mimixbox/loginutils/run_init.atago.yaml`
### Scenario: requires NEW_ROOT and INIT
#### When
```shell
run-init /tmp
```
#### Then
- exit code is not `0`
### Scenario: rejects a non-directory NEW_ROOT
#### When
```shell
run-init /no/such/dir /init
```
#### Then
- exit code is not `0`
## mimixbox run-parts
Source: `test/e2e/tools/mimixbox/loginutils/run_parts.atago.yaml`
### Scenario: runs executables in alphabetical order
#### Given
- Fixture file `parts/20-b` is created.
- Fixture file `parts/10-a` is created.
#### Inputs
_Fixture `parts/20-b`:_
```
#!/bin/sh
echo B
```
_Fixture `parts/10-a`:_
```
#!/bin/sh
echo A
```
#### When
```shell
run-parts parts
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout equals an exact value
### Scenario: requires a directory
#### When
```shell
run-parts
```
#### Then
- exit code is not `0`
## mimixbox runlevel
Source: `test/e2e/tools/mimixbox/loginutils/runlevel.atago.yaml`
### Scenario: reports a runlevel or unknown
#### When
```shell
runlevel
```
#### Then
- stdout matches `/.+/`
## mimixbox start-stop-daemon
Source: `test/e2e/tools/mimixbox/loginutils/start-stop-daemon.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
start-stop-daemon --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: start-stop-daemon`
- stderr is empty
## mimixbox start-stop-daemon
Source: `test/e2e/tools/mimixbox/loginutils/start_stop_daemon.atago.yaml`
### Scenario: starts and stops a background program
#### When
```shell
p="${workdir}/foo.pid"
start-stop-daemon -S -p "$p" -x /bin/sleep -- 30 >/dev/null 2>&1
pid=$(cat "$p")
start-stop-daemon -K -p "$p" >/dev/null 2>&1
i=0
while [ "$i" -lt 50 ] && kill -0 "$pid" 2>/dev/null; do
    sleep 0.1
    i=$((i + 1))
done
if kill -0 "$pid" 2>/dev/null; then echo alive; else echo "stopped"; fi

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: requires a start or stop mode
#### When
```shell
start-stop-daemon -p /tmp/x
```
#### Then
- exit code is not `0`
## mimixbox su
Source: `test/e2e/tools/mimixbox/loginutils/su.atago.yaml`
### Scenario: fails for an unknown user
#### Inputs
_stdin for `su`:_
#### When
```shell
su nonexistent_user_xyz
```
#### Then
- exit code is not `0`
### Scenario: describes itself with --help
#### When
```shell
su --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: su`, `user`
## mimixbox sulogin
Source: `test/e2e/tools/mimixbox/loginutils/sulogin.atago.yaml`
### Scenario: rejects a wrong root password
#### Inputs
_stdin for `sulogin`:_
```
definitely_wrong_password
```
#### When
```shell
sulogin
```
#### Then
- exit code is not `0`
### Scenario: describes itself with --help
#### When
```shell
sulogin --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: sulogin`, `root`
## mimixbox vlock
Source: `test/e2e/tools/mimixbox/loginutils/vlock.atago.yaml`
### Scenario: fails on a wrong password
#### Inputs
_stdin for `vlock`:_
```
definitely_wrong_xyz
```
#### When
```shell
vlock
```
#### Then
- exit code is not `0`
### Scenario: describes itself with --help
#### When
```shell
vlock --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: vlock`, `terminal`
## mimixbox mailutils commands expose a dedicated --help helper
Source: `test/e2e/tools/mimixbox/mailutils/help_helpers_mailutils.atago.yaml`
### Scenario: makemime --help is structured
#### When
```shell
makemime --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: popmaildir --help is structured
#### When
```shell
popmaildir --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: reformime --help is structured
#### When
```shell
reformime --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: sendmail --help is structured
#### When
```shell
sendmail --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
## mimixbox makemime
Source: `test/e2e/tools/mimixbox/mailutils/makemime.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
makemime --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: makemime`
- stderr is empty
## mimixbox popmaildir
Source: `test/e2e/tools/mimixbox/mailutils/popmaildir.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
popmaildir --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: popmaildir`
- stderr is empty
## mimixbox reformime
Source: `test/e2e/tools/mimixbox/mailutils/reformime.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
reformime --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: reformime`
- stderr is empty
## mimixbox sendmail
Source: `test/e2e/tools/mimixbox/mailutils/sendmail.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
sendmail --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: sendmail`
- stderr is empty
## mimixbox arp
Source: `test/e2e/tools/mimixbox/netutils/arp.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
arp --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: arp`
- stderr is empty
## mimixbox arping
Source: `test/e2e/tools/mimixbox/netutils/arping.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
arping --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: arping`
- stderr is empty
## mimixbox brctl
Source: `test/e2e/tools/mimixbox/netutils/brctl.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
brctl --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: brctl`
- stderr is empty
## mimixbox dhcprelay
Source: `test/e2e/tools/mimixbox/netutils/dhcprelay.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
dhcprelay --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: dhcprelay`
- stderr is empty
## mimixbox dnsd
Source: `test/e2e/tools/mimixbox/netutils/dnsd.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
dnsd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: dnsd`
- stderr is empty
## mimixbox dnsdomainname
Source: `test/e2e/tools/mimixbox/netutils/dnsdomainname.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
dnsdomainname --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: dnsdomainname`
- stderr is empty
## mimixbox dumpleases
Source: `test/e2e/tools/mimixbox/netutils/dumpleases.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
dumpleases --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: dumpleases`
- stderr is empty
## mimixbox ether-wake
Source: `test/e2e/tools/mimixbox/netutils/ether-wake.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
ether-wake --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: ether-wake`
- stderr is empty
## mimixbox fakeidentd
Source: `test/e2e/tools/mimixbox/netutils/fakeidentd.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
fakeidentd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: fakeidentd`
- stderr is empty
## mimixbox ftpd
Source: `test/e2e/tools/mimixbox/netutils/ftpd.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
ftpd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: ftpd`
- stderr is empty
## mimixbox ftpget
Source: `test/e2e/tools/mimixbox/netutils/ftpget.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
ftpget --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: ftpget`
- stderr is empty
## mimixbox ftpput
Source: `test/e2e/tools/mimixbox/netutils/ftpput.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
ftpput --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: ftpput`
- stderr is empty
## mimixbox netutils --help helpers
Source: `test/e2e/tools/mimixbox/netutils/help_helpers_netutils.atago.yaml`
### Scenario: arp --help is structured
#### When
```shell
arp --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: arping --help is structured
#### When
```shell
arping --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: brctl --help is structured
#### When
```shell
brctl --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: dhcprelay --help is structured
#### When
```shell
dhcprelay --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: dnsd --help is structured
#### When
```shell
dnsd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: dnsdomainname --help is structured
#### When
```shell
dnsdomainname --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: dumpleases --help is structured
#### When
```shell
dumpleases --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: ether-wake --help is structured
#### When
```shell
ether-wake --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: fakeidentd --help is structured
#### When
```shell
fakeidentd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: ftpd --help is structured
#### When
```shell
ftpd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: ftpget --help is structured
#### When
```shell
ftpget --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: ftpput --help is structured
#### When
```shell
ftpput --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: http-status-code --help is structured
#### When
```shell
http-status-code --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: httpd --help is structured
#### When
```shell
httpd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: ifconfig --help is structured
#### When
```shell
ifconfig --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: ifdown --help is structured
#### When
```shell
ifdown --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: ifenslave --help is structured
#### When
```shell
ifenslave --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: ifplugd --help is structured
#### When
```shell
ifplugd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: ifup --help is structured
#### When
```shell
ifup --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: inetd --help is structured
#### When
```shell
inetd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: ip --help is structured
#### When
```shell
ip --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: ipaddr --help is structured
#### When
```shell
ipaddr --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: iplink --help is structured
#### When
```shell
iplink --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: ipneigh --help is structured
#### When
```shell
ipneigh --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: iproute --help is structured
#### When
```shell
iproute --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: iprule --help is structured
#### When
```shell
iprule --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: iptunnel --help is structured
#### When
```shell
iptunnel --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: nameif --help is structured
#### When
```shell
nameif --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: nbd-client --help is structured
#### When
```shell
nbd-client --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: netcat --help is structured
#### When
```shell
netcat --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: netstat --help is structured
#### When
```shell
netstat --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: nslookup --help is structured
#### When
```shell
nslookup --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: ntpd --help is structured
#### When
```shell
ntpd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: ping6 --help is structured
#### When
```shell
ping6 --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: pscan --help is structured
#### When
```shell
pscan --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: route --help is structured
#### When
```shell
route --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: slattach --help is structured
#### When
```shell
slattach --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: ssl_client --help is structured
#### When
```shell
ssl_client --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: ssl_server --help is structured
#### When
```shell
ssl_server --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: tc --help is structured
#### When
```shell
tc --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: tcpsvd --help is structured
#### When
```shell
tcpsvd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: telnet --help is structured
#### When
```shell
telnet --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: telnetd --help is structured
#### When
```shell
telnetd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: tftp --help is structured
#### When
```shell
tftp --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: tftpd --help is structured
#### When
```shell
tftpd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: traceroute --help is structured
#### When
```shell
traceroute --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: traceroute6 --help is structured
#### When
```shell
traceroute6 --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: tunctl --help is structured
#### When
```shell
tunctl --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: udhcpc --help is structured
#### When
```shell
udhcpc --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: udhcpc6 --help is structured
#### When
```shell
udhcpc6 --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: udhcpd --help is structured
#### When
```shell
udhcpd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: udpsvd --help is structured
#### When
```shell
udpsvd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: vconfig --help is structured
#### When
```shell
vconfig --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: whois --help is structured
#### When
```shell
whois --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: zcip --help is structured
#### When
```shell
zcip --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
## mimixbox http-status-code
Source: `test/e2e/tools/mimixbox/netutils/http-status-code.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
http-status-code --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: http-status-code`
- stderr is empty
### Scenario: looks up a status code by number
#### When
```shell
http-status-code search 404
```
#### Then
- exit code is `0`
- stdout contains `404 Not Found`
## mimixbox httpd
Source: `test/e2e/tools/mimixbox/netutils/httpd.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
httpd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: httpd`
- stderr is empty
## mimixbox http-status-code
Source: `test/e2e/tools/mimixbox/netutils/httpstatus.atago.yaml`
### Scenario: explains a status code
#### When
```shell
http-status-code search 404
```
#### Then
- exit code is `0`
- stdout contains `404 Not Found`
## mimixbox ifconfig
Source: `test/e2e/tools/mimixbox/netutils/ifconfig.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
ifconfig --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: ifconfig`
- stderr is empty
### Scenario: documents its purpose in --help
#### When
```shell
ifconfig --help
```
#### Then
- exit code is `0`
- stdout contains `network interfaces`
## mimixbox ifdown
Source: `test/e2e/tools/mimixbox/netutils/ifdown.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
ifdown --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: ifdown`
- stderr is empty
## mimixbox ifenslave
Source: `test/e2e/tools/mimixbox/netutils/ifenslave.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
ifenslave --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: ifenslave`
- stderr is empty
## mimixbox ifplugd
Source: `test/e2e/tools/mimixbox/netutils/ifplugd.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
ifplugd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: ifplugd`
- stderr is empty
## mimixbox inetd
Source: `test/e2e/tools/mimixbox/netutils/inetd.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
inetd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: inetd`
- stderr is empty
## mimixbox ipcalc
Source: `test/e2e/tools/mimixbox/netutils/ipcalc.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
ipcalc --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: ipcalc`
- stderr is empty
### Scenario: documents its purpose in --help
#### When
```shell
ipcalc --help
```
#### Then
- exit code is `0`
- stdout contains `IPv4 network parameters`
## mimixbox nc loopback
Source: `test/e2e/tools/mimixbox/netutils/nc.atago.yaml`
### Scenario: transfers data over TCP
#### When
```shell
recv="${workdir}/nc_recv.txt"
: > "$recv"
for attempt in $(seq 1 8); do
    port=$((18640 + attempt))
    (nc -l -p "$port" > "$recv") &
    lpid=$!
    for _ in $(seq 1 20); do
        echo "from-client" | nc 127.0.0.1 "$port" >/dev/null 2>&1
        if [ -s "$recv" ]; then
            break
        fi
        sleep 0.1
    done
    kill "$lpid" 2>/dev/null
    wait "$lpid" 2>/dev/null
    if [ -s "$recv" ]; then
        break
    fi
done
head -n 1 "$recv"

```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox netcat
Source: `test/e2e/tools/mimixbox/netutils/netcat.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
netcat --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: netcat`
- stderr is empty
## mimixbox netstat
Source: `test/e2e/tools/mimixbox/netutils/netstat.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
netstat --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: netstat`
- stderr is empty
## mimixbox nslookup
Source: `test/e2e/tools/mimixbox/netutils/nslookup.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
nslookup --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: nslookup`
- stderr is empty
## mimixbox ntpd
Source: `test/e2e/tools/mimixbox/netutils/ntpd.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
ntpd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: ntpd`
- stderr is empty
## mimixbox ping usage
Source: `test/e2e/tools/mimixbox/netutils/ping.atago.yaml`
### Scenario: reports an error when no host is given
#### When
```shell
ping
```
#### Then
- exit code is not `0`
## mimixbox pscan
Source: `test/e2e/tools/mimixbox/netutils/pscan.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
pscan --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: pscan`
- stderr is empty
## mimixbox route
Source: `test/e2e/tools/mimixbox/netutils/route.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
route --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: route`
- stderr is empty
## mimixbox slattach
Source: `test/e2e/tools/mimixbox/netutils/slattach.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
slattach --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: slattach`
- stderr is empty
## mimixbox ssl_client
Source: `test/e2e/tools/mimixbox/netutils/ssl_client.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
ssl_client --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: ssl_client`
- stderr is empty
## mimixbox ssl_server
Source: `test/e2e/tools/mimixbox/netutils/ssl_server.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
ssl_server --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: ssl_server`
- stderr is empty
## mimixbox tc
Source: `test/e2e/tools/mimixbox/netutils/tc.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
tc --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: tc`
- stderr is empty
## mimixbox tcpsvd
Source: `test/e2e/tools/mimixbox/netutils/tcpsvd.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
tcpsvd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: tcpsvd`
- stderr is empty
## mimixbox telnet
Source: `test/e2e/tools/mimixbox/netutils/telnet.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
telnet --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: telnet`
- stderr is empty
## mimixbox telnetd
Source: `test/e2e/tools/mimixbox/netutils/telnetd.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
telnetd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: telnetd`
- stderr is empty
## mimixbox tftp
Source: `test/e2e/tools/mimixbox/netutils/tftp.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
tftp --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: tftp`
- stderr is empty
## mimixbox tftpd
Source: `test/e2e/tools/mimixbox/netutils/tftpd.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
tftpd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: tftpd`
- stderr is empty
## mimixbox traceroute
Source: `test/e2e/tools/mimixbox/netutils/traceroute.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
traceroute --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: traceroute`
- stderr is empty
## mimixbox traceroute6
Source: `test/e2e/tools/mimixbox/netutils/traceroute6.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
traceroute6 --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: traceroute6`
- stderr is empty
## mimixbox tunctl
Source: `test/e2e/tools/mimixbox/netutils/tunctl.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
tunctl --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: tunctl`
- stderr is empty
## mimixbox udhcpc
Source: `test/e2e/tools/mimixbox/netutils/udhcpc.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
udhcpc --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: udhcpc`
- stderr is empty
## mimixbox udhcpc6
Source: `test/e2e/tools/mimixbox/netutils/udhcpc6.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
udhcpc6 --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: udhcpc6`
- stderr is empty
## mimixbox udhcpd
Source: `test/e2e/tools/mimixbox/netutils/udhcpd.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
udhcpd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: udhcpd`
- stderr is empty
## mimixbox udpsvd
Source: `test/e2e/tools/mimixbox/netutils/udpsvd.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
udpsvd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: udpsvd`
- stderr is empty
## mimixbox vconfig
Source: `test/e2e/tools/mimixbox/netutils/vconfig.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
vconfig --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: vconfig`
- stderr is empty
## mimixbox whois
Source: `test/e2e/tools/mimixbox/netutils/whois.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
whois --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: whois`
- stderr is empty
## mimixbox whris usage
Source: `test/e2e/tools/mimixbox/netutils/whris.atago.yaml`
### Scenario: reports an error when no domain is given
#### When
```shell
whris
```
#### Then
- exit code is not `0`
## mimixbox zcip
Source: `test/e2e/tools/mimixbox/netutils/zcip.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
zcip --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: zcip`
- stderr is empty
## mimixbox halt
Source: `test/e2e/tools/mimixbox/pmutils/halt.atago.yaml`
### Scenario: halt --help prints usage and lists the options
#### When
```shell
halt --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: halt`, `--poweroff`, `--wtmp-only`
### Scenario: poweroff --help prints usage
#### When
```shell
poweroff --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: poweroff`
### Scenario: reboot --help prints usage
#### When
```shell
reboot --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: reboot`
### Scenario: halt --version prints the version
#### When
```shell
halt --version
```
#### Then
- exit code is `0`
- stdout contains `halt (mimixbox)`
## mimixbox pmutils --help contract
Source: `test/e2e/tools/mimixbox/pmutils/help_helpers.atago.yaml`
### Scenario: poweroff --help is structured
#### When
```shell
poweroff --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: reboot --help is structured
#### When
```shell
reboot --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
## mimixbox printutils commands expose a dedicated --help helper
Source: `test/e2e/tools/mimixbox/printutils/help_helpers_printutils.atago.yaml`
### Scenario: lpd --help is structured
#### When
```shell
lpd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: lpq --help is structured
#### When
```shell
lpq --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: lpr --help is structured
#### When
```shell
lpr --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
## mimixbox depmod
Source: `test/e2e/tools/mimixbox/procps/depmod.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
depmod --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: depmod`
- stderr is empty
## mimixbox fuser
Source: `test/e2e/tools/mimixbox/procps/fuser.atago.yaml`
### Scenario: finds processes using the current directory
#### When
```shell
fuser .
```
#### Then
- exit code is `0`
- stdout matches `/[0-9]/`
### Scenario: exits non-zero when nothing uses the file
#### When
```shell
fuser /no/such/fuser/file
```
#### Then
- exit code is not `0`
## mimixbox procps --help contract
Source: `test/e2e/tools/mimixbox/procps/help_helpers_procps.atago.yaml`
### Scenario: depmod --help is structured
#### When
```shell
depmod --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: insmod --help is structured
#### When
```shell
insmod --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: lsmod --help is structured
#### When
```shell
lsmod --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: modinfo --help is structured
#### When
```shell
modinfo --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: modprobe --help is structured
#### When
```shell
modprobe --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: pkill --help is structured
#### When
```shell
pkill --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: pwdx --help is structured
#### When
```shell
pwdx --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: rmmod --help is structured
#### When
```shell
rmmod --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: uptime --help is structured
#### When
```shell
uptime --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
## mimixbox iostat
Source: `test/e2e/tools/mimixbox/procps/iostat.atago.yaml`
### Scenario: prints the avg-cpu header
#### When
```shell
iostat
```
#### Then
- exit code is `0`
- stdout contains `avg-cpu`, `%idle`
### Scenario: prints the device table header
#### When
```shell
iostat
```
#### Then
- exit code is `0`
- stdout contains `Device`
## mimixbox killall5
Source: `test/e2e/tools/mimixbox/procps/killall5.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
killall5 --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: killall5`, `signal`
## mimixbox klogd
Source: `test/e2e/tools/mimixbox/procps/klogd.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
klogd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: klogd`, `kernel`
## mimixbox logger
Source: `test/e2e/tools/mimixbox/procps/logger.atago.yaml`
### Scenario: rejects an unknown facility
#### When
```shell
logger -p nosuchfacility.info msg
```
#### Then
- exit code is not `0`
## mimixbox logread
Source: `test/e2e/tools/mimixbox/procps/logread.atago.yaml`
### Scenario: prints a given log file
#### Given
- Fixture file `sys.log` is created.
#### Inputs
_Fixture `sys.log`:_
```
msg one
msg two
```
#### When
```shell
logread sys.log
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout equals an exact value
### Scenario: fails when no readable log is found
#### When
```shell
logread nope.log
```
#### Then
- exit code is not `0`
## mimixbox lsmod
Source: `test/e2e/tools/mimixbox/procps/lsmod.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
lsmod --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: lsmod`
- stderr is empty
## mimixbox lsof
Source: `test/e2e/tools/mimixbox/procps/lsof.atago.yaml`
### Scenario: lists the working directory of a process
#### When
```shell
lsof -p $$
```
#### Then
- exit code is `0`
- stdout contains `cwd`
### Scenario: prints the column header
#### When
```shell
lsof -p $$
```
#### Then
- exit code is `0`
- stdout contains `COMMAND`
- stdout contains `NAME`
## mimixbox minips
Source: `test/e2e/tools/mimixbox/procps/minips.atago.yaml`
### Scenario: prints the PID/USER/COMMAND header
#### When
```shell
minips
```
#### Then
- exit code is `0`
- stdout contains `COMMAND`
### Scenario: lists processes
#### When
```shell
minips
```
#### Then
- exit code is `0`
- stdout matches `/(?m)^[0-9]+/`
## mimixbox modinfo
Source: `test/e2e/tools/mimixbox/procps/modinfo.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
modinfo --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: modinfo`
- stderr is empty
## mimixbox mpstat
Source: `test/e2e/tools/mimixbox/procps/mpstat.atago.yaml`
### Scenario: prints the CPU column header
#### When
```shell
mpstat
```
#### Then
- exit code is `0`
- stdout contains `%usr`, `%idle`
### Scenario: prints the aggregate all row
#### When
```shell
mpstat
```
#### Then
- exit code is `0`
- stdout matches `/(?m)^all /`
## mimixbox nmeter
Source: `test/e2e/tools/mimixbox/procps/nmeter.atago.yaml`
### Scenario: expands a literal percent and copies text
#### When
```shell
nmeter "hello %% world"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: expands the total-memory directive
#### When
```shell
nmeter "mem:%M"
```
#### Then
- exit code is `0`
- stdout matches `/mem:[0-9]+M/`
## mimixbox pgrep / pkill
Source: `test/e2e/tools/mimixbox/procps/pgrep.atago.yaml`
### Scenario: finds a running process by name
#### When
```shell
sleep 30 & p=$!; pgrep sleep | grep -c "^${p}$"; kill $p 2>/dev/null
```
#### Then
- exit code is `0`
- stdout contains `1`
### Scenario: exits non-zero when nothing matches
#### When
```shell
pgrep zzz_no_such_proc_zzz
```
#### Then
- exit code is not `0`
## mimixbox pmap
Source: `test/e2e/tools/mimixbox/procps/pmap.atago.yaml`
### Scenario: prints a total line for a process map
#### When
```shell
pmap $$
```
#### Then
- exit code is `0`
- stdout contains `total`
### Scenario: rejects an invalid PID
#### When
```shell
pmap notapid
```
#### Then
- exit code is not `0`
## mimixbox powertop
Source: `test/e2e/tools/mimixbox/procps/powertop.atago.yaml`
### Scenario: runs and exits zero
#### When
```shell
powertop
```
#### Then
- exit code is `0`
### Scenario: describes itself with --help
#### When
```shell
powertop --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: powertop`, `power`
## mimixbox ps
Source: `test/e2e/tools/mimixbox/procps/ps.atago.yaml`
### Scenario: prints the standard header
#### When
```shell
ps
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: lists running processes
#### When
```shell
ps
```
#### Then
- exit code is `0`
- stdout matches `/(?m)^ *[0-9]+ /`
## mimixbox pstree
Source: `test/e2e/tools/mimixbox/procps/pstree.atago.yaml`
### Scenario: builds a tree containing PID 1
#### When
```shell
pstree
```
#### Then
- exit code is `0`
- stdout contains `(1)`
## mimixbox pwdx
Source: `test/e2e/tools/mimixbox/procps/pwdx.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
pwdx --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: pwdx`
- stderr is empty
## mimixbox rmmod
Source: `test/e2e/tools/mimixbox/procps/rmmod.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
rmmod --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: rmmod`
- stderr is empty
## mimixbox smemcap
Source: `test/e2e/tools/mimixbox/procps/smemcap.atago.yaml`
### Scenario: captures a tar containing meminfo
#### When
```shell
smemcap > cap.tar; tar -tf cap.tar
```
#### Then
- exit code is `0`
- stdout matches `/(?m)^meminfo$/`
## mimixbox sysctl
Source: `test/e2e/tools/mimixbox/procps/sysctl.atago.yaml`
### Scenario: reads a kernel parameter
#### When
```shell
sysctl kernel.ostype
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: lists parameters with -a
#### When
```shell
sysctl -a
```
#### Then
- exit code is `0`
- stdout contains ` = `
## mimixbox syslogd
Source: `test/e2e/tools/mimixbox/procps/syslogd.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
syslogd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: syslogd`, `log`
## mimixbox top
Source: `test/e2e/tools/mimixbox/procps/top.atago.yaml`
### Scenario: prints the top summary line
#### When
```shell
top -bn1
```
#### Then
- exit code is `0`
- stdout matches `/^top -/`
### Scenario: prints the tasks line
#### When
```shell
top -bn1
```
#### Then
- exit code is `0`
- stdout matches `/(?m)^Tasks:/`
## mimixbox uptime
Source: `test/e2e/tools/mimixbox/procps/uptime.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
uptime --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: uptime`
- stderr is empty
### Scenario: prints an uptime/load line
#### When
```shell
uptime
```
#### Then
- exit code is `0`
- stdout contains `load average:`
## mimixbox uptime / pwdx
Source: `test/e2e/tools/mimixbox/procps/uptime_pwdx.atago.yaml`
### Scenario: uptime shows the load averages
#### When
```shell
uptime
```
#### Then
- exit code is `0`
- stdout contains `load average`
### Scenario: pwdx prints a process working directory
#### When
```shell
pwdx $$
```
#### Then
- exit code is `0`
- stdout matches `/^[0-9]+: //`
## mimixbox vmstat
Source: `test/e2e/tools/mimixbox/procps/vmstat.atago.yaml`
### Scenario: prints the column header
#### When
```shell
vmstat
```
#### Then
- exit code is `0`
- stdout contains `swpd`, `free`
### Scenario: prints a numeric data row
#### When
```shell
vmstat
```
#### Then
- exit code is `0`
- stdout matches `/^[ 0-9]+$/`
## mimixbox chpst
Source: `test/e2e/tools/mimixbox/runit/chpst.atago.yaml`
### Scenario: loads an environment directory
#### Given
- Fixture file `env/HELLO` is created.
#### Inputs
_Fixture `env/HELLO`:_
```
world
```
#### When
```shell
chpst -e env sh -c 'echo "$HELLO"'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: requires a program
#### When
```shell
chpst -o 64
```
#### Then
- exit code is not `0`
## mimixbox envdir
Source: `test/e2e/tools/mimixbox/runit/envdir.atago.yaml`
### Scenario: sets a variable from a directory file
#### Given
- Fixture file `env/GREETING` is created.
#### Inputs
_Fixture `env/GREETING`:_
```
hello
```
#### When
```shell
envdir env sh -c 'echo "$GREETING"'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: requires a directory and a program
#### When
```shell
envdir /only
```
#### Then
- exit code is not `0`
## mimixbox envuidgid
Source: `test/e2e/tools/mimixbox/runit/envuidgid.atago.yaml`
### Scenario: exports root uid and gid
#### When
```shell
envuidgid root sh -c 'echo "$UID:$GID"'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: fails for an unknown user
#### When
```shell
envuidgid nonexistent_xyz true
```
#### Then
- exit code is not `0`
## mimixbox runsv
Source: `test/e2e/tools/mimixbox/runit/runsv.atago.yaml`
### Scenario: requires a service directory
#### When
```shell
runsv
```
#### Then
- exit code is not `0`
### Scenario: describes itself with --help
#### When
```shell
runsv --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: runsv`, `service`
## mimixbox runsvdir
Source: `test/e2e/tools/mimixbox/runit/runsvdir.atago.yaml`
### Scenario: requires a services directory
#### When
```shell
runsvdir
```
#### Then
- exit code is not `0`
### Scenario: describes itself with --help
#### When
```shell
runsvdir --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: runsvdir`, `service`
## mimixbox setuidgid
Source: `test/e2e/tools/mimixbox/runit/setuidgid.atago.yaml`
### Scenario: fails for an unknown user
#### When
```shell
setuidgid nonexistent_user_xyz true
```
#### Then
- exit code is not `0`
### Scenario: requires a program
#### When
```shell
setuidgid root
```
#### Then
- exit code is not `0`
## mimixbox softlimit
Source: `test/e2e/tools/mimixbox/runit/softlimit.atago.yaml`
### Scenario: runs a program under the limits
#### When
```shell
softlimit -o 64 true
```
#### Then
- exit code is `0`
### Scenario: requires a program
#### When
```shell
softlimit -o 64
```
#### Then
- exit code is not `0`
## mimixbox sv
Source: `test/e2e/tools/mimixbox/runit/sv.atago.yaml`
### Scenario: writes the up control character
#### Given
- Fixture file `svc/supervise/control` is created.
- Fixture file `svc/supervise/ok` is created.
#### When
```shell
sv up svc && cat svc/supervise/control
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: reports a running service
#### Given
- Fixture file `svc2/supervise/ok` is created.
- Fixture file `svc2/supervise/pid` is created.
#### Inputs
_Fixture `svc2/supervise/pid`:_
```
99
```
#### When
```shell
sv status svc2
```
#### Then
- exit code is `0`
- stdout contains `run`
## mimixbox svc
Source: `test/e2e/tools/mimixbox/runit/svc.atago.yaml`
### Scenario: writes the down control character
#### Given
- Fixture file `svc/supervise/control` is created.
#### When
```shell
svc -d svc && cat svc/supervise/control
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: requires a control command
#### Given
- Fixture file `s2/supervise/control` is created.
#### When
```shell
svc s2
```
#### Then
- exit code is not `0`
## mimixbox svlogd
Source: `test/e2e/tools/mimixbox/runit/svlogd.atago.yaml`
### Scenario: appends stdin to the current log
#### When
```shell
printf 'hello\nworld\n' | svlogd log
cat log/current

```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout equals an exact value
### Scenario: requires a directory
#### Inputs
_stdin for `svlogd`:_
```
x
```
#### When
```shell
svlogd
```
#### Then
- exit code is not `0`
## mimixbox svok
Source: `test/e2e/tools/mimixbox/runit/svok.atago.yaml`
### Scenario: succeeds for a supervised service
#### Given
- Fixture file `svc/supervise/ok` is created.
#### When
```shell
svok svc
```
#### Then
- exit code is `0`
### Scenario: returns 100 for an unsupervised service
#### Given
- Fixture file `down/.keep` is created.
#### When
```shell
svok down
```
#### Then
- exit code is `100`
## mimixbox chcon
Source: `test/e2e/tools/mimixbox/securityutils/chcon.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
chcon --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: chcon`
- stderr is empty
## mimixbox getenforce
Source: `test/e2e/tools/mimixbox/securityutils/getenforce.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
getenforce --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: getenforce`
- stderr is empty
## mimixbox getsebool
Source: `test/e2e/tools/mimixbox/securityutils/getsebool.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
getsebool --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: getsebool`
- stderr is empty
## mimixbox securityutils --help contract
Source: `test/e2e/tools/mimixbox/securityutils/help_helpers.atago.yaml`
### Scenario: chcon --help is structured
#### When
```shell
chcon --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: getenforce --help is structured
#### When
```shell
getenforce --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: getsebool --help is structured
#### When
```shell
getsebool --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: load_policy --help is structured
#### When
```shell
load_policy --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: matchpathcon --help is structured
#### When
```shell
matchpathcon --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: restorecon --help is structured
#### When
```shell
restorecon --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: runcon --help is structured
#### When
```shell
runcon --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: selinuxenabled --help is structured
#### When
```shell
selinuxenabled --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: sestatus --help is structured
#### When
```shell
sestatus --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: setenforce --help is structured
#### When
```shell
setenforce --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: setfiles --help is structured
#### When
```shell
setfiles --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: setsebool --help is structured
#### When
```shell
setsebool --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: zip-pwcrack --help is structured
#### When
```shell
zip-pwcrack --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
## mimixbox pwcrack
Source: `test/e2e/tools/mimixbox/securityutils/pwcrack.atago.yaml`
### Scenario: finds a weak password in the wordlist
#### Given
- Fixture file `words.txt` is created.
#### Inputs
_Fixture `words.txt`:_
```
alpha
secret
beta
```
#### When
```shell
pwcrack -w words.txt $6$abcdefgh$ltjgWl6579NluT/Vi1nwEvcil.G5Nbc4NiXZaNGStk8PSwGfQv72N2CKPPrVACtLtip/cZ/1GM/O6IND4WQhG.
```
#### Then
- exit code is `0`
- stdout contains `: secret`
## mimixbox pwgen
Source: `test/e2e/tools/mimixbox/securityutils/pwgen.atago.yaml`
### Scenario: generates the requested number of passwords
#### When
```shell
pwgen -n 3 -l 8 | wc -l | tr -d ' '
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox pwscore
Source: `test/e2e/tools/mimixbox/securityutils/pwscore.atago.yaml`
### Scenario: scores a common password as zero
#### When
```shell
pwscore password | head -n 1
```
#### Then
- exit code is `0`
- stdout contains `Score: 0/100`
## mimixbox runcon
Source: `test/e2e/tools/mimixbox/securityutils/runcon.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
runcon --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: runcon`
- stderr is empty
## mimixbox selinuxenabled
Source: `test/e2e/tools/mimixbox/securityutils/selinuxenabled.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
selinuxenabled --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: selinuxenabled`
- stderr is empty
## mimixbox sestatus
Source: `test/e2e/tools/mimixbox/securityutils/sestatus.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
sestatus --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: sestatus`
- stderr is empty
## mimixbox setenforce
Source: `test/e2e/tools/mimixbox/securityutils/setenforce.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
setenforce --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: setenforce`
- stderr is empty
## mimixbox setfiles
Source: `test/e2e/tools/mimixbox/securityutils/setfiles.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
setfiles --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: setfiles`
- stderr is empty
## mimixbox setsebool
Source: `test/e2e/tools/mimixbox/securityutils/setsebool.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
setsebool --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: setsebool`
- stderr is empty
## mimixbox unshadow
Source: `test/e2e/tools/mimixbox/securityutils/unshadow.atago.yaml`
### Scenario: merges the shadow hash into the passwd line
#### Given
- Fixture file `passwd` is created.
- Fixture file `shadow` is created.
#### Inputs
_Fixture `passwd`:_
```
alice:x:1000:1000:Alice:/home/alice:/bin/sh
```
_Fixture `shadow`:_
```
alice:$6$abc$HASH:19000:0:99999:7:::
```
#### When
```shell
unshadow passwd shadow
```
#### Then
- exit code is `0`
- stdout contains `alice:$6$abc$HASH:1000:1000`
## mimixbox zip-pwcrack
Source: `test/e2e/tools/mimixbox/securityutils/zip-pwcrack.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
zip-pwcrack --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: zip-pwcrack`
- stderr is empty
## mimixbox zip-pwcrack
Source: `test/e2e/tools/mimixbox/securityutils/zippwcrack.atago.yaml`
### Scenario: recovers the ZIP password from the wordlist
#### Given
- Fixture file `words.txt` is created.
- Fixture file `enc.zip` is created.
#### Inputs
_Fixture `words.txt`:_
```
alpha
hunter2
beta
```
#### When
```shell
zip-pwcrack enc.zip -w words.txt
```
#### Then
- exit code is `0`
- stdout contains `password found: hunter2`
## mimixbox arch
Source: `test/e2e/tools/mimixbox/shellutils/arch.atago.yaml`
### Scenario: prints a non-empty machine name
#### When
```shell
arch
```
#### Then
- exit code is `0`
- stdout is not empty
## mimixbox base64
Source: `test/e2e/tools/mimixbox/shellutils/base64.atago.yaml`
### Scenario: encodes standard input
#### When
```shell
printf 'hello\n' | base64
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: encodes the file contents
#### Given
- Fixture file `base64.txt` is created.
#### Inputs
_Fixture `base64.txt`:_
```
hello
```
#### When
```shell
base64 base64.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: decodes standard input
#### When
```shell
printf 'aGVsbG8K\n' | base64 -d
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: returns the original text (round trip)
#### When
```shell
printf 'MimixBox\n' | base64 | base64 -d
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: reports an error for a non-existent file
#### When
```shell
base64 /no_exist_file
```
#### Then
- exit code is not `0`
- stderr equals an exact value
## mimixbox basename
Source: `test/e2e/tools/mimixbox/shellutils/basename.atago.yaml`
### Scenario: show test.txt
#### When
```shell
basename "/home/nao/test.txt"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: show test
#### When
```shell
basename "/home/nao/test"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: show .test
#### When
```shell
basename "/home/nao/.test"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: show nao for a trailing slash
#### When
```shell
basename "/home/nao/"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: show error without operand
#### When
```shell
basename
```
#### Then
- exit code is not `0`
- stderr equals an exact value
### Scenario: show / for root
#### When
```shell
basename "/"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: show empty string
#### When
```shell
basename ""
```
#### Then
- exit code is `0`
- stdout is empty
### Scenario: show error for extra operand
#### When
```shell
basename /bin/basename /home/nao /home
```
#### Then
- exit code is not `0`
- stderr equals an exact value
### Scenario: show three basenames with -a
#### When
```shell
basename -a /bin/basename /home/nao /home
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
basename
nao
home
```
### Scenario: show three basenames joined with -a -z
#### When
```shell
basename -a -z /bin/basename /home/nao /home
```
#### Then
- exit code is `0`
- stdout matches `/^basename\x00nao\x00home\x00$/`
### Scenario: show basename without the suffix
#### When
```shell
basename -s .txt /home/nao/test.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: show basename built from an environment variable
#### When
```shell
TEST_DIR="/aaa/bbb/ccc"; basename "$TEST_DIR/ddd.txt"
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox bc
Source: `test/e2e/tools/mimixbox/shellutils/bc.atago.yaml`
### Scenario: respects operator precedence
#### When
```shell
echo '2 + 3 * 4' | bc
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: honors scale for division
#### When
```shell
echo 'scale=2; 7/3' | bc
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: supports variables
#### When
```shell
printf 'x = 5\nx * x\n' | bc
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: computes powers
#### When
```shell
echo '2^10' | bc
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox cal
Source: `test/e2e/tools/mimixbox/shellutils/cal.atago.yaml`
### Scenario: prints the month calendar
#### When
```shell
cal 11 2023
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
   November 2023
Su Mo Tu We Th Fr Sa
          1  2  3  4
 5  6  7  8  9 10 11
12 13 14 15 16 17 18
19 20 21 22 23 24 25
26 27 28 29 30
```
## mimixbox chmod
Source: `test/e2e/tools/mimixbox/shellutils/chmod.atago.yaml`
### Scenario: sets the permission bits with an octal mode
#### When
```shell
mkdir -p chmod && touch chmod/file.txt && chmod 600 chmod/file.txt && chmod 644 chmod/file.txt && stat -c '%a' chmod/file.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: adds owner execute to mode 600 with a symbolic mode
#### When
```shell
mkdir -p chmod && touch chmod/file.txt && chmod 600 chmod/file.txt && chmod u+x chmod/file.txt && stat -c '%a' chmod/file.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: reports an error on a missing file
#### When
```shell
chmod 644 /no_such_file
```
#### Then
- exit code is not `0`
- stderr equals an exact value
#### Expected output
_expected stderr:_
```
chmod: cannot access '/no_such_file': no such file or directory
```
## mimixbox chroot
Source: `test/e2e/tools/mimixbox/shellutils/chroot.atago.yaml`
### Scenario: prints usage with --help and exits 0
#### When
```shell
chroot --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: chroot`
### Scenario: documents the --userspec identity option in --help
#### When
```shell
chroot --help
```
#### Then
- exit code is `0`
- stdout contains `--userspec`
### Scenario: fails with a message when given no operand
#### When
```shell
chroot
```
#### Then
- exit code is not `0`
- stderr contains `chroot`
## mimixbox cmp
Source: `test/e2e/tools/mimixbox/shellutils/cmp.atago.yaml`
### Scenario: prints nothing and succeeds on identical files
#### Given
- Fixture file `cmp/a.txt` is created.
- Fixture file `cmp/same.txt` is created.
#### Inputs
_Fixture `cmp/a.txt`:_
```
abc
```
_Fixture `cmp/same.txt`:_
```
abc
```
#### When
```shell
cmp cmp/a.txt cmp/same.txt; echo "rc=$?"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: reports the first differing byte and line on differing files
#### Given
- Fixture file `cmp/a.txt` is created.
- Fixture file `cmp/diff.txt` is created.
#### Inputs
_Fixture `cmp/a.txt`:_
```
abc
```
_Fixture `cmp/diff.txt`:_
```
abd
```
#### When
```shell
cmp ${workdir}/cmp/a.txt ${workdir}/cmp/diff.txt
```
#### Then
- exit code is not `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
${workdir}/cmp/a.txt ${workdir}/cmp/diff.txt differ: byte 3, line 1
```
### Scenario: cmp -s prints nothing but exits non-zero
#### Given
- Fixture file `cmp/a.txt` is created.
- Fixture file `cmp/diff.txt` is created.
#### Inputs
_Fixture `cmp/a.txt`:_
```
abc
```
_Fixture `cmp/diff.txt`:_
```
abd
```
#### When
```shell
cmp -s cmp/a.txt cmp/diff.txt; echo "rc=$?"
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox cmp (GNU options)
Source: `test/e2e/tools/mimixbox/shellutils/cmp_gnu.atago.yaml`
### Scenario: -n reports equality when the difference is past the byte limit
#### Given
- Fixture file `cmp_gnu/a.txt` is created.
- Fixture file `cmp_gnu/b.txt` is created.
#### Inputs
_Fixture `cmp_gnu/a.txt`:_
```
abcXdef
```
_Fixture `cmp_gnu/b.txt`:_
```
abcYdef
```
#### When
```shell
cmp -n 3 cmp_gnu/a.txt cmp_gnu/b.txt; echo "rc=$?"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: --bytes reports the difference within the byte limit
#### Given
- Fixture file `cmp_gnu/a.txt` is created.
- Fixture file `cmp_gnu/b.txt` is created.
#### Inputs
_Fixture `cmp_gnu/a.txt`:_
```
abcXdef
```
_Fixture `cmp_gnu/b.txt`:_
```
abcYdef
```
#### When
```shell
cmp --bytes=4 ${workdir}/cmp_gnu/a.txt ${workdir}/cmp_gnu/b.txt
```
#### Then
- exit code is not `0`
- stdout contains `differ: byte 4, line 1`
### Scenario: -i skips the first N bytes of both files
#### Given
- Fixture file `cmp_gnu/skip_a.txt` is created.
- Fixture file `cmp_gnu/skip_b.txt` is created.
#### Inputs
_Fixture `cmp_gnu/skip_a.txt`:_
```
XXXcommon
```
_Fixture `cmp_gnu/skip_b.txt`:_
```
YYYcommon
```
#### When
```shell
cmp -i 3 cmp_gnu/skip_a.txt cmp_gnu/skip_b.txt; echo "rc=$?"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: -i N:M skips N bytes of file1 and M of file2
#### Given
- Fixture file `cmp_gnu/pair_a.txt` is created.
- Fixture file `cmp_gnu/pair_b.txt` is created.
#### Inputs
_Fixture `cmp_gnu/pair_a.txt`:_
```
_abZ
```
_Fixture `cmp_gnu/pair_b.txt`:_
```
___abQ
```
#### When
```shell
cmp -i 1:3 ${workdir}/cmp_gnu/pair_a.txt ${workdir}/cmp_gnu/pair_b.txt
```
#### Then
- exit code is not `0`
- stdout contains `differ: byte 3, line 1`
### Scenario: -b prints the differing byte values in the message
#### Given
- Fixture file `cmp_gnu/pb_a.txt` is created.
- Fixture file `cmp_gnu/pb_b.txt` is created.
#### Inputs
_Fixture `cmp_gnu/pb_a.txt`:_
```
first
second
```
_Fixture `cmp_gnu/pb_b.txt`:_
```
first
SECOND
```
#### When
```shell
cmp -b ${workdir}/cmp_gnu/pb_a.txt ${workdir}/cmp_gnu/pb_b.txt
```
#### Then
- exit code is not `0`
- stdout contains `differ: byte 7, line 2 is 163 s 123 S`
## mimixbox cp (permission preservation)
Source: `test/e2e/tools/mimixbox/shellutils/cp_perm.atago.yaml`
### Scenario: keeps the source file mode (execute bit)
#### When
```shell
printf '#!/bin/sh\necho hi\n' > script.sh && chmod 755 script.sh && cp script.sh copy.sh && stat -c '%a' copy.sh
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: keeps a private directory mode
#### When
```shell
mkdir -m 700 private && cp -r private private_copy && stat -c '%a' private_copy
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: overwrites a read-only destination with -f
#### When
```shell
printf 'old\n' > dst.txt && chmod 444 dst.txt && printf 'new\n' > src.txt && cp -f src.txt dst.txt && cat dst.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox cut
Source: `test/e2e/tools/mimixbox/shellutils/cut.atago.yaml`
### Scenario: prints the chosen field
#### When
```shell
printf 'a,b,c\n' | cut -f 2 -d ,
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: prints the chosen fields joined by the delimiter
#### When
```shell
printf 'a,b,c,d\n' | cut -f 1,3 -d ,
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: prints from the field to the end
#### When
```shell
printf 'a,b,c,d\n' | cut -f 2- -d ,
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: prints the chosen character range
#### When
```shell
printf 'abcdef\n' | cut -c 1-3
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: reports an error without a list
#### When
```shell
printf 'a,b\n' | cut -d ,
```
#### Then
- exit code is not `0`
- stderr equals an exact value
## mimixbox cut (GNU options)
Source: `test/e2e/tools/mimixbox/shellutils/cut_gnu.atago.yaml`
### Scenario: --complement keeps the fields not selected
#### When
```shell
printf 'a,b,c\n' | cut -f 2 -d , --complement
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: --complement keeps the bytes not selected
#### When
```shell
printf 'abcde\n' | cut -b 2-3 --complement
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: -z splits and joins records on NUL (fields)
#### When
```shell
printf 'a,b,c\000d,e,f\000' | cut -f 2 -d , -z | tr '\000' '|'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: -z cuts bytes from each NUL-delimited record
#### When
```shell
printf 'abc\000def\000' | cut -b 1-2 -z | tr '\000' '|'
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox date
Source: `test/e2e/tools/mimixbox/shellutils/date.atago.yaml`
### Scenario: formats the date portion of an epoch
#### When
```shell
date -u -d @0 +%F
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: formats the time portion of an epoch
#### When
```shell
date -u -d @0 +%T
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: prints a literal percent sign
#### When
```shell
date +%%
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: prints a four-digit year
#### When
```shell
date +%Y | grep -E '^[0-9]{4}$' > /dev/null && echo "ok"
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox dc
Source: `test/e2e/tools/mimixbox/shellutils/dc.atago.yaml`
### Scenario: performs integer division
#### When
```shell
echo '6 3 / p' | dc
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: honors the precision register
#### When
```shell
echo '2k 7 3 / p' | dc
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: evaluates -e expressions
#### When
```shell
dc -e '2 10 ^ p'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: stores and loads registers
#### When
```shell
echo '5 sa 3 la + p' | dc
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox dd
Source: `test/e2e/tools/mimixbox/shellutils/dd.atago.yaml`
### Scenario: reproduces the input (stdin to stdout)
#### When
```shell
printf 'hello world' | dd status=none
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: copies only the requested blocks with count
#### When
```shell
printf 'hello world' | dd bs=1 count=5 status=none
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: conv=ucase upper-cases the data
#### When
```shell
printf 'abc' | dd conv=ucase status=none
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox df
Source: `test/e2e/tools/mimixbox/shellutils/df.atago.yaml`
### Scenario: shows the column header
#### When
```shell
df . | head -n 1
```
#### Then
- exit code is `0`
- stdout contains `Filesystem`
### Scenario: exits zero for the current directory
#### When
```shell
df . > /dev/null; echo "rc=$?"
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox df GNU flags
Source: `test/e2e/tools/mimixbox/shellutils/df_gnu.atago.yaml`
### Scenario: --output prints the selected column headers in order
#### When
```shell
df --output=source,fstype,size,used,avail,pcent,target
```
#### Then
- exit code is `0`
- stdout contains `Filesystem`
- stdout contains `Type`
- stdout contains `Size`
- stdout contains `Used`
- stdout contains `Avail`
- stdout contains `Use%`
- stdout contains `Mounted on`
### Scenario: --output honors a reordered field list
#### When
```shell
df --output=target,source
```
#### Then
- exit code is `0`
- stdout matches `/Mounted on.*Filesystem.*/`
### Scenario: --output rejects an unknown field
#### When
```shell
df --output=bogus
```
#### Then
- exit code is not `0`
- stderr contains `bogus`
### Scenario: --total emits a row labeled total
#### When
```shell
df --total --output=source,size,used,avail,target
```
#### Then
- exit code is `0`
- stdout contains `total`
### Scenario: --total works with the classic layout too
#### When
```shell
df --total
```
#### Then
- exit code is `0`
- stdout contains `total`
### Scenario: --type accepts a type filter and exits cleanly
#### When
```shell
df --type=tmpfs --output=fstype,target
```
#### Then
- exit code is `0`
- stdout contains `Type`
### Scenario: --type is repeatable
#### When
```shell
df -t tmpfs -t ext4 --output=fstype
```
#### Then
- exit code is `0`
- stdout contains `Type`
### Scenario: --block-size labels the block-size in the classic header
#### When
```shell
df --block-size=1M
```
#### Then
- exit code is `0`
- stdout contains `1048576-blocks`
### Scenario: --block-size rejects an invalid size
#### When
```shell
df --block-size=1Z
```
#### Then
- exit code is not `0`
- stderr contains `block-size`
### Scenario: --all lists at least as many rows with -a as without
#### When
```shell
test "$(df -a --output=target | wc -l)" -ge "$(df --output=target | wc -l)"
```
#### Then
- exit code is `0`
## mimixbox dirname
Source: `test/e2e/tools/mimixbox/shellutils/dirname.atago.yaml`
### Scenario: print /home/nao for an absolute file path
#### When
```shell
dirname "/home/nao/test.txt"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: print /home/nao for a filename without extension
#### When
```shell
dirname "/home/nao/test"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: print /home/nao for a hidden file
#### When
```shell
dirname "/home/nao/.test"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: print error without operand
#### When
```shell
dirname
```
#### Then
- exit code is not `0`
- stderr equals an exact value
### Scenario: print / for the root directory
#### When
```shell
dirname "/"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: print . for an empty string
#### When
```shell
dirname ""
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: print /bin /home / with line feed for three arguments
#### When
```shell
dirname /bin/dirname /home/nao /home
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
/bin
/home
/
```
### Scenario: print NUL-separated dirnames for three arguments with -z
#### When
```shell
dirname -z /bin/dirname /home/nao /home
```
#### Then
- exit code is `0`
- stdout matches `/^/bin\x00/home\x00/\x00$/`
### Scenario: print /aaa/bbb/ccc built from an environment variable
#### When
```shell
TEST_DIR="/aaa/bbb/ccc"; dirname $TEST_DIR/ddd.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox du
Source: `test/e2e/tools/mimixbox/shellutils/du.atago.yaml`
### Scenario: -b reports the total apparent byte size
#### When
```shell
mkdir -p du/sub && printf '%0.s.' $(seq 1 100) > du/a.txt && printf '%0.s.' $(seq 1 50) > du/sub/b.txt && du -s -b ${workdir}/du
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: -s reports the total in 1K blocks
#### When
```shell
mkdir -p du/sub && printf '%0.s.' $(seq 1 100) > du/a.txt && printf '%0.s.' $(seq 1 50) > du/sub/b.txt && du -s ${workdir}/du
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox du GNU flags
Source: `test/e2e/tools/mimixbox/shellutils/du_gnu.atago.yaml`
### Scenario: omits directories deeper than --max-depth
#### When
```shell
mkdir -p sub/deep && head -c 1000 /dev/zero > a.txt && head -c 2000 /dev/zero > sub/b.txt && head -c 3000 /dev/zero > sub/deep/c.txt && du --max-depth=1 "${workdir}" | sed "s#${workdir}#ROOT#"

```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout equals an exact value
### Scenario: prints only the operand total with --max-depth=0
#### When
```shell
mkdir -p sub/deep && head -c 1000 /dev/zero > a.txt && head -c 2000 /dev/zero > sub/b.txt && head -c 3000 /dev/zero > sub/deep/c.txt && du --max-depth=0 "${workdir}" | sed "s#${workdir}#ROOT#"

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: skips entries matching --exclude
#### When
```shell
mkdir -p sub/deep && head -c 1000 /dev/zero > a.txt && head -c 2000 /dev/zero > sub/b.txt && head -c 3000 /dev/zero > sub/deep/c.txt && du --exclude='sub' "${workdir}" | sed "s#${workdir}#ROOT#"

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: skips glob-matching files under -a
#### When
```shell
mkdir -p sub/deep && head -c 1000 /dev/zero > a.txt && head -c 2000 /dev/zero > sub/b.txt && head -c 3000 /dev/zero > sub/deep/c.txt && head -c 4096 /dev/zero > drop.tmp && du -a --exclude='*.tmp' "${workdir}" | grep -c 'drop.tmp' || true

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: reports exact bytes with --apparent-size
#### When
```shell
mkdir -p sub/deep && head -c 1000 /dev/zero > a.txt && head -c 2000 /dev/zero > sub/b.txt && head -c 3000 /dev/zero > sub/deep/c.txt && du -s --apparent-size "${workdir}" | sed "s#${workdir}#ROOT#"

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: reports block counts by default
#### When
```shell
mkdir -p sub/deep && head -c 1000 /dev/zero > a.txt && head -c 2000 /dev/zero > sub/b.txt && head -c 3000 /dev/zero > sub/deep/c.txt && du -s "${workdir}" | sed "s#${workdir}#ROOT#"

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: matches a plain run on a single filesystem with -x
#### When
```shell
mkdir -p sub/deep && head -c 1000 /dev/zero > a.txt && head -c 2000 /dev/zero > sub/b.txt && head -c 3000 /dev/zero > sub/deep/c.txt && du -x -s "${workdir}" | sed "s#${workdir}#ROOT#"

```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox echo
Source: `test/e2e/tools/mimixbox/shellutils/echo.atago.yaml`
### Scenario: says Hello World!
#### When
```shell
echo "Hello World!"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: says Hello World! (helper ignores the positional argument)
#### When
```shell
echo "Hello World!"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: expands an environment variable
#### When
```shell
export TEST_ENV="TEST_ENV_VAR"; echo ${TEST_ENV}
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: pipes data through xargs
#### When
```shell
echo "pipe" | xargs echo
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: says nothing with no arguments
#### When
```shell
echo
```
#### Then
- exit code is `0`
- stdout is empty
### Scenario: redirects data to a file and shows it
#### When
```shell
echo "MimixBox" > echo.txt && cat echo.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: --help as the first argument prints usage instead of the literal text
#### When
```shell
echo --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: echo`
### Scenario: --version as the first argument prints the version line
#### When
```shell
echo --version
```
#### Then
- exit code is `0`
- stdout contains `echo (mimixbox)`
### Scenario: --help that is not the first argument stays literal
#### When
```shell
echo foo --help
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox env
Source: `test/e2e/tools/mimixbox/shellutils/env.atago.yaml`
### Scenario: adds the assignment to the printed environment
#### When
```shell
env FOO=bar | grep '^FOO=bar$'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: -i prints only the given assignment
#### When
```shell
env -i ONLY=here
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: passes the variable to the run command
#### When
```shell
env GREETING=hi sh -c 'echo $GREETING'
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox env GNU flags
Source: `test/e2e/tools/mimixbox/shellutils/env_gnu.atago.yaml`
### Scenario: --chdir reports the chdir target via pwd (long flag with =)
#### When
```shell
env --chdir=/tmp pwd
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: --chdir reports the chdir target via pwd (-C short flag)
#### When
```shell
mkdir -p sub && env -C "${workdir}/sub" pwd
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: --chdir fails when the directory does not exist
#### When
```shell
env --chdir=${workdir}/nope pwd
```
#### Then
- exit code is `125`
- stderr contains `cannot change directory`
### Scenario: --split-string splits the command and its arguments (-S)
#### When
```shell
env -S 'printf %s-%s a b'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: --split-string splits with the long flag and whitespace runs
#### When
```shell
env --split-string='printf   %s   hi'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: --ignore-signal accepts known names and still runs the command
#### When
```shell
env --ignore-signal=INT,TERM printf '%s' ok
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: --ignore-signal rejects an unknown signal name
#### When
```shell
env --ignore-signal=BOGUS printf x
```
#### Then
- exit code is `125`
- stderr contains `invalid signal`
## mimixbox expr
Source: `test/e2e/tools/mimixbox/shellutils/expr.atago.yaml`
### Scenario: adds two numbers
#### When
```shell
expr 6 + 7
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: multiplies two numbers
#### When
```shell
expr 3 \* 4
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: respects parentheses
#### When
```shell
expr \( 1 + 2 \) \* 3
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: prints the string length
#### When
```shell
expr length abcd
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: prints 0 and exits non-zero for a zero result
#### When
```shell
expr 0
```
#### Then
- exit code is not `0`
- stdout equals an exact value
## mimixbox factor
Source: `test/e2e/tools/mimixbox/shellutils/factor.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
factor --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: factor`
- stderr is empty
### Scenario: documents its purpose in --help
#### When
```shell
factor --help
```
#### Then
- exit code is `0`
- stdout contains `Print the prime factors`
### Scenario: factors a small integer
#### When
```shell
factor 12
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: fails on a non-numeric operand
#### When
```shell
factor notanumber
```
#### Then
- exit code is not `0`
- stderr contains `factor:`
## mimixbox false
Source: `test/e2e/tools/mimixbox/shellutils/false.atago.yaml`
### Scenario: prints nothing and exits 1
#### When
```shell
false
```
#### Then
- exit code is not `0`
- stdout is empty
## mimixbox free
Source: `test/e2e/tools/mimixbox/shellutils/free.atago.yaml`
### Scenario: prints the column header
#### When
```shell
free | head -n 1
```
#### Then
- exit code is `0`
- stdout contains `total`, `available`
## mimixbox fsync
Source: `test/e2e/tools/mimixbox/shellutils/fsync.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
fsync --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: fsync`
- stderr is empty
## mimixbox ghrdc
Source: `test/e2e/tools/mimixbox/shellutils/ghrdc.atago.yaml`
### Scenario: prints usage with --help and exits 0
#### When
```shell
ghrdc --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: ghrdc`
### Scenario: fails with a message when given no operand
#### When
```shell
ghrdc
```
#### Then
- exit code is not `0`
- stderr contains `ghrdc`
## mimixbox groups
Source: `test/e2e/tools/mimixbox/shellutils/groups.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
groups --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: groups`
- stderr is empty
### Scenario: prints the groups of a named user
#### When
```shell
groups root
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox gzip
Source: `test/e2e/tools/mimixbox/shellutils/gzip.atago.yaml`
### Scenario: compresses and decompresses back to the original
#### Given
- Fixture file `g.txt` is created.
#### Inputs
_Fixture `g.txt`:_
```
hello gzip roundtrip
```
#### When
```shell
gzip ${workdir}/g.txt && gunzip ${workdir}/g.txt.gz && cat ${workdir}/g.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox shellutils --help helpers
Source: `test/e2e/tools/mimixbox/shellutils/help_helpers_shellutils.atago.yaml`
### Scenario: fsync --help is structured
#### When
```shell
env -- fsync --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: log-collect --help is structured
#### When
```shell
env -- log-collect --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: sddf --help is structured
#### When
```shell
env -- sddf --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: time --help is structured
#### When
```shell
env -- time --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: usleep --help is structured
#### When
```shell
env -- usleep --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
## mimixbox hermetic harness
Source: `test/e2e/tools/mimixbox/shellutils/hermetic.atago.yaml`
### Scenario: resolves cat to the MimixBox binary, not the host command
#### When
```shell
path=$(command -v cat) && basename "$(readlink -f "${path}")"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: resolves head to the MimixBox binary, not the host command
#### When
```shell
path=$(command -v head) && basename "$(readlink -f "${path}")"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: resolves base64 to the MimixBox binary, not the host command
#### When
```shell
path=$(command -v base64) && basename "$(readlink -f "${path}")"
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox hostid
Source: `test/e2e/tools/mimixbox/shellutils/hostid.atago.yaml`
### Scenario: prints 8 hexadecimal digits
#### When
```shell
hostid
```
#### Then
- exit code is `0`
- stdout matches `/^[0-9a-f]{8}/`
### Scenario: prints the same value on repeated calls
#### When
```shell
a=$(hostid); b=$(hostid); test "$a" = "$b" && echo "stable"
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox hostname
Source: `test/e2e/tools/mimixbox/shellutils/hostname.atago.yaml`
### Scenario: prints a non-empty host name
#### When
```shell
hostname
```
#### Then
- exit code is `0`
- stdout is not empty
## mimixbox id
Source: `test/e2e/tools/mimixbox/shellutils/id.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
id --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: id`
- stderr is empty
### Scenario: prints the current uid/gid line
#### When
```shell
id
```
#### Then
- exit code is `0`
- stdout contains `uid=`, `gid=`
## mimixbox install
Source: `test/e2e/tools/mimixbox/shellutils/install.atago.yaml`
### Scenario: copies the file content
#### Given
- Fixture file `install/src` is created.
#### Inputs
_Fixture `install/src`:_
```
hello
```
#### When
```shell
install -m 640 install/src install/dst && printf '%s' "$(cat install/dst)"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: sets the requested mode
#### Given
- Fixture file `install/src` is created.
#### Inputs
_Fixture `install/src`:_
```
hello
```
#### When
```shell
install -m 640 install/src install/dst2 && stat -c '%a' install/dst2
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: creates directories with -d
#### When
```shell
install -d install/a/b/c && [ -d install/a/b/c ] && echo ok
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: fails without a destination
#### When
```shell
install only-source
```
#### Then
- exit code is not `0`
- stderr contains `missing destination`
## mimixbox install_gnu
Source: `test/e2e/tools/mimixbox/shellutils/install_gnu.atago.yaml`
### Scenario: makes a simple backup before overwriting with --backup=simple
#### When
```shell
printf 'new\n' > src
printf 'old\n' > dst
install --backup=simple src dst
cat dst dst~

```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
new
old
```
### Scenario: honors --suffix when backing up
#### When
```shell
printf 'new\n' > src
printf 'old\n' > dst
install --backup=simple -S .bak src dst
cat dst.bak

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: makes numbered backups with --backup=numbered
#### When
```shell
printf 'v1\n' > src
printf 'orig\n' > dst
install --backup=numbered src dst
printf 'v2\n' > src
install --backup=numbered src dst
cat dst dst.~1~ dst.~2~

```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
v2
orig
v1
```
### Scenario: uses a simple backup with --backup=existing when none are numbered
#### When
```shell
printf 'new\n' > src
printf 'old\n' > dst
install --backup=existing src dst
cat dst~

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: attempts chown and fails as non-root with --owner/--group
#### When
```shell
printf 'data\n' > src
install -o 0 -g 0 src dst

```
#### Then
- exit code is not `0`
- stderr contains `ownership`
- file `dst` exists
#### Generated artifacts
- `dst`
### Scenario: rejects an invalid owner name
#### When
```shell
printf 'data\n' > src
install -o no-such-user-xyz src dst

```
#### Then
- exit code is not `0`
- stderr contains `invalid user`
### Scenario: rejects an invalid --backup control
#### When
```shell
printf 'data\n' > src
install --backup=bogus src dst

```
#### Then
- exit code is not `0`
- stderr contains `invalid argument`
## mimixbox kill
Source: `test/e2e/tools/mimixbox/shellutils/kill.atago.yaml`
### Scenario: lists signal names with -l
#### When
```shell
kill -l | grep -q KILL && echo ok
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox killall
Source: `test/e2e/tools/mimixbox/shellutils/killall.atago.yaml`
### Scenario: kills a process by name
#### When
```shell
sleep 30 & sleep 0.2; killall sleep; echo "killed:$?"
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox leadtime
Source: `test/e2e/tools/mimixbox/shellutils/leadtime.atago.yaml`
### Scenario: prints usage with --help and exits 0
#### When
```shell
leadtime --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: leadtime`, `LT_GITHUB_ACCESS_TOKEN`
### Scenario: fails when no subcommand is given
#### When
```shell
leadtime
```
#### Then
- exit code is not `0`
- stderr contains `stat`
### Scenario: fails on an unknown subcommand
#### When
```shell
leadtime bogus --owner=a --repo=b
```
#### Then
- exit code is not `0`
- stderr contains `unknown subcommand`
### Scenario: fails with a deterministic error when no token is set
#### When
```shell
LT_GITHUB_ACCESS_TOKEN= GITHUB_TOKEN= leadtime stat --owner=acme --repo=demo
```
#### Then
- exit code is not `0`
- stderr contains `no GitHub token`
### Scenario: fails when --owner/--repo are missing
#### When
```shell
LT_GITHUB_ACCESS_TOKEN=x leadtime stat
```
#### Then
- exit code is not `0`
- stderr contains `--owner and --repo are required`
### Scenario: rejects --json with --markdown
#### When
```shell
LT_GITHUB_ACCESS_TOKEN=x leadtime stat --owner=a --repo=b --json --markdown
```
#### Then
- exit code is not `0`
- stderr contains `mutually exclusive`
## mimixbox top-level list/suggestion CLI
Source: `test/e2e/tools/mimixbox/shellutils/list_topcli.atago.yaml`
### Scenario: --list --json emits a JSON array containing cat and ls on stdout
#### When
```shell
mimixbox --list --json
```
#### Then
- exit code is `0`
- stdout contains `"name": "cat"`, `"name": "ls"`, `"subsystem":`, `"stability":`, `[`, `]`
### Scenario: --list --filter=cat includes cat and excludes ls
#### When
```shell
mimixbox --list --filter=cat
```
#### Then
- exit code is `0`
- stdout contains `cat`
- stdout contains `cat`
### Scenario: --list --subsystem=textutils includes cat and excludes ls
#### When
```shell
mimixbox --list --subsystem=textutils
```
#### Then
- exit code is `0`
- stdout contains `cat`
- stdout does not contain ` ls -`
### Scenario: an unknown command suggests the nearest applet, error-first
#### When
```shell
mimixbox lss
```
#### Then
- exit code is not `0`
- stdout is empty
- stderr contains `'lss' is not a mimixbox command.`, `Did you mean:`, `ls`
## mimixbox log-collect
Source: `test/e2e/tools/mimixbox/shellutils/logcollect.atago.yaml`
### Scenario: copies log files into the output directory
#### Given
- Fixture file `src/a.log` is created.
#### Inputs
_Fixture `src/a.log`:_
```
log
```
#### When
```shell
log-collect -o ${workdir}/out ${workdir}/src >/dev/null && cat ${workdir}/out/a.log
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox logname
Source: `test/e2e/tools/mimixbox/shellutils/logname.atago.yaml`
### Scenario: prints the login name from LOGNAME
#### When
```shell
LOGNAME=mimixuser logname
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox mbsh
Source: `test/e2e/tools/mimixbox/shellutils/mbsh.atago.yaml`
### Scenario: runs an external command and shows a cwd-aware prompt
#### When
```shell
printf 'echo hello\nexit\n' | mbsh 2>/dev/null
```
#### Then
- exit code is `0`
- stdout contains `hello`, `mbsh:`
### Scenario: ignores comment lines
#### When
```shell
printf '# a comment\necho ok\nexit\n' | mbsh 2>/dev/null
```
#### Then
- exit code is `0`
- stdout contains `ok`
### Scenario: expands $? to the last exit status
#### When
```shell
printf 'false\necho status=$?\nexit\n' | mbsh 2>/dev/null
```
#### Then
- exit code is `0`
- stdout contains `status=1`
### Scenario: lets a stdin-consuming command read the remaining script input
#### When
```shell
printf 'cat\nhello\nexit\n' | mbsh 2>/dev/null
```
#### Then
- exit code is `0`
- stdout contains `hello`
### Scenario: does not reparse command-consumed stdin as later commands
#### When
```shell
err=$(printf 'cat\nhello\nexit\n' | mbsh 2>&1 >/dev/null)
case "${err}" in
    *"not a mimixbox command"*) echo reparsed ;;
    *) echo ok ;;
esac

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: keeps double-quoted spaces in one argument
#### When
```shell
printf 'echo "a b"\nexit\n' | mbsh 2>/dev/null
```
#### Then
- exit code is `0`
- stdout contains `a b`
### Scenario: expands $HOME
#### When
```shell
out=$(printf 'echo $HOME\nexit\n' | mbsh 2>/dev/null)
case "${out}" in
    *"${HOME}"*) echo expanded ;;
    *) echo literal ;;
esac

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: passes a NAME=value prefix to the command environment
#### When
```shell
printf 'FOO=bar env\nexit\n' | mbsh 2>/dev/null | grep '^FOO=bar$'
```
#### Then
- exit code is `0`
- stdout contains `FOO=bar`
### Scenario: runs commands in sequence and redirects output
#### When
```shell
d=$(mktemp -d)
printf 'echo one > %s/o; echo two >> %s/o\nexit\n' "$d" "$d" | mbsh >/dev/null 2>&1
cat "$d/o"
rm -rf "$d"

```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
one
two
```
### Scenario: pipes one command into another
#### When
```shell
d=$(mktemp -d)
printf 'printf foo | wc -c > %s/o\nexit\n' "$d" | mbsh >/dev/null 2>&1
tr -d ' \n' < "$d/o"
rm -rf "$d"

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: redirects input with <
#### When
```shell
d=$(mktemp -d)
printf 'a\nb\nc\n' > "$d/in"
printf 'wc -l < %s/in > %s/o\nexit\n' "$d" "$d" | mbsh >/dev/null 2>&1
tr -d ' \n' < "$d/o"
rm -rf "$d"

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: changes directory with cd
#### When
```shell
mkdir -p ${workdir}/mbsh && printf 'cd %s\npwd\nexit\n' "${workdir}/mbsh" | mbsh 2>/dev/null
```
#### Then
- exit code is `0`
- stdout contains `${workdir}/mbsh`
## mimixbox top-level CLI
Source: `test/e2e/tools/mimixbox/shellutils/mimixbox.atago.yaml`
### Scenario: prints usage to stdout with --help and exits success
#### When
```shell
mimixbox --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: mimixbox`, `Examples:`
### Scenario: lists the applets with --list
#### When
```shell
mimixbox --list
```
#### Then
- exit code is `0`
- stdout contains `cat`, `pidof`
### Scenario: rejects an unknown option on stderr without polluting stdout
#### When
```shell
mimixbox --definitely-not-an-option
```
#### Then
- exit code is not `0`
- stdout is empty
- stderr contains `is not a mimixbox command or option`
### Scenario: installs and removes applet symlinks in a temp directory
#### When
```shell
d=$(mktemp -d) || exit 1
trap "rm -rf '$d'" EXIT
mimixbox --full-install "$d" >/dev/null 2>&1 || exit 1
[ -L "$d/cat" ] || exit 1
[ -L "$d/pidof" ] || exit 1
mimixbox --remove "$d" >/dev/null 2>&1 || exit 1
[ -L "$d/cat" ] && exit 1
[ -L "$d/pidof" ] && exit 1
echo ok

```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox mknod
Source: `test/e2e/tools/mimixbox/shellutils/mknod.atago.yaml`
### Scenario: creates a FIFO with type p
#### When
```shell
mknod ${workdir}/pipe p && [ -p ${workdir}/pipe ] && echo fifo
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: rejects an invalid device type
#### When
```shell
mknod ${workdir}/x z
```
#### Then
- exit code is not `0`
- stderr contains `invalid device type`
## mimixbox nice
Source: `test/e2e/tools/mimixbox/shellutils/nice.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
nice --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: nice`
- stderr is empty
## mimixbox nohup
Source: `test/e2e/tools/mimixbox/shellutils/nohup.atago.yaml`
### Scenario: runs the command and passes output through
#### When
```shell
nohup echo hello
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox nproc
Source: `test/e2e/tools/mimixbox/shellutils/nproc.atago.yaml`
### Scenario: prints a positive number
#### When
```shell
nproc
```
#### Then
- exit code is `0`
- stdout matches `/^[0-9]+\n?$/`
## mimixbox od
Source: `test/e2e/tools/mimixbox/shellutils/od.atago.yaml`
### Scenario: dumps characters with C escapes
#### When
```shell
printf 'ABC\n' | od -c
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
0000000   A   B   C  \n
0000004
```
### Scenario: dumps hex bytes with hex addresses
#### When
```shell
printf 'AB' | od -A x -t x1
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
000000 41 42
000002
```
### Scenario: suppresses the address column
#### When
```shell
printf 'A' | od -A n -t o1
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox path
Source: `test/e2e/tools/mimixbox/shellutils/path.atago.yaml`
### Scenario: prints the base name with --basename
#### When
```shell
path -b /home/nao/test.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: prints the directory with --dirname
#### When
```shell
path -d /home/nao/test.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: prints the extension with --extension
#### When
```shell
path -e /home/nao/test.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: prints the cleaned path with --canonical
#### When
```shell
path -c /home/nao/../nao/./test.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: reports an error with no operand
#### When
```shell
path
```
#### Then
- exit code is not `0`
- stderr equals an exact value
## mimixbox pidof
Source: `test/e2e/tools/mimixbox/shellutils/pidof.atago.yaml`
### Scenario: finds the PID of a running process via MimixBox pidof
#### When
```shell
mimixbox pidof mimixbox >/dev/null 2>&1
sleep 5 &
SLEEP_PID=$!
RESULT=$(mimixbox pidof sleep)
kill ${SLEEP_PID} 2>/dev/null
echo "${RESULT}" | grep -q "${SLEEP_PID}" && echo found

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: resolves bare pidof to the MimixBox-installed symlink
#### When
```shell
pidof_path=$(command -v pidof) || exit 1
[ -L "${pidof_path}" ] || exit 1
target=$(readlink "${pidof_path}")
case "${target}" in
  *mimixbox) echo linked ;;
  *) exit 1 ;;
esac

```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox posixer
Source: `test/e2e/tools/mimixbox/shellutils/posixer.atago.yaml`
### Scenario: prints a table header
#### When
```shell
posixer | head -n 1
```
#### Then
- exit code is `0`
- stdout contains `NAME`, `INSTALLED`
## mimixbox printenv
Source: `test/e2e/tools/mimixbox/shellutils/printenv.atago.yaml`
### Scenario: prints an environment variable
#### When
```shell
MB_TEST_VAR=hello printenv MB_TEST_VAR
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox printf
Source: `test/e2e/tools/mimixbox/shellutils/printf.atago.yaml`
### Scenario: formats arguments
#### When
```shell
printf %s-%s\n foo bar
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox printf_meta
Source: `test/e2e/tools/mimixbox/shellutils/printf_meta.atago.yaml`
### Scenario: prints help for a leading --help
#### When
```shell
printf --help
```
#### Then
- exit code is `0`
- stdout contains `Examples:`
- stdout matches `/^Usage: printf/`
### Scenario: prints the version banner for a leading --version
#### When
```shell
printf --version
```
#### Then
- exit code is `0`
- stdout contains `printf (mimixbox)`
### Scenario: treats a later --help as an ordinary operand
#### When
```shell
printf 'foo --help\n'
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox pwd
Source: `test/e2e/tools/mimixbox/shellutils/pwd.atago.yaml`
### Scenario: prints the working directory
#### When
```shell
cd /tmp && pwd
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox realpath
Source: `test/e2e/tools/mimixbox/shellutils/realpath.atago.yaml`
### Scenario: resolves an existing file to its absolute path
#### Given
- Fixture file `file.txt` is created.
#### When
```shell
realpath file.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: prints the cleaned absolute path with -m on a missing path
#### When
```shell
realpath -m ${workdir}/does/not/exist
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: reports an error with no operand
#### When
```shell
realpath
```
#### Then
- exit code is not `0`
- stderr equals an exact value
## mimixbox realpath_gnu
Source: `test/e2e/tools/mimixbox/shellutils/realpath_gnu.atago.yaml`
### Scenario: prints a path relative to --relative-to
#### Given
- Fixture file `a/b/.keep` is created.
#### When
```shell
realpath --relative-to ${workdir}/a ${workdir}/a/b
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: resolves .. lexically with -L -m
#### When
```shell
realpath -L -m /tmp/../etc/./hosts
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox richhelp
Source: `test/e2e/tools/mimixbox/shellutils/richhelp.atago.yaml`
### Scenario: cp --help has a description, examples, and exit status
#### When
```shell
cp --help
```
#### Then
- exit code is `0`
- stdout contains `Examples:`, `Exit status:`
### Scenario: tail --help documents follow mode with examples
#### When
```shell
tail --help
```
#### Then
- exit code is `0`
- stdout contains `Examples:`, `Follow the file`
### Scenario: wget --help has examples and compatibility notes
#### When
```shell
wget --help
```
#### Then
- exit code is `0`
- stdout contains `Examples:`, `Notes:`
### Scenario: mbsh --help describes the shell and its limits
#### When
```shell
mbsh --help
```
#### Then
- exit code is `0`
- stdout contains `Examples:`, `Notes:`
### Scenario: vi --help lists the supported keys
#### When
```shell
vi --help
```
#### Then
- exit code is `0`
- stdout contains `Motions:`
### Scenario: find --help has examples and notes
#### When
```shell
find --help
```
#### Then
- exit code is `0`
- stdout contains `Examples:`, `Notes:`
## mimixbox sddf
Source: `test/e2e/tools/mimixbox/shellutils/sddf.atago.yaml`
### Scenario: prints usage with --help and exits 0
#### When
```shell
sddf --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: sddf`
### Scenario: fails with a message when given no operand
#### When
```shell
sddf
```
#### Then
- exit code is not `0`
- stderr contains `sddf`
## mimixbox seq
Source: `test/e2e/tools/mimixbox/shellutils/seq.atago.yaml`
### Scenario: counts from 1 to LAST
#### When
```shell
seq 3
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
1
2
3
```
### Scenario: counts from FIRST to LAST
#### When
```shell
seq 2 5
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
2
3
4
5
```
### Scenario: counts by INCREMENT
#### When
```shell
seq 1 2 9
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
1
3
5
7
9
```
### Scenario: joins the numbers with the separator
#### When
```shell
seq -s , 1 3
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: pads numbers with leading zeros
#### When
```shell
seq -w 8 10
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
08
09
10
```
### Scenario: reports an error for an invalid operand
#### When
```shell
seq abc
```
#### Then
- exit code is not `0`
- stderr equals an exact value
## mimixbox sleep
Source: `test/e2e/tools/mimixbox/shellutils/sleep.atago.yaml`
### Scenario: sleeps then returns
#### When
```shell
sleep 0.1 && echo slept
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox sort
Source: `test/e2e/tools/mimixbox/shellutils/sort.atago.yaml`
### Scenario: sorts lines alphabetically
#### When
```shell
printf 'banana\napple\ncherry\n' | sort
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
apple
banana
cherry
```
### Scenario: sorts by numeric value
#### When
```shell
printf '10\n2\n1\n' | sort -n
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
1
2
10
```
### Scenario: reverses the order
#### When
```shell
printf 'a\nb\nc\n' | sort -r
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
c
b
a
```
### Scenario: drops duplicate lines
#### When
```shell
printf 'a\na\nb\n' | sort -u
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
a
b
```
## mimixbox sort (GNU extensions)
Source: `test/e2e/tools/mimixbox/shellutils/sort_gnu.atago.yaml`
### Scenario: -V orders version numbers by value
#### When
```shell
printf '1.10\n1.2\n1.1\n' | sort -V
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
1.1
1.2
1.10
```
### Scenario: -g orders floating-point values including exponents
#### When
```shell
printf '1e3\n2.5\n100\n0.5\n' | sort -g
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
0.5
2.5
100
1e3
```
### Scenario: -h orders human-readable sizes by magnitude
#### When
```shell
printf '1G\n2K\n1M\n' | sort -h
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
2K
1M
1G
```
### Scenario: -s keeps input order for equal keys
#### When
```shell
printf '5 zebra\n5 apple\n5 mango\n' | sort -s -n
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
5 zebra
5 apple
5 mango
```
### Scenario: -z reads and writes NUL-delimited records
#### When
```shell
printf 'banana\000apple\000cherry\000' | sort -z | tr '\000' '|'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: -m merges already-sorted input
#### When
```shell
printf 'apple\nbanana\ncherry\n' | sort -m
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
apple
banana
cherry
```
### Scenario: --parallel is accepted without error
#### When
```shell
printf 'b\na\n' | sort --parallel=4
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
a
b
```
### Scenario: --temporary-directory is accepted without error
#### When
```shell
printf 'b\na\n' | sort --temporary-directory=/tmp
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
a
b
```
## mimixbox speaker
Source: `test/e2e/tools/mimixbox/shellutils/speaker.atago.yaml`
### Scenario: errors when no text is given
#### When
```shell
speaker 2>&1; echo "rc:$?"
```
#### Then
- exit code is `0`
- stdout contains `rc:1`
## mimixbox sync
Source: `test/e2e/tools/mimixbox/shellutils/sync.atago.yaml`
### Scenario: flushes filesystem buffers
#### When
```shell
sync && echo synced
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox tee
Source: `test/e2e/tools/mimixbox/shellutils/tee.atago.yaml`
### Scenario: echoes standard input to stdout
#### When
```shell
printf 'hello\n' | tee ${workdir}/out.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: also writes the input to the file
#### When
```shell
printf 'hello\n' | tee ${workdir}/out.txt > /dev/null; cat ${workdir}/out.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: appends to a file with -a keeping the existing content
#### When
```shell
printf 'one\n' | tee ${workdir}/log.txt > /dev/null
printf 'two\n' | tee -a ${workdir}/log.txt > /dev/null
cat ${workdir}/log.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
one
two
```
## mimixbox tee --output-error
Source: `test/e2e/tools/mimixbox/shellutils/tee_output_error.atago.yaml`
### Scenario: copies input and succeeds with an explicit MODE
#### When
```shell
printf 'hello\n' | tee --output-error=warn ${workdir}/good.txt > /dev/null; cat ${workdir}/good.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: warn mode still writes the good file but exits nonzero
#### When
```shell
printf 'payload\n' | tee --output-error=warn \
  ${workdir}/missing-dir/bad.txt ${workdir}/good.txt > /dev/null 2>&1
rc=$?
cat ${workdir}/good.txt 2>/dev/null
exit ${rc}

```
#### Then
- exit code is not `0`
- stdout equals an exact value
### Scenario: exit mode does not create the later good file and exits nonzero
#### When
```shell
printf 'payload\n' | tee --output-error=exit \
  ${workdir}/missing-dir/bad.txt ${workdir}/good.txt > /dev/null 2>&1
rc=$?
if [ -f ${workdir}/good.txt ]; then
  printf 'present\n'
else
  printf 'absent\n'
fi
exit ${rc}

```
#### Then
- exit code is not `0`
- stdout equals an exact value
## mimixbox test
Source: `test/e2e/tools/mimixbox/shellutils/test.atago.yaml`
### Scenario: string equality is true for equal strings
#### When
```shell
test abc = abc; echo "rc=$?"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: integer comparison is true when 2 > 1
#### When
```shell
test 2 -gt 1; echo "rc=$?"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: integer comparison is false when 1 > 2
#### When
```shell
test 1 -gt 2; echo "rc=$?"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: file existence is true for an existing file
#### Given
- Fixture file `file.txt` is created.
#### When
```shell
test -f ${workdir}/file.txt; echo "rc=$?"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: negation negates the expression
#### When
```shell
test ! -f /no_such_file; echo "rc=$?"
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox test (meta)
Source: `test/e2e/tools/mimixbox/shellutils/test_meta.atago.yaml`
### Scenario: prints help for a sole --help
#### When
```shell
env test --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: test`
### Scenario: prints the version banner for a sole --version
#### When
```shell
env test --version
```
#### Then
- exit code is `0`
- stdout contains `test (mimixbox)`
### Scenario: evaluates an expression when --help is not the sole argument
#### When
```shell
env test foo = --help
```
#### Then
- exit code is not `0`
- stdout does not contain `Usage: test`
## mimixbox time
Source: `test/e2e/tools/mimixbox/shellutils/time.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
time --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: time`
- stderr is empty
## mimixbox time / fsync
Source: `test/e2e/tools/mimixbox/shellutils/timefsync.atago.yaml`
### Scenario: time runs the command and passes its output through
#### When
```shell
env time echo timed 2>/dev/null
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: time reports the real elapsed time on stderr
#### When
```shell
env time echo x 2>&1 1>/dev/null | grep -c real
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: fsync succeeds on an existing file
#### Given
- Fixture file `f.txt` is created.
#### Inputs
_Fixture `f.txt`:_
```
data
```
#### When
```shell
fsync ${workdir}/f.txt; echo $?
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: fsync fails on a missing file
#### When
```shell
fsync /no/such/mimixbox/file 2>/dev/null; echo $?
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox timeout
Source: `test/e2e/tools/mimixbox/shellutils/timeout.atago.yaml`
### Scenario: runs the command to completion
#### When
```shell
timeout 5 echo done
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: returns exit code 124 on timeout
#### When
```shell
timeout 0.1 sleep 5; echo "exit:$?"
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox tree / nice
Source: `test/e2e/tools/mimixbox/shellutils/tree.atago.yaml`
### Scenario: tree counts directories and files in its summary
#### Given
- Fixture file `sub/leaf.txt` is created.
- Fixture file `root.txt` is created.
#### When
```shell
tree ${workdir} | grep directories
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: tree exits successfully on a readable directory
#### Given
- Fixture file `sub/leaf.txt` is created.
- Fixture file `root.txt` is created.
#### When
```shell
tree ${workdir} > /dev/null; echo $?
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: nice prints a numeric niceness
#### When
```shell
nice
```
#### Then
- exit code is `0`
- stdout matches `/[0-9]/`
## mimixbox true
Source: `test/e2e/tools/mimixbox/shellutils/true.atago.yaml`
### Scenario: prints nothing and exits 0
#### When
```shell
true
```
#### Then
- exit code is `0`
- stdout is empty
## mimixbox tsort
Source: `test/e2e/tools/mimixbox/shellutils/tsort.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
tsort --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: tsort`
- stderr is empty
### Scenario: documents its purpose in --help
#### When
```shell
tsort --help
```
#### Then
- exit code is `0`
- stdout contains `total ordering`
### Scenario: produces a topological order
#### Inputs
_stdin for `tsort`:_
```
a b
b c
```
#### When
```shell
tsort
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
a
b
c
```
## mimixbox tty
Source: `test/e2e/tools/mimixbox/shellutils/tty.atago.yaml`
### Scenario: reports not a tty when stdin is a pipe
#### When
```shell
echo "" | tty
```
#### Then
- exit code is not `0`
- stdout equals an exact value
## mimixbox uname
Source: `test/e2e/tools/mimixbox/shellutils/uname.atago.yaml`
### Scenario: prints the kernel name
#### When
```shell
uname -s
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox uniq
Source: `test/e2e/tools/mimixbox/shellutils/uniq.atago.yaml`
### Scenario: collapses repeated adjacent lines
#### When
```shell
printf 'a\na\nb\nc\nc\nc\n' | uniq
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
a
b
c
```
### Scenario: -c prefixes each line with its count
#### When
```shell
printf 'a\na\nb\nc\nc\nc\n' | uniq -c
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
      2 a
      1 b
      3 c
```
### Scenario: -d prints only repeated lines once
#### When
```shell
printf 'a\na\nb\nc\nc\n' | uniq -d
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
a
c
```
### Scenario: -u prints only lines that never repeat
#### When
```shell
printf 'a\na\nb\nc\nc\n' | uniq -u
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox users
Source: `test/e2e/tools/mimixbox/shellutils/users.atago.yaml`
### Scenario: runs and exits successfully
#### When
```shell
users >/dev/null 2>&1; echo $?
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: treats a missing utmp as nobody logged in
#### When
```shell
out=$(users /no/such/mimixbox/utmp); echo "[$out] rc=$?"
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox usleep
Source: `test/e2e/tools/mimixbox/shellutils/usleep.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
usleep --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: usleep`
- stderr is empty
### Scenario: rejects a non-numeric microsecond count
#### When
```shell
usleep notanumber
```
#### Then
- exit code is not `0`
- stderr contains `usleep:`
## mimixbox uuidgen
Source: `test/e2e/tools/mimixbox/shellutils/uuidgen.atago.yaml`
### Scenario: prints a well-formed UUID
#### When
```shell
uuidgen | grep -Eq '^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$' && echo ok
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox w
Source: `test/e2e/tools/mimixbox/shellutils/w.atago.yaml`
### Scenario: prints a summary header with the load averages
#### When
```shell
w | sed -n '1p'
```
#### Then
- exit code is `0`
- stdout contains `load average:`, `up `
### Scenario: prints the column header
#### When
```shell
w | sed -n '2p'
```
#### Then
- exit code is `0`
- stdout contains `USER`, `LOGIN@`
## mimixbox watch
Source: `test/e2e/tools/mimixbox/shellutils/watch.atago.yaml`
### Scenario: runs the command and shows its output
#### When
```shell
timeout 0.6 watch -t -n 0.2 echo tick 2>/dev/null
```
#### Then
- exit code is not `0`
- stdout contains `tick`
## mimixbox wget
Source: `test/e2e/tools/mimixbox/shellutils/wget.atago.yaml`
### Scenario: prints usage with --help and exits 0
#### When
```shell
wget --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: wget`
### Scenario: fails with a message when given no operand
#### When
```shell
wget
```
#### Then
- exit code is not `0`
- stderr contains `wget`
### Scenario: documents the added download options
#### When
```shell
wget --help
```
#### Then
- exit code is `0`
- stdout contains `--directory-prefix`, `--continue`, `--timeout`, `--tries`, `--user-agent`
## mimixbox which
Source: `test/e2e/tools/mimixbox/shellutils/which.atago.yaml`
### Scenario: prints the MimixBox path
#### When
```shell
[ "$(which mimixbox)" = "$(dirname "$(command -v mimixbox)")/mimixbox" ] && echo match
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: prints nothing for a binary that does not exist
#### When
```shell
which no_exist_binary
```
#### Then
- exit code is not `0`
- stdout is empty
### Scenario: prints paths of three binaries
#### When
```shell
d=$(dirname "$(command -v mimixbox)")
[ "$(which mimixbox cat tac)" = "$(printf '%s\n%s\n%s' "$d/mimixbox" "$d/cat" "$d/tac")" ] && echo match

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: prints paths of two binaries and fails when one of three is missing
#### When
```shell
which mimixbox not_exist_binary tac
```
#### Then
- exit code is not `0`
- stdout matches `//mimixbox\n.*/tac\n?$/`
### Scenario: prints nothing without an operand
#### When
```shell
which
```
#### Then
- exit code is not `0`
- stdout is empty
### Scenario: prints nothing when data comes from a pipe
#### When
```shell
echo "test" | which
```
#### Then
- exit code is not `0`
- stdout is empty
## mimixbox who
Source: `test/e2e/tools/mimixbox/shellutils/who.atago.yaml`
### Scenario: prints nothing and succeeds on an empty utmp
#### When
```shell
who /dev/null; echo "rc=$?"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: -q reports zero users on an empty utmp
#### When
```shell
who -q /dev/null
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: --help prints usage
#### When
```shell
who --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: who`
## mimixbox whoami
Source: `test/e2e/tools/mimixbox/shellutils/whoami.atago.yaml`
### Scenario: prints the current user name
#### When
```shell
[ "$(whoami)" = "$(id -un)" ] && echo match
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: reports an error with an extra operand
#### When
```shell
whoami extra
```
#### Then
- exit code is not `0`
- stderr equals an exact value
## mimixbox yes
Source: `test/e2e/tools/mimixbox/shellutils/yes.atago.yaml`
### Scenario: repeats y until the reader closes
#### When
```shell
yes | head -n 3
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
y
y
y
```
### Scenario: repeats the given string
#### When
```shell
yes mimix | head -n 2
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
mimix
mimix
```
## mimixbox base32
Source: `test/e2e/tools/mimixbox/textutils/base32.atago.yaml`
### Scenario: encodes standard input
#### When
```shell
printf 'hello\n' | base32
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: decodes standard input
#### When
```shell
printf 'NBSWY3DPBI======\n' | base32 -d
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox cat
Source: `test/e2e/tools/mimixbox/textutils/cat.atago.yaml`
### Scenario: show shell family name
#### Given
- Fixture file `cat.txt` is created.
#### Inputs
_Fixture `cat.txt`:_
```
sh
ash
csh
bash
```
#### When
```shell
cat cat.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
sh
ash
csh
bash
```
### Scenario: show shell family name with line numbers
#### Given
- Fixture file `cat.txt` is created.
#### Inputs
_Fixture `cat.txt`:_
```
sh
ash
csh
bash
```
#### When
```shell
cat -n cat.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
     1	sh
     2	ash
     3	csh
     4	bash
```
### Scenario: show the piped path unchanged
#### When
```shell
echo "${workdir}/cat.txt" | cat
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: cat only the file operand, ignoring pipe data
#### Given
- Fixture file `cat.txt` is created.
#### Inputs
_Fixture `cat.txt`:_
```
sh
ash
csh
bash
```
#### When
```shell
echo "${workdir}/cat2.txt" | cat cat.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
sh
ash
csh
bash
```
### Scenario: concatenate two files
#### Given
- Fixture file `cat.txt` is created.
- Fixture file `cat2.txt` is created.
#### Inputs
_Fixture `cat.txt`:_
```
sh
ash
csh
bash
```
_Fixture `cat2.txt`:_
```
fish
zsh
```
#### When
```shell
cat cat.txt cat2.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
sh
ash
csh
bash
fish
zsh
```
### Scenario: concatenate two files with line numbers
#### Given
- Fixture file `cat.txt` is created.
- Fixture file `cat2.txt` is created.
#### Inputs
_Fixture `cat.txt`:_
```
sh
ash
csh
bash
```
_Fixture `cat2.txt`:_
```
fish
zsh
```
#### When
```shell
cat -n cat.txt cat2.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
     1	sh
     2	ash
     3	csh
     4	bash
     5	fish
     6	zsh
```
### Scenario: concatenate a heredoc and a file via redirect
#### Given
- Fixture file `cat.txt` is created.
#### Inputs
_Fixture `cat.txt`:_
```
sh
ash
csh
bash
```
#### When
```shell
cat - << EOS cat.txt > cat2.txt
fish
zsh
EOS
cat cat2.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
fish
zsh
sh
ash
csh
bash
```
### Scenario: show error for a missing file
#### When
```shell
cat no_exist_file
```
#### Then
- exit code is not `0`
- stderr equals an exact value
## mimixbox cat_showall
Source: `test/e2e/tools/mimixbox/textutils/cat_showall.atago.yaml`
### Scenario: -A and --show-all are aliases
#### Given
- Fixture file `cat_nonprinting.bin` is created.
#### When
```shell
cat -A "${workdir}/cat_nonprinting.bin" > short.out
cat --show-all "${workdir}/cat_nonprinting.bin" > long.out
if cmp -s short.out long.out; then echo "identical"; else echo "different"; fi

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: -v and --show-nonprinting are aliases
#### Given
- Fixture file `cat_nonprinting.bin` is created.
#### When
```shell
cat -v "${workdir}/cat_nonprinting.bin" > short.out
cat --show-nonprinting "${workdir}/cat_nonprinting.bin" > long.out
if cmp -s short.out long.out; then echo "identical"; else echo "different"; fi

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: --show-all renders tabs as ^I, non-printing bytes, and $ line ends
#### Given
- Fixture file `cat_nonprinting.bin` is created.
#### When
```shell
cat --show-all ${workdir}/cat_nonprinting.bin
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
a^Ib^A$
$
^?M-^@$
```
### Scenario: --show-nonprinting leaves TAB alone and renders ^X, ^?, and M- notation
#### Given
- Fixture file `cat_nonprinting.bin` is created.
#### When
```shell
cat --show-nonprinting ${workdir}/cat_nonprinting.bin
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
a	b^A

^?M-^@
```
## mimixbox checksum
Source: `test/e2e/tools/mimixbox/textutils/checksum.atago.yaml`
### Scenario: sum prints the BSD checksum and block count
#### When
```shell
printf 'hello\n' | sum
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: crc32 prints the CRC-32 of stdin
#### When
```shell
printf 'hello\n' | crc32
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: sha384sum prints the SHA-384 digest
#### When
```shell
printf 'hello\n' | sha384sum
```
#### Then
- exit code is `0`
- stdout contains `1d0f284efe3edea4b9ca3bd514fa134b17eae361ccc7a1eefeff801b9bd6604e`
## mimixbox cksum
Source: `test/e2e/tools/mimixbox/textutils/cksum.atago.yaml`
### Scenario: prints the CRC checksum and byte count
#### When
```shell
printf 'hello\n' | cksum
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox comm
Source: `test/e2e/tools/mimixbox/textutils/comm.atago.yaml`
### Scenario: print lines common to both files
#### Given
- Fixture file `comm_a.txt` is created.
- Fixture file `comm_b.txt` is created.
#### Inputs
_Fixture `comm_a.txt`:_
```
apple
banana
```
_Fixture `comm_b.txt`:_
```
banana
cherry
```
#### When
```shell
comm -1 -2 ${workdir}/comm_a.txt ${workdir}/comm_b.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox comm_gnu
Source: `test/e2e/tools/mimixbox/textutils/comm_gnu.atago.yaml`
### Scenario: separate the columns with --output-delimiter
#### Given
- Fixture file `comm_gnu/a.txt` is created.
- Fixture file `comm_gnu/b.txt` is created.
#### Inputs
_Fixture `comm_gnu/a.txt`:_
```
apple
banana
cherry
```
_Fixture `comm_gnu/b.txt`:_
```
banana
cherry
date
```
#### When
```shell
comm --output-delimiter=, ${workdir}/comm_gnu/a.txt ${workdir}/comm_gnu/b.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout equals an exact value
- stdout equals an exact value
- stdout equals an exact value
### Scenario: read and write NUL-terminated records with -z
#### Given
- Fixture file `comm_gnu/za.txt` is created.
- Fixture file `comm_gnu/zb.txt` is created.
#### When
```shell
comm -z -1 -2 ${workdir}/comm_gnu/za.txt ${workdir}/comm_gnu/zb.txt | tr '\0' '#'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: report an unsorted input on stderr and fail with --check-order
#### Given
- Fixture file `comm_gnu/unsorted.txt` is created.
- Fixture file `comm_gnu/b.txt` is created.
#### Inputs
_Fixture `comm_gnu/unsorted.txt`:_
```
cherry
banana
```
_Fixture `comm_gnu/b.txt`:_
```
banana
cherry
date
```
#### When
```shell
comm --check-order ${workdir}/comm_gnu/unsorted.txt ${workdir}/comm_gnu/b.txt
echo "rc=$?"

```
#### Then
- exit code is `0`
- stdout equals an exact value
- stderr contains `file 1 is not in sorted order`
## mimixbox convert_mode
Source: `test/e2e/tools/mimixbox/textutils/convert_mode.atago.yaml`
### Scenario: dos2unix keeps the original mode
#### Given
- Fixture file `d2u.txt` is created.
#### Inputs
_Fixture `d2u.txt`:_
```
a
b
```
#### When
```shell
dos2unix "${workdir}/d2u.txt" >/dev/null 2>&1
stat -c '%a' "${workdir}/d2u.txt"

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: unix2dos keeps the original mode
#### Given
- Fixture file `u2d.txt` is created.
#### Inputs
_Fixture `u2d.txt`:_
```
a
b
```
#### When
```shell
unix2dos "${workdir}/u2d.txt" >/dev/null 2>&1
stat -c '%a' "${workdir}/u2d.txt"

```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox crc32
Source: `test/e2e/tools/mimixbox/textutils/crc32.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
crc32 --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: crc32`
- stderr is empty
### Scenario: prints the CRC-32 of stdin
#### Inputs
_stdin for `crc32`:_
```
hello
```
#### When
```shell
crc32
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox dos2unix
Source: `test/e2e/tools/mimixbox/textutils/dos2unix.atago.yaml`
### Scenario: convert a CRLF file to LF and reclassify it
#### Given
- Fixture file `dos2unix/1.txt` is created.
#### Inputs
_Fixture `dos2unix/1.txt`:_
```
abc
def
ghi
```
#### When
```shell
dos2unix "${workdir}/dos2unix/1.txt"
file "${workdir}/dos2unix/1.txt"

```
#### Then
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
dos2unix: converting file ${workdir}/dos2unix/1.txt to Unix format...
${workdir}/dos2unix/1.txt: ASCII text
```
### Scenario: convert a CRLF file and exit success
#### Given
- Fixture file `dos2unix/1.txt` is created.
#### Inputs
_Fixture `dos2unix/1.txt`:_
```
abc
def
ghi
```
#### When
```shell
dos2unix "${workdir}/dos2unix/1.txt"
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
dos2unix: converting file ${workdir}/dos2unix/1.txt to Unix format...
```
### Scenario: convert three CRLF files at once and reclassify each
#### Given
- Fixture file `dos2unix/1.txt` is created.
- Fixture file `dos2unix/2.txt` is created.
- Fixture file `dos2unix/3.txt` is created.
#### Inputs
_Fixture `dos2unix/1.txt`:_
```
abc
def
ghi
```
_Fixture `dos2unix/2.txt`:_
```
abc
def
ghi
```
_Fixture `dos2unix/3.txt`:_
```
abc
def
ghi
```
#### When
```shell
dos2unix "${workdir}/dos2unix/1.txt" "${workdir}/dos2unix/2.txt" "${workdir}/dos2unix/3.txt"
file "${workdir}/dos2unix/1.txt"
file "${workdir}/dos2unix/2.txt"
file "${workdir}/dos2unix/3.txt"

```
#### Then
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
dos2unix: converting file ${workdir}/dos2unix/1.txt to Unix format...
dos2unix: converting file ${workdir}/dos2unix/2.txt to Unix format...
dos2unix: converting file ${workdir}/dos2unix/3.txt to Unix format...
${workdir}/dos2unix/1.txt: ASCII text
${workdir}/dos2unix/2.txt: ASCII text
${workdir}/dos2unix/3.txt: ASCII text
```
### Scenario: convert three CRLF files at once and exit success
#### Given
- Fixture file `dos2unix/1.txt` is created.
- Fixture file `dos2unix/2.txt` is created.
- Fixture file `dos2unix/3.txt` is created.
#### Inputs
_Fixture `dos2unix/1.txt`:_
```
abc
def
ghi
```
_Fixture `dos2unix/2.txt`:_
```
abc
def
ghi
```
_Fixture `dos2unix/3.txt`:_
```
abc
def
ghi
```
#### When
```shell
dos2unix "${workdir}/dos2unix/1.txt" "${workdir}/dos2unix/2.txt" "${workdir}/dos2unix/3.txt"
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
dos2unix: converting file ${workdir}/dos2unix/1.txt to Unix format...
dos2unix: converting file ${workdir}/dos2unix/2.txt to Unix format...
dos2unix: converting file ${workdir}/dos2unix/3.txt to Unix format...
```
### Scenario: refuse a directory with a not-regular-file error
#### Given
- Fixture file `dos2unix/1.txt` is created.
#### Inputs
_Fixture `dos2unix/1.txt`:_
```
abc
def
ghi
```
#### When
```shell
dos2unix ${workdir}/dos2unix
```
#### Then
- exit code is not `0`
- stderr equals an exact value
### Scenario: convert the two files but fail on the directory operand
#### Given
- Fixture file `dos2unix/1.txt` is created.
- Fixture file `dos2unix/3.txt` is created.
#### Inputs
_Fixture `dos2unix/1.txt`:_
```
abc
def
ghi
```
_Fixture `dos2unix/3.txt`:_
```
abc
def
ghi
```
#### When
```shell
dos2unix ${workdir}/dos2unix/1.txt ${workdir}/dos2unix ${workdir}/dos2unix/3.txt
```
#### Then
- exit code is not `0`
- stdout equals an exact value
- stderr equals an exact value
#### Expected output
_expected stdout:_
```
dos2unix: converting file ${workdir}/dos2unix/1.txt to Unix format...
dos2unix: converting file ${workdir}/dos2unix/3.txt to Unix format...
```
## mimixbox expand
Source: `test/e2e/tools/mimixbox/textutils/expand.atago.yaml`
### Scenario: converts tabs to spaces (default tab stop 8)
#### When
```shell
printf 'a\tb\n' | expand
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: converts tabs to the given width
#### When
```shell
printf 'a\tb\n' | expand -t 4
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: converts tabs in the file
#### Given
- Fixture file `expand.txt` is created.
#### Inputs
_Fixture `expand.txt`:_
```
a	b
```
#### When
```shell
expand expand.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: reports an error for a non-existent file
#### When
```shell
expand /no_exist_file
```
#### Then
- exit code is not `0`
- stderr equals an exact value
## mimixbox fmt
Source: `test/e2e/tools/mimixbox/textutils/fmt.atago.yaml`
### Scenario: reflows text to the given width
#### When
```shell
printf 'aa bb cc dd\n' | fmt -w 5
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
aa bb
cc dd
```
## mimixbox fold
Source: `test/e2e/tools/mimixbox/textutils/fold.atago.yaml`
### Scenario: wraps lines to the given width
#### When
```shell
printf 'abcdefgh\n' | fold -w 3
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
abc
def
gh
```
## mimixbox head
Source: `test/e2e/tools/mimixbox/textutils/head.atago.yaml`
### Scenario: print the first 10 lines
#### Given
- Fixture file `head.txt` is created.
#### Inputs
_Fixture `head.txt`:_
```
1
2
3
4
5
6
7
8
9
10
11
12
```
#### When
```shell
head head.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
1
2
3
4
5
6
7
8
9
10
```
### Scenario: print the first N lines
#### Given
- Fixture file `head.txt` is created.
#### Inputs
_Fixture `head.txt`:_
```
1
2
3
4
5
6
7
8
9
10
11
12
```
#### When
```shell
head -n 3 head.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
1
2
3
```
### Scenario: print the first N bytes
#### When
```shell
printf 'hello world' | head -c 5
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: print the first N lines of stdin
#### When
```shell
printf 'a\nb\nc\nd\n' | head -n 2
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
a
b
```
### Scenario: report an error for a non-existent file
#### When
```shell
head /no_exist_file
```
#### Then
- exit code is not `0`
- stderr equals an exact value
## mimixbox head --zero-terminated
Source: `test/e2e/tools/mimixbox/textutils/head_zero.atago.yaml`
### Scenario: prints the first NUL-delimited record, preserving the embedded newline
#### When
```shell
printf 'a\nb\0c\nd\0' | head -z -n 1 | tr '\0' '|'
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
a
b|
```
### Scenario: prints two NUL-delimited records with embedded newlines preserved
#### When
```shell
printf 'a\nb\0c\nd\0' | head --zero-terminated -n 2 | tr '\0' '|'
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
a
b|c
d|
```
## mimixbox textutils --help helpers
Source: `test/e2e/tools/mimixbox/textutils/help_helpers_textutils.atago.yaml`
### Scenario: crc32 --help is structured
#### When
```shell
env -- crc32 --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: sha384sum --help is structured
#### When
```shell
env -- sha384sum --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: sha3sum --help is structured
#### When
```shell
env -- sha3sum --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: sum --help is structured
#### When
```shell
env -- sum --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: uudecode --help is structured
#### When
```shell
env -- uudecode --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: uuencode --help is structured
#### When
```shell
env -- uuencode --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
## mimixbox man
Source: `test/e2e/tools/mimixbox/textutils/man.atago.yaml`
### Scenario: show a plain manual page
#### Given
- Fixture file `man/man1/foo.1` is created.
#### Inputs
_Fixture `man/man1/foo.1`:_
```
FOO(1)
the foo page
```
#### When
```shell
man -M "${workdir}/man" foo
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
FOO(1)
the foo page
```
### Scenario: decompress a gzipped manual page
#### When
```shell
mkdir -p "${workdir}/man/man1"; printf 'BAR(1)\nthe bar page\n' | gzip > "${workdir}/man/man1/bar.1.gz"
man -M "${workdir}/man" bar
```
#### Then
- after `man -M "${workdir}/man" bar`:
  - exit code is `0`
  - stdout equals an exact value
#### Expected output
_expected stdout:_
```
BAR(1)
the bar page
```
### Scenario: report a missing page with exit 16
#### Given
- Fixture file `man/man1/foo.1` is created.
#### Inputs
_Fixture `man/man1/foo.1`:_
```
FOO(1)
the foo page
```
#### When
```shell
man -M "${workdir}/man" missing 2>/dev/null; echo "exit=$?"
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox md5sum
Source: `test/e2e/tools/mimixbox/textutils/md5sum.atago.yaml`
### Scenario: get md5sum of one file
#### Given
- Fixture file `md5sum/1.txt` is created.
#### Inputs
_Fixture `md5sum/1.txt`:_
```
Dungeon of Regalias
```
#### When
```shell
md5sum ${workdir}/md5sum/1.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: cannot get md5sum of one directory
#### Given
- Fixture file `md5sum/1.txt` is created.
#### Inputs
_Fixture `md5sum/1.txt`:_
```
Dungeon of Regalias
```
#### When
```shell
md5sum ${workdir}/md5sum
```
#### Then
- exit code is not `0`
- stderr equals an exact value
### Scenario: cannot get md5sum of not exist file
#### When
```shell
md5sum /not_exist_file
```
#### Then
- exit code is not `0`
- stderr equals an exact value
### Scenario: get md5sum of three files
#### Given
- Fixture file `md5sum/1.txt` is created.
- Fixture file `md5sum/2.txt` is created.
- Fixture file `md5sum/3.txt` is created.
#### Inputs
_Fixture `md5sum/1.txt`:_
```
Dungeon of Regalias
```
_Fixture `md5sum/2.txt`:_
```
DEMONION
```
_Fixture `md5sum/3.txt`:_
```
Dungeon Crusadearz
```
#### When
```shell
md5sum ${workdir}/md5sum/1.txt ${workdir}/md5sum/2.txt ${workdir}/md5sum/3.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
d0d8ffef81b3c7160ac655d5939548c5  ${workdir}/md5sum/1.txt
07e280ad4bd77b9321f0ce3386775019  ${workdir}/md5sum/2.txt
15e924f84517598e828f49dc85765bc5  ${workdir}/md5sum/3.txt
```
### Scenario: check md5sum with --check option
#### Given
- Fixture file `md5sum/1.txt` is created.
- Fixture file `md5sum/2.txt` is created.
- Fixture file `md5sum/3.txt` is created.
- Fixture file `md5sum/checksum.txt` is created.
#### Inputs
_Fixture `md5sum/1.txt`:_
```
Dungeon of Regalias
```
_Fixture `md5sum/2.txt`:_
```
DEMONION
```
_Fixture `md5sum/3.txt`:_
```
Dungeon Crusadearz
```
_Fixture `md5sum/checksum.txt`:_
```
d0d8ffef81b3c7160ac655d5939548c5  ${workdir}/md5sum/1.txt
07e280ad4bd77b9321f0ce3386775019  ${workdir}/md5sum/2.txt
15e924f84517598e828f49dc85765bc5  ${workdir}/md5sum/3.txt
```
#### When
```shell
md5sum -c ${workdir}/md5sum/checksum.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
${workdir}/md5sum/1.txt: OK
${workdir}/md5sum/2.txt: OK
${workdir}/md5sum/3.txt: OK
```
### Scenario: get md5sum for pipe data
#### When
```shell
echo "test" | md5sum
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: get md5sum for pipe data and file at same time
#### Given
- Fixture file `md5sum/1.txt` is created.
#### Inputs
_Fixture `md5sum/1.txt`:_
```
Dungeon of Regalias
```
#### When
```shell
echo "test" | md5sum ${workdir}/md5sum/1.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox nl
Source: `test/e2e/tools/mimixbox/textutils/nl.atago.yaml`
### Scenario: number each line of a file
#### Given
- Fixture file `nl.txt` is created.
#### Inputs
_Fixture `nl.txt`:_
```
sh
ash
csh
bash
```
#### When
```shell
nl nl.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
     1	sh
     2	ash
     3	csh
     4	bash
```
### Scenario: number the single line read from pipe data
#### Given
- Fixture file `nl.txt` is created.
#### Inputs
_Fixture `nl.txt`:_
```
sh
ash
csh
bash
```
#### When
```shell
echo "${workdir}/nl.txt" | nl
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: number only the file operand, ignoring pipe data
#### Given
- Fixture file `nl.txt` is created.
#### Inputs
_Fixture `nl.txt`:_
```
sh
ash
csh
bash
```
#### When
```shell
echo "${workdir}/nl.txt" | nl nl.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
     1	sh
     2	ash
     3	csh
     4	bash
```
### Scenario: number lines across two files
#### Given
- Fixture file `nl.txt` is created.
- Fixture file `nl2.txt` is created.
#### Inputs
_Fixture `nl.txt`:_
```
sh
ash
csh
bash
```
_Fixture `nl2.txt`:_
```
fish
zsh
```
#### When
```shell
nl nl.txt nl2.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
     1	sh
     2	ash
     3	csh
     4	bash
     5	fish
     6	zsh
```
### Scenario: number heredoc then file via a redirect
#### Given
- Fixture file `nl.txt` is created.
#### Inputs
_Fixture `nl.txt`:_
```
sh
ash
csh
bash
```
#### When
```shell
nl - << EOS nl.txt > nl2.txt
fish
zsh
EOS
cat nl2.txt

```
#### Then
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
     1	fish
     2	zsh
     3	sh
     4	ash
     5	csh
     6	bash
```
### Scenario: number heredoc then file via a redirect with success status
#### Given
- Fixture file `nl.txt` is created.
#### Inputs
_Fixture `nl.txt`:_
```
sh
ash
csh
bash
```
#### When
```shell
nl - << EOS nl.txt > nl2.txt
fish
zsh
EOS
cat nl2.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
     1	fish
     2	zsh
     3	sh
     4	ash
     5	csh
     6	bash
```
### Scenario: report an error for a non-existent file
#### When
```shell
nl no_exist_file
```
#### Then
- exit code is not `0`
- stderr equals an exact value
## mimixbox nl sections
Source: `test/e2e/tools/mimixbox/textutils/nl_sections.atago.yaml`
### Scenario: number every line in every section with -h a -b a -f a
#### Given
- Fixture file `nl_sec.txt` is created.
#### Inputs
_Fixture `nl_sec.txt`:_
```
H1
\:\:\:
HDR
\:\:
B1
\:
F1
```
#### When
```shell
nl -h a -b a -f a nl_sec.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
     1	H1

     1	HDR

     1	B1

     1	F1
```
### Scenario: number header (a) and body (t) but not footer (n)
#### Given
- Fixture file `nl_sec.txt` is created.
#### Inputs
_Fixture `nl_sec.txt`:_
```
H1
\:\:\:
HDR
\:\:
B1
\:
F1
```
#### When
```shell
nl -h a -b t -f n nl_sec.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
     1	H1

     1	HDR

     1	B1

       F1
```
### Scenario: number every second blank line with -l 2
#### Given
- Fixture file `nl_blank.txt` is created.
#### Inputs
_Fixture `nl_blank.txt`:_
```
a




b
```
#### When
```shell
nl -b a -l 2 nl_blank.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
     1	a
       
     2	
       
     3	
     4	b
```
## mimixbox paste
Source: `test/e2e/tools/mimixbox/textutils/paste.atago.yaml`
### Scenario: joins lines with a delimiter
#### When
```shell
printf 'a\nb\nc\n' | paste -s -d ,
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox rev
Source: `test/e2e/tools/mimixbox/textutils/rev.atago.yaml`
### Scenario: reverses the characters of a line
#### When
```shell
printf 'abc\n' | rev
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox sha1sum
Source: `test/e2e/tools/mimixbox/textutils/sha1sum.atago.yaml`
### Scenario: get sha1sum of one file
#### Given
- Fixture file `sha1sum/1.txt` is created.
#### Inputs
_Fixture `sha1sum/1.txt`:_
```
Dungeon of Regalias
```
#### When
```shell
sha1sum ${workdir}/sha1sum/1.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
9dc2936d38932f9ffc6738cb677e4a8722116070  ${workdir}/sha1sum/1.txt
```
### Scenario: cannot get sha1sum of one directory
#### Given
- Fixture file `sha1sum/1.txt` is created.
#### Inputs
_Fixture `sha1sum/1.txt`:_
```
Dungeon of Regalias
```
#### When
```shell
sha1sum ${workdir}/sha1sum
```
#### Then
- exit code is not `0`
- stderr equals an exact value
### Scenario: cannot get sha1sum of not exist file
#### When
```shell
sha1sum /not_exist_file
```
#### Then
- exit code is not `0`
- stderr equals an exact value
### Scenario: get sha1sum of three files
#### Given
- Fixture file `sha1sum/1.txt` is created.
- Fixture file `sha1sum/2.txt` is created.
- Fixture file `sha1sum/3.txt` is created.
#### Inputs
_Fixture `sha1sum/1.txt`:_
```
Dungeon of Regalias
```
_Fixture `sha1sum/2.txt`:_
```
DEMONION
```
_Fixture `sha1sum/3.txt`:_
```
Dungeon Crusadearz
```
#### When
```shell
sha1sum ${workdir}/sha1sum/1.txt ${workdir}/sha1sum/2.txt ${workdir}/sha1sum/3.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
9dc2936d38932f9ffc6738cb677e4a8722116070  ${workdir}/sha1sum/1.txt
317e30648976d62fae4662fe4435e6568648e8a7  ${workdir}/sha1sum/2.txt
d4e9619d949de0c0182a09757346ad22e80114b3  ${workdir}/sha1sum/3.txt
```
### Scenario: check sha1sum with --check option
#### Given
- Fixture file `sha1sum/1.txt` is created.
- Fixture file `sha1sum/2.txt` is created.
- Fixture file `sha1sum/3.txt` is created.
- Fixture file `sha1sum/checksum.txt` is created.
#### Inputs
_Fixture `sha1sum/1.txt`:_
```
Dungeon of Regalias
```
_Fixture `sha1sum/2.txt`:_
```
DEMONION
```
_Fixture `sha1sum/3.txt`:_
```
Dungeon Crusadearz
```
_Fixture `sha1sum/checksum.txt`:_
```
9dc2936d38932f9ffc6738cb677e4a8722116070  ${workdir}/sha1sum/1.txt
317e30648976d62fae4662fe4435e6568648e8a7  ${workdir}/sha1sum/2.txt
d4e9619d949de0c0182a09757346ad22e80114b3  ${workdir}/sha1sum/3.txt
```
#### When
```shell
sha1sum -c ${workdir}/sha1sum/checksum.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
${workdir}/sha1sum/1.txt: OK
${workdir}/sha1sum/2.txt: OK
${workdir}/sha1sum/3.txt: OK
```
### Scenario: get sha1sum for pipe data
#### When
```shell
echo "test" | sha1sum
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: get sha1sum for pipe data and file at same time
#### Given
- Fixture file `sha1sum/1.txt` is created.
#### Inputs
_Fixture `sha1sum/1.txt`:_
```
Dungeon of Regalias
```
#### When
```shell
echo "test" | sha1sum ${workdir}/sha1sum/1.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
9dc2936d38932f9ffc6738cb677e4a8722116070  ${workdir}/sha1sum/1.txt
```
## mimixbox sha256sum
Source: `test/e2e/tools/mimixbox/textutils/sha256sum.atago.yaml`
### Scenario: get sha256sum of one file
#### Given
- Fixture file `sha256sum/1.txt` is created.
#### Inputs
_Fixture `sha256sum/1.txt`:_
```
Dungeon of Regalias
```
#### When
```shell
sha256sum ${workdir}/sha256sum/1.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
5f2864b5833190b07b0b95228682ff5ec43a13a2a3f31514c57d5c92aa3fb2e7  ${workdir}/sha256sum/1.txt
```
### Scenario: cannot get sha256sum of one directory
#### Given
- Fixture file `sha256sum/1.txt` is created.
#### Inputs
_Fixture `sha256sum/1.txt`:_
```
Dungeon of Regalias
```
#### When
```shell
sha256sum ${workdir}/sha256sum
```
#### Then
- exit code is not `0`
- stderr equals an exact value
### Scenario: cannot get sha256sum of not exist file
#### When
```shell
sha256sum /not_exist_file
```
#### Then
- exit code is not `0`
- stderr equals an exact value
### Scenario: get sha256sum of three files
#### Given
- Fixture file `sha256sum/1.txt` is created.
- Fixture file `sha256sum/2.txt` is created.
- Fixture file `sha256sum/3.txt` is created.
#### Inputs
_Fixture `sha256sum/1.txt`:_
```
Dungeon of Regalias
```
_Fixture `sha256sum/2.txt`:_
```
DEMONION
```
_Fixture `sha256sum/3.txt`:_
```
Dungeon Crusadearz
```
#### When
```shell
sha256sum ${workdir}/sha256sum/1.txt ${workdir}/sha256sum/2.txt ${workdir}/sha256sum/3.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
5f2864b5833190b07b0b95228682ff5ec43a13a2a3f31514c57d5c92aa3fb2e7  ${workdir}/sha256sum/1.txt
833d8136112b60552a0f83165a2ebffeac4b0c0249480d651ea58b9073ec925b  ${workdir}/sha256sum/2.txt
8e774f75a5a23c83e6f7d5e92863a2615e0335e06aec18d9c3ec1c5315d1a777  ${workdir}/sha256sum/3.txt
```
### Scenario: check sha256sum with --check option
#### Given
- Fixture file `sha256sum/1.txt` is created.
- Fixture file `sha256sum/2.txt` is created.
- Fixture file `sha256sum/3.txt` is created.
- Fixture file `sha256sum/checksum.txt` is created.
#### Inputs
_Fixture `sha256sum/1.txt`:_
```
Dungeon of Regalias
```
_Fixture `sha256sum/2.txt`:_
```
DEMONION
```
_Fixture `sha256sum/3.txt`:_
```
Dungeon Crusadearz
```
_Fixture `sha256sum/checksum.txt`:_
```
5f2864b5833190b07b0b95228682ff5ec43a13a2a3f31514c57d5c92aa3fb2e7  ${workdir}/sha256sum/1.txt
833d8136112b60552a0f83165a2ebffeac4b0c0249480d651ea58b9073ec925b  ${workdir}/sha256sum/2.txt
8e774f75a5a23c83e6f7d5e92863a2615e0335e06aec18d9c3ec1c5315d1a777  ${workdir}/sha256sum/3.txt
```
#### When
```shell
sha256sum -c ${workdir}/sha256sum/checksum.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
${workdir}/sha256sum/1.txt: OK
${workdir}/sha256sum/2.txt: OK
${workdir}/sha256sum/3.txt: OK
```
### Scenario: get sha256sum for pipe data
#### When
```shell
echo "test" | sha256sum
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2  -
```
### Scenario: get sha256sum for pipe data and file at same time
#### Given
- Fixture file `sha256sum/1.txt` is created.
#### Inputs
_Fixture `sha256sum/1.txt`:_
```
Dungeon of Regalias
```
#### When
```shell
echo "test" | sha256sum ${workdir}/sha256sum/1.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
5f2864b5833190b07b0b95228682ff5ec43a13a2a3f31514c57d5c92aa3fb2e7  ${workdir}/sha256sum/1.txt
```
## mimixbox sha384sum
Source: `test/e2e/tools/mimixbox/textutils/sha384sum.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
sha384sum --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: sha384sum`
- stderr is empty
### Scenario: prints the SHA-384 digest of stdin
#### Inputs
_stdin for `sha384sum`:_
```
hello
```
#### When
```shell
sha384sum
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
1d0f284efe3edea4b9ca3bd514fa134b17eae361ccc7a1eefeff801b9bd6604e01f21f6bf249ef030599f0c218f2ba8c  -
```
## mimixbox sha3sum
Source: `test/e2e/tools/mimixbox/textutils/sha3sum.atago.yaml`
### Scenario: defaults to SHA3-256
#### When
```shell
printf 'hello\n' | sha3sum | cut -d' ' -f1
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
b314e28493eae9dab57ac4f0c6d887bddbbeb810e900d818395ace558e96516d
```
### Scenario: selects SHA3-512 with -a
#### When
```shell
printf 'hello\n' | sha3sum -a 512 | cut -c1-16
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox sha512sum
Source: `test/e2e/tools/mimixbox/textutils/sha512sum.atago.yaml`
### Scenario: get sha512sum of one file
#### Given
- Fixture file `sha512sum/1.txt` is created.
#### Inputs
_Fixture `sha512sum/1.txt`:_
```
Dungeon of Regalias
```
#### When
```shell
sha512sum ${workdir}/sha512sum/1.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
05eec7dcf412f63d5a291d019f6b3d62d4f8f5236592815ed171f7d6d0a7969f65a589a092740bd04a2f181d7d5a27ff36808e04a69bd84a854aad0a01da3612  ${workdir}/sha512sum/1.txt
```
### Scenario: cannot get sha512sum of one directory
#### Given
- Fixture file `sha512sum/1.txt` is created.
#### Inputs
_Fixture `sha512sum/1.txt`:_
```
Dungeon of Regalias
```
#### When
```shell
sha512sum ${workdir}/sha512sum
```
#### Then
- exit code is not `0`
- stderr equals an exact value
### Scenario: cannot get sha512sum of not exist file
#### When
```shell
sha512sum /not_exist_file
```
#### Then
- exit code is not `0`
- stderr equals an exact value
### Scenario: get sha512sum of three files
#### Given
- Fixture file `sha512sum/1.txt` is created.
- Fixture file `sha512sum/2.txt` is created.
- Fixture file `sha512sum/3.txt` is created.
#### Inputs
_Fixture `sha512sum/1.txt`:_
```
Dungeon of Regalias
```
_Fixture `sha512sum/2.txt`:_
```
DEMONION
```
_Fixture `sha512sum/3.txt`:_
```
Dungeon Crusadearz
```
#### When
```shell
sha512sum ${workdir}/sha512sum/1.txt ${workdir}/sha512sum/2.txt ${workdir}/sha512sum/3.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
05eec7dcf412f63d5a291d019f6b3d62d4f8f5236592815ed171f7d6d0a7969f65a589a092740bd04a2f181d7d5a27ff36808e04a69bd84a854aad0a01da3612  ${workdir}/sha512sum/1.txt
cb2389a103184f607973b1acd073dc15310c8172b03f340a52bdc3843621cf9fbc6263c7dbbd786ceb0244f5147a83aa32ce09a485f544093b7fc5c7533e564f  ${workdir}/sha512sum/2.txt
3dafa5f1ec7f09cbe551dc0d4bdb153dedb81104b7e930b7c20733965f7ebb86ee2abea64b6bfa1c54045032865044a3feca5dcc89c28def410b2954094a1890  ${workdir}/sha512sum/3.txt
```
### Scenario: check sha512sum with --check option
#### Given
- Fixture file `sha512sum/1.txt` is created.
- Fixture file `sha512sum/2.txt` is created.
- Fixture file `sha512sum/3.txt` is created.
- Fixture file `sha512sum/checksum.txt` is created.
#### Inputs
_Fixture `sha512sum/1.txt`:_
```
Dungeon of Regalias
```
_Fixture `sha512sum/2.txt`:_
```
DEMONION
```
_Fixture `sha512sum/3.txt`:_
```
Dungeon Crusadearz
```
_Fixture `sha512sum/checksum.txt`:_
```
05eec7dcf412f63d5a291d019f6b3d62d4f8f5236592815ed171f7d6d0a7969f65a589a092740bd04a2f181d7d5a27ff36808e04a69bd84a854aad0a01da3612  ${workdir}/sha512sum/1.txt
cb2389a103184f607973b1acd073dc15310c8172b03f340a52bdc3843621cf9fbc6263c7dbbd786ceb0244f5147a83aa32ce09a485f544093b7fc5c7533e564f  ${workdir}/sha512sum/2.txt
3dafa5f1ec7f09cbe551dc0d4bdb153dedb81104b7e930b7c20733965f7ebb86ee2abea64b6bfa1c54045032865044a3feca5dcc89c28def410b2954094a1890  ${workdir}/sha512sum/3.txt
```
#### When
```shell
sha512sum -c ${workdir}/sha512sum/checksum.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
${workdir}/sha512sum/1.txt: OK
${workdir}/sha512sum/2.txt: OK
${workdir}/sha512sum/3.txt: OK
```
### Scenario: get sha512sum for pipe data
#### When
```shell
echo "test" | sha512sum
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
0e3e75234abc68f4378a86b3f4b32a198ba301845b0cd6e50106e874345700cc6663a86c1ea125dc5e92be17c98f9a0f85ca9d5f595db2012f7cc3571945c123  -
```
### Scenario: get sha512sum for pipe data and file at same time
#### Given
- Fixture file `sha512sum/1.txt` is created.
#### Inputs
_Fixture `sha512sum/1.txt`:_
```
Dungeon of Regalias
```
#### When
```shell
echo "test" | sha512sum ${workdir}/sha512sum/1.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
05eec7dcf412f63d5a291d019f6b3d62d4f8f5236592815ed171f7d6d0a7969f65a589a092740bd04a2f181d7d5a27ff36808e04a69bd84a854aad0a01da3612  ${workdir}/sha512sum/1.txt
```
## mimixbox shuf
Source: `test/e2e/tools/mimixbox/textutils/shuf.atago.yaml`
### Scenario: shuffles a single-element range
#### When
```shell
shuf -i 1-1
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox split
Source: `test/e2e/tools/mimixbox/textutils/split.atago.yaml`
### Scenario: split input into files of N lines
#### When
```shell
printf '1\n2\n3\n' | split -l 2 - "${workdir}/part-"; cat "${workdir}/part-aa"
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
1
2
```
## mimixbox split (GNU flags)
Source: `test/e2e/tools/mimixbox/textutils/split_gnu.atago.yaml`
### Scenario: use numeric suffixes with -d
#### When
```shell
printf '1\n2\n3\n4\n5\n' | split -l 2 -d - "${workdir}/num-"; ls "${workdir}"/num-* | sed "s#${workdir}/##" | sort | tr '\n' ' ' | sed 's/ $//'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: write expected content to the first numeric piece
#### When
```shell
printf '1\n2\n3\n4\n5\n' | split -l 2 -d - "${workdir}/num-"; cat "${workdir}/num-00"
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
1
2
```
### Scenario: append an additional suffix to each name
#### When
```shell
printf '1\n2\n3\n' | split -l 2 --additional-suffix=.txt - "${workdir}/add-"; ls "${workdir}"/add-* | sed "s#${workdir}/##" | sort | tr '\n' ' ' | sed 's/ $//'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: honor a custom suffix length with -a
#### When
```shell
printf '1\n2\n' | split -l 1 -a 3 - "${workdir}/len-"; ls "${workdir}"/len-* | sed "s#${workdir}/##" | sort | tr '\n' ' ' | sed 's/ $//'
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox sqluv
Source: `test/e2e/tools/mimixbox/textutils/sqluv.atago.yaml`
### Scenario: print usage with --help and exit 0
#### When
```shell
sqluv --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: sqluv`, `headless`
### Scenario: print the version and exit 0
#### When
```shell
sqluv --version
```
#### Then
- exit code is `0`
- stdout contains `sqluv (mimixbox)`
### Scenario: fail with a message when given no operand
#### When
```shell
sqluv
```
#### Then
- exit code is not `0`
- stderr contains `sqluv`
### Scenario: query a CSV fixture in headless mode
#### When
```shell
dir=$(mktemp -d "${TMPDIR:-/tmp}/sqluv-it.XXXXXX")
printf 'id,name\n1,alice\n2,bob\n3,carol\n' > "$dir/data.csv"
sqluv "$dir/data.csv" \
    --history-file "$dir/history.log" \
    --output csv \
    --execute 'select name from data order by id limit 2'
rm -rf "$dir"

```
#### Then
- exit code is `0`
- stdout contains `alice`, `bob`
### Scenario: query a SQLite-style table as JSON
#### When
```shell
dir=$(mktemp -d "${TMPDIR:-/tmp}/sqluv-it.XXXXXX")
printf 'title\ngo-in-action\nthe-go-programming-language\n' > "$dir/books.csv"
sqluv "$dir/books.csv" \
    --history-file "$dir/history.log" \
    --output json \
    --execute 'select title from books order by title'
rm -rf "$dir"

```
#### Then
- exit code is `0`
- stdout contains `go-in-action`
### Scenario: fail deterministically on an unsupported S3 source
#### When
```shell
sqluv --execute 'select 1' 's3://bucket/data.csv'
```
#### Then
- exit code is not `0`
- stderr contains `S3 sources are not migrated`
## mimixbox sqluv (compressed input)
Source: `test/e2e/tools/mimixbox/textutils/sqluv_compressed.atago.yaml`
### Scenario: query a gzip-compressed CSV fixture
#### When
```shell
dir=$(mktemp -d "${TMPDIR:-/tmp}/sqluv-it.XXXXXX")
printf 'id,name\n1,alice\n2,bob\n' | gzip -c > "$dir/data.csv.gz"
sqluv "$dir/data.csv.gz" \
    --history-file "$dir/history.log" \
    --output csv \
    --execute 'select count(*) as n from data'
rm -rf "$dir"

```
#### Then
- exit code is `0`
- stdout contains `2`
## mimixbox sqluv (history file)
Source: `test/e2e/tools/mimixbox/textutils/sqluv_history.atago.yaml`
### Scenario: write query history to the path given by --history-file
#### When
```shell
dir=$(mktemp -d "${TMPDIR:-/tmp}/sqluv-it.XXXXXX")
hist="$dir/sqluv-history.log"
printf 'id\n1\n2\n' > "$dir/nums.csv"
sqluv "$dir/nums.csv" --history-file "$hist" \
    --execute 'select count(*) from nums' >/dev/null
cat "$hist"
rm -rf "$dir"

```
#### Then
- exit code is `0`
- stdout contains `select count(*) from nums`
## mimixbox sqluv (TUI smoke)
Source: `test/e2e/tools/mimixbox/textutils/sqluv_tui_smoke.atago.yaml`
### Scenario: render the minimal viewer and exit cleanly on quit
#### When
```shell
dir=$(mktemp -d "${TMPDIR:-/tmp}/sqluv-it.XXXXXX")
printf 'id,name\n1,alice\n' > "$dir/data.csv"
printf 'q\n' | sqluv "$dir/data.csv" --history-file "$dir/history.log"
rm -rf "$dir"

```
#### Then
- exit code is `0`
- stdout contains `minimal viewer`, `bye`
## mimixbox strings
Source: `test/e2e/tools/mimixbox/textutils/strings.atago.yaml`
### Scenario: prints printable sequences
#### When
```shell
printf 'hi\000hello\000world' | strings
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
hello
world
```
## mimixbox sum
Source: `test/e2e/tools/mimixbox/textutils/sum.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
sum --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: sum`
- stderr is empty
### Scenario: documents its purpose in --help
#### When
```shell
sum --help
```
#### Then
- exit code is `0`
- stdout contains `BSD algorithm`
### Scenario: prints a BSD checksum and block count for stdin
#### Inputs
_stdin for `sum`:_
```
hello
```
#### When
```shell
sum
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox tac
Source: `test/e2e/tools/mimixbox/textutils/tac.atago.yaml`
### Scenario: print the lines in reverse order
#### Given
- Fixture file `tac.txt` is created.
#### Inputs
_Fixture `tac.txt`:_
```
first
second
third
```
#### When
```shell
tac tac.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
third
second
first
```
### Scenario: reverse standard input
#### When
```shell
printf 'a\nb\nc\n' | tac
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
c
b
a
```
### Scenario: report an error for a non-existent file
#### When
```shell
tac /no_exist_file
```
#### Then
- exit code is not `0`
- stderr equals an exact value
## mimixbox tail
Source: `test/e2e/tools/mimixbox/textutils/tail.atago.yaml`
### Scenario: print the last 10 lines
#### Given
- Fixture file `tail.txt` is created.
#### Inputs
_Fixture `tail.txt`:_
```
1
2
3
4
5
6
7
8
9
10
11
12
```
#### When
```shell
tail tail.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
3
4
5
6
7
8
9
10
11
12
```
### Scenario: print the last N lines
#### Given
- Fixture file `tail.txt` is created.
#### Inputs
_Fixture `tail.txt`:_
```
1
2
3
4
5
6
7
8
9
10
11
12
```
#### When
```shell
tail -n 3 tail.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
10
11
12
```
### Scenario: print the last N bytes
#### When
```shell
printf 'hello world' | tail -c 5
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: print the last N lines of stdin
#### When
```shell
printf 'a\nb\nc\nd\n' | tail -n 2
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
c
d
```
### Scenario: report an error for a non-existent file
#### When
```shell
tail /no_exist_file
```
#### Then
- exit code is not `0`
- stderr equals an exact value
### Scenario: print data appended while following
#### Given
- Fixture file `follow.txt` is created.
#### Inputs
_Fixture `follow.txt`:_
```
start
```
#### When
```shell
( sleep 0.2; printf 'appended\n' >> follow.txt ) &
timeout 0.5 tail -f -s 0.05 follow.txt

```
#### Then
- exit code is not `0`
- stdout contains `start`, `appended`
## mimixbox tail --pid
Source: `test/e2e/tools/mimixbox/textutils/tail_pid.atago.yaml`
### Scenario: stop following once the watched process exits
#### Given
- Fixture file `follow_pid.txt` is created.
#### Inputs
_Fixture `follow_pid.txt`:_
```
start
```
#### When
```shell
sleep 1 &
sleeper_pid=$!
( sleep 0.2; printf 'appended\n' >> follow_pid.txt ) &
timeout 5 tail -f -s 0.1 --pid="${sleeper_pid}" follow_pid.txt

```
#### Then
- exit code is `0`
- stdout contains `start`, `appended`
## mimixbox tail --zero-terminated
Source: `test/e2e/tools/mimixbox/textutils/tail_zero.atago.yaml`
### Scenario: prints the last NUL-delimited record, preserving the embedded newline
#### When
```shell
printf 'a\nb\0c\nd\0' | tail -z -n 1 | tr '\0' '|'
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
c
d|
```
### Scenario: prints two NUL-delimited records with embedded newlines preserved
#### When
```shell
printf 'a\nb\0c\nd\0' | tail --zero-terminated -n 2 | tr '\0' '|'
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
a
b|c
d|
```
## mimixbox tr
Source: `test/e2e/tools/mimixbox/textutils/tr.atago.yaml`
### Scenario: translates lowercase to uppercase
#### When
```shell
printf 'abc\n' | tr a-z A-Z
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox tr --truncate-set1
Source: `test/e2e/tools/mimixbox/textutils/tr_truncate.atago.yaml`
### Scenario: truncates SET1 to SET2 length, leaving extra chars unchanged
#### When
```shell
printf 'abc\n' | tr --truncate-set1 abc xy
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: accepts the -t short form
#### When
```shell
printf 'abc\n' | tr -t abc xy
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox unexpand
Source: `test/e2e/tools/mimixbox/textutils/unexpand.atago.yaml`
### Scenario: convert leading spaces to a tab
#### When
```shell
printf '        a\n' | unexpand
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: convert internal space runs to tabs with --all
#### When
```shell
printf 'a        b\n' | unexpand -a
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: report an error for a non-existent file
#### When
```shell
unexpand /no_exist_file
```
#### Then
- exit code is not `0`
- stderr equals an exact value
## mimixbox unix2dos
Source: `test/e2e/tools/mimixbox/textutils/unix2dos.atago.yaml`
### Scenario: convert an LF file to CRLF and reclassify it
#### Given
- Fixture file `unix2dos/1.txt` is created.
#### Inputs
_Fixture `unix2dos/1.txt`:_
```
abc
def
ghi
```
#### When
```shell
unix2dos "${workdir}/unix2dos/1.txt"
file "${workdir}/unix2dos/1.txt"

```
#### Then
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
unix2dos: converting file ${workdir}/unix2dos/1.txt to DOS format...
${workdir}/unix2dos/1.txt: ASCII text, with CRLF line terminators
```
### Scenario: convert an LF file and exit success
#### Given
- Fixture file `unix2dos/1.txt` is created.
#### Inputs
_Fixture `unix2dos/1.txt`:_
```
abc
def
ghi
```
#### When
```shell
unix2dos "${workdir}/unix2dos/1.txt"
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
unix2dos: converting file ${workdir}/unix2dos/1.txt to DOS format...
```
### Scenario: convert three LF files at once and reclassify each
#### Given
- Fixture file `unix2dos/1.txt` is created.
- Fixture file `unix2dos/2.txt` is created.
- Fixture file `unix2dos/3.txt` is created.
#### Inputs
_Fixture `unix2dos/1.txt`:_
```
abc
def
ghi
```
_Fixture `unix2dos/2.txt`:_
```
abc
def
ghi
```
_Fixture `unix2dos/3.txt`:_
```
abc
def
ghi
```
#### When
```shell
unix2dos "${workdir}/unix2dos/1.txt" "${workdir}/unix2dos/2.txt" "${workdir}/unix2dos/3.txt"
file "${workdir}/unix2dos/1.txt"
file "${workdir}/unix2dos/2.txt"
file "${workdir}/unix2dos/3.txt"

```
#### Then
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
unix2dos: converting file ${workdir}/unix2dos/1.txt to DOS format...
unix2dos: converting file ${workdir}/unix2dos/2.txt to DOS format...
unix2dos: converting file ${workdir}/unix2dos/3.txt to DOS format...
${workdir}/unix2dos/1.txt: ASCII text, with CRLF line terminators
${workdir}/unix2dos/2.txt: ASCII text, with CRLF line terminators
${workdir}/unix2dos/3.txt: ASCII text, with CRLF line terminators
```
### Scenario: convert three LF files at once and exit success
#### Given
- Fixture file `unix2dos/1.txt` is created.
- Fixture file `unix2dos/2.txt` is created.
- Fixture file `unix2dos/3.txt` is created.
#### Inputs
_Fixture `unix2dos/1.txt`:_
```
abc
def
ghi
```
_Fixture `unix2dos/2.txt`:_
```
abc
def
ghi
```
_Fixture `unix2dos/3.txt`:_
```
abc
def
ghi
```
#### When
```shell
unix2dos "${workdir}/unix2dos/1.txt" "${workdir}/unix2dos/2.txt" "${workdir}/unix2dos/3.txt"
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
unix2dos: converting file ${workdir}/unix2dos/1.txt to DOS format...
unix2dos: converting file ${workdir}/unix2dos/2.txt to DOS format...
unix2dos: converting file ${workdir}/unix2dos/3.txt to DOS format...
```
### Scenario: refuse a directory with a not-regular-file error
#### Given
- Fixture file `unix2dos/1.txt` is created.
#### Inputs
_Fixture `unix2dos/1.txt`:_
```
abc
def
ghi
```
#### When
```shell
unix2dos ${workdir}/unix2dos
```
#### Then
- exit code is not `0`
- stderr equals an exact value
### Scenario: convert the two files but fail on the directory operand
#### Given
- Fixture file `unix2dos/1.txt` is created.
- Fixture file `unix2dos/3.txt` is created.
#### Inputs
_Fixture `unix2dos/1.txt`:_
```
abc
def
ghi
```
_Fixture `unix2dos/3.txt`:_
```
abc
def
ghi
```
#### When
```shell
unix2dos ${workdir}/unix2dos/1.txt ${workdir}/unix2dos ${workdir}/unix2dos/3.txt
```
#### Then
- exit code is not `0`
- stdout equals an exact value
- stderr equals an exact value
#### Expected output
_expected stdout:_
```
unix2dos: converting file ${workdir}/unix2dos/1.txt to DOS format...
unix2dos: converting file ${workdir}/unix2dos/3.txt to DOS format...
```
## mimixbox uuencode/uudecode/usleep
Source: `test/e2e/tools/mimixbox/textutils/uucode.atago.yaml`
### Scenario: uuencode then uudecode round-trips (traditional)
#### When
```shell
d=$(mktemp -d); printf 'round trip data\n' > "$d/in"
uuencode "$d/in" out | uudecode -o -
rm -rf "$d"

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: uuencode -m then uudecode round-trips (base64)
#### When
```shell
d=$(mktemp -d); printf 'base64 round trip\n' > "$d/in"
uuencode -m "$d/in" out | uudecode -o -
rm -rf "$d"

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: usleep waits and exits 0
#### When
```shell
usleep 1000 && echo slept
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox uudecode
Source: `test/e2e/tools/mimixbox/textutils/uudecode.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
uudecode --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: uudecode`
- stderr is empty
## mimixbox uuencode
Source: `test/e2e/tools/mimixbox/textutils/uuencode.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
uuencode --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: uuencode`
- stderr is empty
### Scenario: uuencodes stdin with a begin header
#### Inputs
_stdin for `uuencode`:_
```
hello
```
#### When
```shell
uuencode hi.txt
```
#### Then
- exit code is `0`
- stdout contains `begin 644 hi.txt`, `end`
## mimixbox wc
Source: `test/e2e/tools/mimixbox/textutils/wc.atago.yaml`
### Scenario: count lines/words/bytes of one file
#### Given
- Fixture file `game.txt` is created.
#### Inputs
_Fixture `game.txt`:_
```
NieR Replicant ver.1.22474487139...
NieR:Automata
The Legend of Zelda: Majora's Mask
KICHIKUOU RANCE
DARK SOULS
SHADOW HEARTS
```
#### When
```shell
wc game.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: count only lines with --lines
#### Given
- Fixture file `game.txt` is created.
#### Inputs
_Fixture `game.txt`:_
```
NieR Replicant ver.1.22474487139...
NieR:Automata
The Legend of Zelda: Majora's Mask
KICHIKUOU RANCE
DARK SOULS
SHADOW HEARTS
```
#### When
```shell
wc -l game.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: count only bytes with --bytes
#### Given
- Fixture file `game.txt` is created.
#### Inputs
_Fixture `game.txt`:_
```
NieR Replicant ver.1.22474487139...
NieR:Automata
The Legend of Zelda: Majora's Mask
KICHIKUOU RANCE
DARK SOULS
SHADOW HEARTS
```
#### When
```shell
wc -c game.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: report the longest line with --max-line-length
#### Given
- Fixture file `game.txt` is created.
#### Inputs
_Fixture `game.txt`:_
```
NieR Replicant ver.1.22474487139...
NieR:Automata
The Legend of Zelda: Majora's Mask
KICHIKUOU RANCE
DARK SOULS
SHADOW HEARTS
```
#### When
```shell
wc -L game.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: count an empty file as all zeros
#### Given
- Fixture file `empty.txt` is created.
#### When
```shell
wc empty.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: count three files and print a total
#### Given
- Fixture file `empty.txt` is created.
- Fixture file `game.txt` is created.
- Fixture file `metal.txt` is created.
#### Inputs
_Fixture `game.txt`:_
```
NieR Replicant ver.1.22474487139...
NieR:Automata
The Legend of Zelda: Majora's Mask
KICHIKUOU RANCE
DARK SOULS
SHADOW HEARTS
```
_Fixture `metal.txt`:_
```
MEGADETH
GALNERYUS
SYSTEM OF A DOWN
```
#### When
```shell
wc empty.txt game.txt metal.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
  0   0   0 empty.txt
  6  16 126 game.txt
  3   6  36 metal.txt
  9  22 162 total
```
### Scenario: count piped data
#### When
```shell
echo "${workdir}/game.txt" | wc
```
#### Then
- exit code is `0`
- stdout matches `/1 +1 +[0-9]+/`
### Scenario: count only the file operand, ignoring pipe data
#### Given
- Fixture file `game.txt` is created.
#### Inputs
_Fixture `game.txt`:_
```
NieR Replicant ver.1.22474487139...
NieR:Automata
The Legend of Zelda: Majora's Mask
KICHIKUOU RANCE
DARK SOULS
SHADOW HEARTS
```
#### When
```shell
echo "${workdir}/game.txt" | wc game.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: report a directory as not a regular file
#### When
```shell
wc "${workdir}"
```
#### Then
- exit code is not `0`
- stdout equals an exact value
- stderr equals an exact value
### Scenario: count the file but zero the directory when given both
#### Given
- Fixture file `game.txt` is created.
#### Inputs
_Fixture `game.txt`:_
```
NieR Replicant ver.1.22474487139...
NieR:Automata
The Legend of Zelda: Majora's Mask
KICHIKUOU RANCE
DARK SOULS
SHADOW HEARTS
```
#### When
```shell
wc "${workdir}" "${workdir}/game.txt"
```
#### Then
- exit code is not `0`
- stdout equals an exact value
- stderr equals an exact value
#### Expected output
_expected stdout:_
```
      0       0       0 ${workdir}
      6      16     126 ${workdir}/game.txt
      6      16     126 total
```
### Scenario: count a single line piped in with --lines
#### When
```shell
echo "no_exist_file" | wc -l
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox wc (GNU flags)
Source: `test/e2e/tools/mimixbox/textutils/wc_gnu.atago.yaml`
### Scenario: print only the combined total with --total=only
#### Given
- Fixture file `wc_a.txt` is created.
- Fixture file `wc_b.txt` is created.
#### Inputs
_Fixture `wc_a.txt`:_
```
a
b
c
```
_Fixture `wc_b.txt`:_
```
x
```
#### When
```shell
wc --total=only wc_a.txt wc_b.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: suppress the total line with --total=never
#### Given
- Fixture file `wc_a.txt` is created.
- Fixture file `wc_b.txt` is created.
#### Inputs
_Fixture `wc_a.txt`:_
```
a
b
c
```
_Fixture `wc_b.txt`:_
```
x
```
#### When
```shell
wc --total=never "${workdir}/wc_a.txt" "${workdir}/wc_b.txt"
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
3 3 6 ${workdir}/wc_a.txt
1 1 2 ${workdir}/wc_b.txt
```
### Scenario: print a total even for one file with --total=always
#### Given
- Fixture file `wc_a.txt` is created.
#### Inputs
_Fixture `wc_a.txt`:_
```
a
b
c
```
#### When
```shell
wc --total=always "${workdir}/wc_a.txt"
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
3 3 6 ${workdir}/wc_a.txt
3 3 6 total
```
### Scenario: read a NUL-separated name list with --files0-from
#### Given
- Fixture file `wc_a.txt` is created.
- Fixture file `wc_b.txt` is created.
#### Inputs
_Fixture `wc_a.txt`:_
```
a
b
c
```
_Fixture `wc_b.txt`:_
```
x
```
#### When
```shell
printf '%s\0%s\0' "${workdir}/wc_a.txt" "${workdir}/wc_b.txt" > wc_list.nul
wc --files0-from=wc_list.nul
```
#### Then
- after `wc --files0-from=wc_list.nul`:
  - exit code is `0`
  - stdout equals an exact value
#### Expected output
_expected stdout:_
```
3 3 6 ${workdir}/wc_a.txt
1 1 2 ${workdir}/wc_b.txt
4 4 8 total
```
### Scenario: read the name list from standard input with --files0-from=-
#### Given
- Fixture file `wc_a.txt` is created.
- Fixture file `wc_b.txt` is created.
#### Inputs
_Fixture `wc_a.txt`:_
```
a
b
c
```
_Fixture `wc_b.txt`:_
```
x
```
#### When
```shell
printf '%s\0%s\0' "${workdir}/wc_a.txt" "${workdir}/wc_b.txt" | wc --files0-from=-
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```
3 3 6 ${workdir}/wc_a.txt
1 1 2 ${workdir}/wc_b.txt
4 4 8 total
```
### Scenario: combine --files0-from with --total=only
#### Given
- Fixture file `wc_a.txt` is created.
- Fixture file `wc_b.txt` is created.
#### Inputs
_Fixture `wc_a.txt`:_
```
a
b
c
```
_Fixture `wc_b.txt`:_
```
x
```
#### When
```shell
printf '%s\0%s\0' "${workdir}/wc_a.txt" "${workdir}/wc_b.txt" > wc_list.nul
wc --files0-from=wc_list.nul --total=only
```
#### Then
- after `wc --files0-from=wc_list.nul --total=only`:
  - exit code is `0`
  - stdout equals an exact value
## mimixbox xxd
Source: `test/e2e/tools/mimixbox/textutils/xxd.atago.yaml`
### Scenario: prints a hex dump
#### When
```shell
printf 'hello\n' | xxd
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: reverses a hex dump
#### When
```shell
printf 'hello\n' | xxd | xxd -r
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox blkdiscard
Source: `test/e2e/tools/mimixbox/util-linux/blkdiscard.atago.yaml`
### Scenario: requires a device
#### When
```shell
blkdiscard
```
#### Then
- exit code is not `0`
### Scenario: describes itself with --help
#### When
```shell
blkdiscard --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: blkdiscard`, `Discard`
## mimixbox blkid
Source: `test/e2e/tools/mimixbox/util-linux/blkid.atago.yaml`
### Scenario: identifies an ext filesystem
#### When
```shell
dd if=/dev/zero of=ext.img bs=1024 count=2 2>/dev/null; printf '\123\357' | dd of=ext.img bs=1 seek=1080 conv=notrunc 2>/dev/null; blkid ext.img
```
#### Then
- exit code is `0`
- stdout contains `TYPE="ext2"`
### Scenario: identifies an xfs filesystem
#### When
```shell
printf 'XFSB' > xfs.img; blkid xfs.img
```
#### Then
- exit code is `0`
- stdout contains `TYPE="xfs"`
### Scenario: exits 2 when nothing is identified
#### When
```shell
printf 'nothing here' > blank.img; blkid blank.img
```
#### Then
- exit code is `2`
## mimixbox blockdev
Source: `test/e2e/tools/mimixbox/util-linux/blockdev.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
blockdev --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: blockdev`, `DEVICE`
### Scenario: fails when no query flag is given
#### When
```shell
blockdev /dev/null
```
#### Then
- exit code is not `0`
## mimixbox chattr
Source: `test/e2e/tools/mimixbox/util-linux/chattr.atago.yaml`
### Scenario: rejects a malformed mode
#### When
```shell
chattr xi /tmp/f
```
#### Then
- exit code is not `0`
### Scenario: rejects an unknown attribute
#### When
```shell
chattr +Z /tmp/f
```
#### Then
- exit code is not `0`
## mimixbox chrt
Source: `test/e2e/tools/mimixbox/util-linux/chrt.atago.yaml`
### Scenario: prints a process scheduling policy
#### When
```shell
chrt -p $$ | grep -c 'scheduling policy'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: runs a command under a scheduling policy
#### When
```shell
chrt -o 0 -- echo scheduled
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox dmesg
Source: `test/e2e/tools/mimixbox/util-linux/dmesg.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
dmesg --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: dmesg`, `kernel ring buffer`
## mimixbox eject
Source: `test/e2e/tools/mimixbox/util-linux/eject.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
eject --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: eject`, `media`
### Scenario: fails on a missing device
#### When
```shell
eject /dev/no_such_cdrom
```
#### Then
- exit code is not `0`
## mimixbox fallocate
Source: `test/e2e/tools/mimixbox/util-linux/fallocate.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
fallocate --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: fallocate`
- stderr is empty
## mimixbox fatattr
Source: `test/e2e/tools/mimixbox/util-linux/fatattr.atago.yaml`
### Scenario: requires a file
#### When
```shell
fatattr
```
#### Then
- exit code is not `0`
### Scenario: rejects an unknown attribute
#### When
```shell
fatattr +Z /tmp/x
```
#### Then
- exit code is not `0`
## mimixbox fbset
Source: `test/e2e/tools/mimixbox/util-linux/fbset.atago.yaml`
### Scenario: fails on a missing framebuffer
#### When
```shell
fbset -fb /dev/no_such_fb
```
#### Then
- exit code is not `0`
### Scenario: describes itself with --help
#### When
```shell
fbset --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: fbset`, `framebuffer`
## mimixbox fdflush
Source: `test/e2e/tools/mimixbox/util-linux/fdflush.atago.yaml`
### Scenario: requires a device
#### When
```shell
fdflush
```
#### Then
- exit code is not `0`
### Scenario: describes itself with --help
#### When
```shell
fdflush --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: fdflush`, `floppy`
## mimixbox fdformat
Source: `test/e2e/tools/mimixbox/util-linux/fdformat.atago.yaml`
### Scenario: requires a device
#### When
```shell
fdformat
```
#### Then
- exit code is not `0`
### Scenario: describes itself with --help
#### When
```shell
fdformat --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: fdformat`, `floppy`
## mimixbox fdisk
Source: `test/e2e/tools/mimixbox/util-linux/fdisk.atago.yaml`
### Scenario: lists an MBR Linux partition
#### When
```shell
dd if=/dev/zero of=disk.img bs=512 count=1 2>/dev/null; printf '\203' | dd of=disk.img bs=1 seek=450 count=1 conv=notrunc 2>/dev/null; printf '\000\010\000\000' | dd of=disk.img bs=1 seek=454 count=4 conv=notrunc 2>/dev/null; printf '\144\000\000\000' | dd of=disk.img bs=1 seek=458 count=4 conv=notrunc 2>/dev/null; printf '\125\252' | dd of=disk.img bs=1 seek=510 count=2 conv=notrunc 2>/dev/null; fdisk -l disk.img | grep -c 'Linux'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: rejects an image without an MBR signature
#### When
```shell
dd if=/dev/zero of=n.img bs=512 count=1 2>/dev/null; fdisk -l n.img
```
#### Then
- exit code is not `0`
## mimixbox findfs
Source: `test/e2e/tools/mimixbox/util-linux/findfs.atago.yaml`
### Scenario: fails for an unknown label
#### When
```shell
findfs LABEL=no_such_label_xyz
```
#### Then
- exit code is not `0`
### Scenario: rejects a malformed tag
#### When
```shell
findfs notatag
```
#### Then
- exit code is not `0`
## mimixbox flock
Source: `test/e2e/tools/mimixbox/util-linux/flock.atago.yaml`
### Scenario: runs a command while holding the lock
#### When
```shell
flock lock echo locked-run
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: fails -n when the lock is already held
#### When
```shell
flock lock sleep 1 & sleep 0.2; flock -n lock echo nope
```
#### Then
- exit code is not `0`
## mimixbox freeramdisk
Source: `test/e2e/tools/mimixbox/util-linux/freeramdisk.atago.yaml`
### Scenario: requires a device
#### When
```shell
freeramdisk
```
#### Then
- exit code is not `0`
### Scenario: describes itself with --help
#### When
```shell
freeramdisk --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: freeramdisk`, `ramdisk`
## mimixbox fsck
Source: `test/e2e/tools/mimixbox/util-linux/fsck.atago.yaml`
### Scenario: detects a Minix filesystem
#### When
```shell
dd if=/dev/zero of=m.img bs=1024 count=2048 2>/dev/null; mkfs.minix m.img >/dev/null; fsck m.img
```
#### Then
- exit code is `0`
- stdout contains `minix`
### Scenario: fails on an unrecognized image
#### When
```shell
dd if=/dev/zero of=u.img bs=1024 count=8 2>/dev/null; fsck u.img
```
#### Then
- exit code is not `0`
## mimixbox fsck.minix
Source: `test/e2e/tools/mimixbox/util-linux/fsck.minix.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
fsck.minix --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: fsck.minix`
- stderr is empty
## mimixbox fsck.minix
Source: `test/e2e/tools/mimixbox/util-linux/fsck_minix.atago.yaml`
### Scenario: validates a freshly made Minix filesystem
#### When
```shell
dd if=/dev/zero of=f.img bs=1024 count=2048 2>/dev/null; mkfs.minix f.img >/dev/null; fsck.minix f.img | sed -n '1p'
```
#### Then
- exit code is `0`
- stdout contains `Minix v1`
### Scenario: rejects a non-Minix image
#### When
```shell
dd if=/dev/zero of=b.img bs=1024 count=4 2>/dev/null; fsck.minix b.img
```
#### Then
- exit code is not `0`
## mimixbox fsfreeze
Source: `test/e2e/tools/mimixbox/util-linux/fsfreeze.atago.yaml`
### Scenario: requires a freeze or unfreeze mode
#### When
```shell
fsfreeze /mnt
```
#### Then
- exit code is not `0`
### Scenario: rejects both modes at once
#### When
```shell
fsfreeze -f -u /mnt
```
#### Then
- exit code is not `0`
## mimixbox fstrim
Source: `test/e2e/tools/mimixbox/util-linux/fstrim.atago.yaml`
### Scenario: requires a mount point
#### When
```shell
fstrim
```
#### Then
- exit code is not `0`
### Scenario: describes itself with --help
#### When
```shell
fstrim --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: fstrim`, `Discard`
## mimixbox getopt
Source: `test/e2e/tools/mimixbox/util-linux/getopt.atago.yaml`
### Scenario: normalizes short and long options with quoted args
#### When
```shell
getopt -o ab: --long alpha,beta: -- -a -b val --alpha pos
```
#### Then
- exit code is `0`
- stdout contains `-b 'val'`, `--alpha`, `-- 'pos'`
### Scenario: produces output a script can eval
#### When
```shell
eval set -- "$(getopt -o n: -- -n hello world)"; printf '%s|' "$@"
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox hd
Source: `test/e2e/tools/mimixbox/util-linux/hd.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
hd --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: hd`
- stderr is empty
### Scenario: hexdumps stdin in canonical form
#### Inputs
_stdin for `hd`:_
```
hi
```
#### When
```shell
hd
```
#### Then
- exit code is `0`
- stdout contains `68 69 0a`, `|hi.|`
## mimixbox util-linux --help helpers
Source: `test/e2e/tools/mimixbox/util-linux/help_helpers_util-linux.atago.yaml`
### Scenario: fallocate --help is structured
#### When
```shell
env -- fallocate --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: fsck.minix --help is structured
#### When
```shell
env -- fsck.minix --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: linux32 --help is structured
#### When
```shell
env -- linux32 --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: linux64 --help is structured
#### When
```shell
env -- linux64 --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: mkdosfs --help is structured
#### When
```shell
env -- mkdosfs --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: mkfs.ext2 --help is structured
#### When
```shell
env -- mkfs.ext2 --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: mkfs.minix --help is structured
#### When
```shell
env -- mkfs.minix --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: mkfs.reiser --help is structured
#### When
```shell
env -- mkfs.reiser --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: mkfs.vfat --help is structured
#### When
```shell
env -- mkfs.vfat --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: scriptreplay --help is structured
#### When
```shell
env -- scriptreplay --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: setsid --help is structured
#### When
```shell
env -- setsid --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: sh --help is structured
#### When
```shell
env -- sh --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: swapoff --help is structured
#### When
```shell
env -- swapoff --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
### Scenario: swapon --help is structured
#### When
```shell
env -- swapon --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`
## mimixbox hexdump / hd
Source: `test/e2e/tools/mimixbox/util-linux/hexdump.atago.yaml`
### Scenario: hd shows the canonical hex+ASCII layout
#### When
```shell
printf 'hello world\n' | hd
```
#### Then
- exit code is `0`
- stdout contains `00000000  68 65 6c 6c 6f 20 77 6f  72 6c 64 0a`, `|hello world.|`
### Scenario: hexdump -C matches hd
#### When
```shell
printf 'hello world\n' | hexdump -C
```
#### Then
- exit code is `0`
- stdout contains `|hello world.|`
### Scenario: hexdump default shows two-byte words
#### When
```shell
printf 'hello world\n' | hexdump
```
#### Then
- exit code is `0`
- stdout contains `6568 6c6c 206f`
## mimixbox hwclock
Source: `test/e2e/tools/mimixbox/util-linux/hwclock.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
hwclock --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: hwclock`, `RTC`
## mimixbox ionice
Source: `test/e2e/tools/mimixbox/util-linux/ionice.atago.yaml`
### Scenario: prints a process I/O class
#### When
```shell
ionice -p $$ | grep -cE 'prio|idle'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: runs a command at a given I/O class
#### When
```shell
ionice -c 3 -- echo idled
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox ipcrm
Source: `test/e2e/tools/mimixbox/util-linux/ipcrm.atago.yaml`
### Scenario: fails when nothing is requested
#### When
```shell
ipcrm
```
#### Then
- exit code is not `0`
### Scenario: fails to remove a non-existent id
#### When
```shell
ipcrm -q 2147483647
```
#### Then
- exit code is not `0`
## mimixbox ipcs
Source: `test/e2e/tools/mimixbox/util-linux/ipcs.atago.yaml`
### Scenario: shows the IPC facility sections
#### When
```shell
ipcs | grep -c 'Message Queues'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: limits to shared memory with -m
#### When
```shell
ipcs -m
```
#### Then
- exit code is `0`
- stdout contains `Shared Memory Segments`
- stdout does not contain `Message Queues`
## mimixbox last
Source: `test/e2e/tools/mimixbox/util-linux/last.atago.yaml`
### Scenario: treats an empty wtmp as no history and exits 0
#### When
```shell
out=$(last /dev/null); echo "[$out] rc=$?"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: fails on a missing wtmp file
#### When
```shell
last /no/such/mimixbox/wtmp
```
#### Then
- exit code is not `0`
## mimixbox losetup
Source: `test/e2e/tools/mimixbox/util-linux/losetup.atago.yaml`
### Scenario: lists active loop devices cleanly
#### When
```shell
losetup -a >/dev/null
```
#### Then
- exit code is `0`
### Scenario: refuses to associate a loop device
#### When
```shell
losetup /dev/loop0 /tmp/img
```
#### Then
- exit code is not `0`
## mimixbox lsattr
Source: `test/e2e/tools/mimixbox/util-linux/lsattr.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
lsattr --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: lsattr`, `attribute`
## mimixbox lsblk
Source: `test/e2e/tools/mimixbox/util-linux/lsblk.atago.yaml`
### Scenario: prints the column header
#### When
```shell
lsblk | sed -n '1p'
```
#### Then
- exit code is `0`
- stdout contains `NAME`, `SIZE`, `TYPE`
### Scenario: runs and exits successfully
#### When
```shell
lsblk >/dev/null 2>&1
```
#### Then
- exit code is `0`
## mimixbox lspci
Source: `test/e2e/tools/mimixbox/util-linux/lspci.atago.yaml`
### Scenario: lists PCI devices and exits successfully
#### When
```shell
lspci
```
#### Then
- exit code is `0`
## mimixbox lsusb
Source: `test/e2e/tools/mimixbox/util-linux/lsusb.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
lsusb --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: lsusb`, `USB`
## mimixbox mdev
Source: `test/e2e/tools/mimixbox/util-linux/mdev.atago.yaml`
### Scenario: requires scan mode
#### When
```shell
mdev
```
#### Then
- exit code is not `0`
### Scenario: describes itself with --help
#### When
```shell
mdev --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: mdev`, `device`
## mimixbox mesg
Source: `test/e2e/tools/mimixbox/util-linux/mesg.atago.yaml`
### Scenario: reports an error when stdin is not a terminal
#### When
```shell
mesg
```
#### Then
- exit code is `2`
- stderr contains `cannot get terminal name`
## mimixbox mke2fs / mkfs.ext2
Source: `test/e2e/tools/mimixbox/util-linux/mke2fs.atago.yaml`
### Scenario: writes the ext2 magic
#### When
```shell
dd if=/dev/zero of=e.img bs=1024 count=1024 2>/dev/null; mke2fs e.img >/dev/null; dd if=e.img bs=1 skip=1080 count=2 2>/dev/null | od -An -tx1 | tr -d ' \n'

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: mkfs.ext2 refuses an oversized image
#### When
```shell
dd if=/dev/zero of=b.img bs=1024 count=10000 2>/dev/null
mkfs.ext2 b.img
```
#### Then
- after `mkfs.ext2 b.img`:
  - exit code is not `0`
## mimixbox mkfs.minix
Source: `test/e2e/tools/mimixbox/util-linux/mkfs_minix.atago.yaml`
### Scenario: writes the Minix v1 magic
#### When
```shell
dd if=/dev/zero of=m.img bs=1024 count=2048 2>/dev/null; mkfs.minix m.img >/dev/null; dd if=m.img bs=1 skip=1040 count=2 2>/dev/null | od -An -tx1 | tr -d ' \n'

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: refuses a too-small device
#### When
```shell
dd if=/dev/zero of=s.img bs=1024 count=4 2>/dev/null
mkfs.minix s.img
```
#### Then
- after `mkfs.minix s.img`:
  - exit code is not `0`
## mimixbox mkfs.reiser
Source: `test/e2e/tools/mimixbox/util-linux/mkfs_reiser.atago.yaml`
### Scenario: refuses deterministically
#### When
```shell
mkfs.reiser /tmp/x.img
```
#### Then
- exit code is not `0`
### Scenario: explains that ReiserFS is deprecated
#### When
```shell
mkfs.reiser /tmp/x.img
```
#### Then
- exit code is not `0`
- stderr contains `deprecated`
## mimixbox mkfs.vfat / mkdosfs
Source: `test/e2e/tools/mimixbox/util-linux/mkfs_vfat.atago.yaml`
### Scenario: writes the FAT16 type label
#### When
```shell
dd if=/dev/zero of=v.img bs=1024 count=8192 2>/dev/null; mkfs.vfat v.img >/dev/null; dd if=v.img bs=1 skip=54 count=5 2>/dev/null

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: mkdosfs refuses a too-small image
#### When
```shell
dd if=/dev/zero of=s.img bs=1024 count=512 2>/dev/null
mkdosfs s.img
```
#### Then
- after `mkdosfs s.img`:
  - exit code is not `0`
## mimixbox mkswap
Source: `test/e2e/tools/mimixbox/util-linux/mkswap.atago.yaml`
### Scenario: formats an image as swap
#### When
```shell
dd if=/dev/zero of=swap.img bs=1024 count=64 2>/dev/null; chmod 0600 swap.img; mkswap swap.img

```
#### Then
- exit code is `0`
- stdout contains `version 1`
### Scenario: writes the swap signature
#### When
```shell
dd if=/dev/zero of=swap2.img bs=1024 count=64 2>/dev/null; chmod 0600 swap2.img; mkswap swap2.img >/dev/null; od -c swap2.img | grep -c 'S   W   A   P   S   P   A   C   E   2'

```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox mount
Source: `test/e2e/tools/mimixbox/util-linux/mount.atago.yaml`
### Scenario: lists the root filesystem
#### When
```shell
mount | grep -cE ' on / type '
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: refuses to perform a mount
#### When
```shell
mount /dev/sda1 /mnt
```
#### Then
- exit code is not `0`
## mimixbox nsenter
Source: `test/e2e/tools/mimixbox/util-linux/nsenter.atago.yaml`
### Scenario: requires a target PID
#### When
```shell
nsenter -n echo x
```
#### Then
- exit code is not `0`
### Scenario: requires a namespace flag
#### When
```shell
nsenter -t 1 echo x
```
#### Then
- exit code is not `0`
## mimixbox pivot_root
Source: `test/e2e/tools/mimixbox/util-linux/pivot_root.atago.yaml`
### Scenario: requires two directories
#### When
```shell
pivot_root /onlyone
```
#### Then
- exit code is not `0`
### Scenario: describes itself with --help
#### When
```shell
pivot_root --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: pivot_root`, `root`
## mimixbox rdate
Source: `test/e2e/tools/mimixbox/util-linux/rdate.atago.yaml`
### Scenario: fails when no host is given
#### When
```shell
rdate
```
#### Then
- exit code is not `0`
### Scenario: fails when the host has no time service
#### When
```shell
rdate 127.0.0.1
```
#### Then
- exit code is not `0`
## mimixbox rdev
Source: `test/e2e/tools/mimixbox/util-linux/rdev.atago.yaml`
### Scenario: prints the root device with the / mountpoint
#### When
```shell
rdev
```
#### Then
- exit code is `0`
- stdout contains ` /`
## mimixbox readprofile
Source: `test/e2e/tools/mimixbox/util-linux/readprofile.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
readprofile --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: readprofile`, `profiling`
## mimixbox renice
Source: `test/e2e/tools/mimixbox/util-linux/renice.atago.yaml`
### Scenario: reports the priority change
#### When
```shell
renice -n 0 -p $$ | grep -c 'process ID'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: rejects a non-numeric PID
#### When
```shell
renice 5 notapid
```
#### Then
- exit code is not `0`
## mimixbox rtcwake
Source: `test/e2e/tools/mimixbox/util-linux/rtcwake.atago.yaml`
### Scenario: rejects a suspend mode
#### When
```shell
rtcwake -m mem -s 10
```
#### Then
- exit code is not `0`
### Scenario: requires a wake time
#### When
```shell
rtcwake -m no
```
#### Then
- exit code is not `0`
## mimixbox script / scriptreplay
Source: `test/e2e/tools/mimixbox/util-linux/script.atago.yaml`
### Scenario: records command output to a typescript
#### When
```shell
script -c 'printf recorded' -T timing out.txt >/dev/null 2>&1; grep -c 'Script started' out.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: replays a recorded typescript
#### When
```shell
script -c 'printf replayed' -T timing out.txt >/dev/null 2>&1; scriptreplay timing out.txt

```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox script / scriptreplay round-trip
Source: `test/e2e/tools/mimixbox/util-linux/script_roundtrip.atago.yaml`
### Scenario: records the transcript framing
#### When
```shell
script -c 'printf "hello\nworld\n"' -T timing transcript >/dev/null 2>&1; cat transcript
```
#### Then
- exit code is `0`
- stdout contains `Script started`, `Script done`
### Scenario: records the command payload in the transcript
#### When
```shell
script -c 'printf "hello\nworld\n"' -T timing transcript >/dev/null 2>&1; cat transcript
```
#### Then
- exit code is `0`
- stdout contains `hello`, `world`
### Scenario: writes a timing file of "delay bytes" records
#### When
```shell
script -c 'printf "hello\nworld\n"' -T timing transcript >/dev/null 2>&1; awk 'NF && $1 ~ /^[0-9]+\.[0-9]+$/ && $2 ~ /^[0-9]+$/ { ok++ } END { exit !(NR>0 && ok==NR) }' timing
```
#### Then
- exit code is `0`
### Scenario: replays the captured payload from the timing + transcript
#### When
```shell
script -c 'printf "hello\nworld\n"' -T timing transcript >/dev/null 2>&1; scriptreplay timing transcript
```
#### Then
- exit code is `0`
- stdout contains `hello`, `world`
## mimixbox scriptreplay
Source: `test/e2e/tools/mimixbox/util-linux/scriptreplay.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
scriptreplay --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: scriptreplay`
- stderr is empty
## mimixbox setarch / linux32 / linux64
Source: `test/e2e/tools/mimixbox/util-linux/setarch.atago.yaml`
### Scenario: linux32 makes uname report a 32-bit machine
#### When
```shell
linux32 uname -m
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: linux64 reports the native machine
#### When
```shell
linux64 uname -m
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: linux32 passes the command output through
#### When
```shell
linux32 echo passed
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: setarch selects the personality from ARCH
#### When
```shell
setarch i686 uname -m
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox setpriv
Source: `test/e2e/tools/mimixbox/util-linux/setpriv.atago.yaml`
### Scenario: dumps the current privileges
#### When
```shell
setpriv --dump | grep -cE 'uid:|no_new_privs:'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: runs a command with --no-new-privs
#### When
```shell
setpriv --no-new-privs -- echo ran
```
#### Then
- exit code is `0`
- stdout equals an exact value
## mimixbox setsid
Source: `test/e2e/tools/mimixbox/util-linux/setsid.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
setsid --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: setsid`
- stderr is empty
## mimixbox setsid / fallocate
Source: `test/e2e/tools/mimixbox/util-linux/setsid_fallocate.atago.yaml`
### Scenario: setsid runs a program in a new session
#### When
```shell
setsid echo "session ok"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: fallocate sizes a file to the requested length
#### When
```shell
fallocate -l 4096 f
wc -c < f
```
#### Then
- after `fallocate -l 4096 f`:
  - exit code is `0`
- after `wc -c < f`:
  - stdout contains `4096`
### Scenario: fallocate without -l fails
#### When
```shell
fallocate x
```
#### Then
- exit code is not `0`
## mimixbox swapon / swapoff
Source: `test/e2e/tools/mimixbox/util-linux/swap.atago.yaml`
### Scenario: swapon -s prints the swaps header
#### When
```shell
swapon -s | sed -n '1p' | grep -c 'Filename'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: swapoff requires a target
#### When
```shell
swapoff
```
#### Then
- exit code is not `0`
## mimixbox swapoff
Source: `test/e2e/tools/mimixbox/util-linux/swapoff.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
swapoff --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: swapoff`
- stderr is empty
## mimixbox swapon
Source: `test/e2e/tools/mimixbox/util-linux/swapon.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
swapon --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: swapon`
- stderr is empty
## mimixbox switch_root
Source: `test/e2e/tools/mimixbox/util-linux/switch_root.atago.yaml`
### Scenario: requires NEW_ROOT and INIT
#### When
```shell
switch_root /tmp
```
#### Then
- exit code is not `0`
### Scenario: rejects a non-directory NEW_ROOT
#### When
```shell
switch_root /no/such/dir /init
```
#### Then
- exit code is not `0`
## mimixbox taskset
Source: `test/e2e/tools/mimixbox/util-linux/taskset.atago.yaml`
### Scenario: prints a process affinity mask
#### When
```shell
taskset -p $$ | grep -c 'affinity mask'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: runs a command bound to a CPU
#### When
```shell
taskset -c 0 echo affined
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: rejects an invalid mask
#### When
```shell
taskset zzz echo x
```
#### Then
- exit code is not `0`
## mimixbox tune2fs
Source: `test/e2e/tools/mimixbox/util-linux/tune2fs.atago.yaml`
### Scenario: rejects a non-ext image
#### Given
- Fixture file `bad.img` is created.
#### Inputs
_Fixture `bad.img`:_
```
not a filesystem
```
#### When
```shell
tune2fs -l bad.img
```
#### Then
- exit code is not `0`
### Scenario: describes itself with --help
#### When
```shell
tune2fs --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: tune2fs`, `filesystem`
## mimixbox uevent
Source: `test/e2e/tools/mimixbox/util-linux/uevent.atago.yaml`
### Scenario: describes itself with --help
#### When
```shell
uevent --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: uevent`, `uevent`
## mimixbox umount
Source: `test/e2e/tools/mimixbox/util-linux/umount.atago.yaml`
### Scenario: fails for a target that is not mounted
#### When
```shell
umount /not/a/real/mountpoint
```
#### Then
- exit code is not `0`
### Scenario: requires a target
#### When
```shell
umount
```
#### Then
- exit code is not `0`
## mimixbox unshare
Source: `test/e2e/tools/mimixbox/util-linux/unshare.atago.yaml`
### Scenario: requires a namespace flag
#### When
```shell
unshare echo x
```
#### Then
- exit code is not `0`
### Scenario: describes itself with --help
#### When
```shell
unshare --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: unshare`, `namespace`
## mimixbox wall
Source: `test/e2e/tools/mimixbox/util-linux/wall.atago.yaml`
### Scenario: runs and exits successfully
#### When
```shell
echo "broadcast" | wall
```
#### Then
- exit code is `0`