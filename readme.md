# rsum

Rsum is a file hashing tool roughly styled after the *sha*x*sum* tools in GNU Coreutils. Rsum supports SHA512 and Blake2b 512 bit hashes, recursive scanning into directories, file-level resume of interrupted hashing sessions, and will process multiple small files in parallel.

In SHA512 mode, `rsum` should be signficiantly faster than `sha512sum` when hashing collections of files smaller than 16 MiB, but `sha512sum` handles very large files more quickly due to its more efficient SHA512 implemenation. `Rsum` in Blake2b mode is faster than SHA512.

256 bit hashing algorithms are not supported as they perform badly on 64-bit platforms without hardware SHA acceleration, which Go does not support. Older hashing algorithms are not supported due to poor security.

The intended use case for `rsum` is to create hash snapshots for use by other tools, such as `gsc`. These snapshots can be used for detecting silent data corruption ('bit rot'), verifying the integrity of copies or backups, intrusion detection, and so forth.

## Usage

```
> rsum --help
rsum release 0. THIS IS EXPERIMENTAL SOFTWARE. DO NOT USE IT FOR ANYTHING IMPORTANT.

Usage: rsum --output=STRING <paths> ... [flags]

Recursive Sum Tool

Arguments:
  <paths> ...    Input files

Flags:
  -h, --help                      Show context-sensitive help.
  -r, --recurse                   Recurse into subdirectories.
  -t, --hashtype="blake2b-512"    Select hash algorithm. Valid options are blake2b-512 and sha512.
      --base64                    Write sums in RFC 4648 base 64 rather than hexadecimal.
  -v, --verbose                   Show progress.
  -c, --usecwd                    Store resume database in the current working directory rather than in the user cache directory.
  -o, --output=STRING             Output filename

```
## Usage notes

Rsum will accept directories among its input arguments, but *will not* recurse into subdirectories without the `-r` switch.

The `-c` switch is intended to allow multiple concurrent instances of rsum use separate resume databases to independently recover from unplanned failures.

## License

Copyright 2024 Coridon Henshaw

Permission is granted to all natural persons to execute, distribute, and/or modify this software (including its documentation) subject to the following terms:

1. Subject to point \#2, below, **all commercial use and distribution is prohibited.** This software has been released for personal and academic use for the betterment of society through any purpose that does not create income or revenue. *It has not been made available for businesses to profit from unpaid labor.*

2. Re-distribution of this software on for-profit, public use, repository hosting sites (for example: Github) is permitted provided no fees are charged specifically to access this software.

3. **This software is provided on an as-is basis and may only be used at your own risk.** This software is the product of a single individual's recreational project. The author does not have the resources to perform the degree of code review, testing, or other verification required to extend any assurances that this software is suitable for any purpose, or to offer any assurances that it is safe to execute without causing data loss or other damage.

4. **This software is intended for experimental use in situations where data loss (or any other undesired behavior) will not cause unacceptable harm.** Users with critical data safety needs must not use this software and, instead, should use equivalent tools that have a proven track record.

5. If this software is redistributed, this copyright notice and license text must be included without modification.

6. Distribution of modified copies of this software is discouraged but is not prohibited. It is strongly encouraged that fixes, modifications, and additions be submitted for inclusion into the main release rather than distributed independently.

7. This software reverts to the public domain 10 years after its final update or immediately upon the death of its author, whichever happens first.