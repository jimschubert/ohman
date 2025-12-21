# ohman

[![Apache 2.0 License](https://img.shields.io/badge/License-Apache%202.0-blue)](./LICENSE)
![Go Version](https://img.shields.io/github/go-mod/go-version/jimschubert/ohman)
[![Go Build](https://github.com/jimschubert/ohman/actions/workflows/build.yml/badge.svg)](https://github.com/jimschubert/ohman/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jimschubert/ohman)](https://goreportcard.com/report/github.com/jimschubert/ohman)

> Oh, Manning. You don't need to create so many duplicate files in my Dropbox.
> So many, in fact, that I need to write a tool to clean it up.

A small CLI tool to find and optionally remove duplicate files that follow a common "name (n).ext" pattern (for example `book.pdf`, `book (1).pdf`). It's intended to help clean up duplicate files created by sync services, downloads, or other processes.

## Features
- Find duplicate files matching a configurable regex pattern.
- Dry-run mode to preview matches (no filesystem changes).
- Delete duplicates while keeping the original, or keep the newest with `--inverse`.
- Optionally rename the kept newest duplicate back to the original filename (`--inverse-and-rename`).
- Output results to a file via `--out` (defaults to `results.txt` when deleting and `--out` not provided).

## Quick Install

Build locally with Go:

```bash
go build -o ohman ./...
```

Or install using `go install`:

```bash
go install ./...
```

## Usage

Basic command structure:

```bash
ohman [flags] <path>...
```

Common examples:

- Dry-run, list duplicate files to stdout (no deletions):

```bash
ohman --dryrun /path/to/search
```

Example:

```shell
$ ohman --dry-run '/Volumes/jim/Dropbox/Apps/Manning Books/Secure by Design'
Original: /Volumes/jim/Dropbox/Apps/Manning Books/Secure by Design/Secure_by_Design.pdf
  - Duplicate: /Volumes/jim/Dropbox/Apps/Manning Books/Secure by Design/Secure_by_Design (1).pdf
  - Duplicate: /Volumes/jim/Dropbox/Apps/Manning Books/Secure by Design/Secure_by_Design (2).pdf
```

- Delete duplicates and write results to a file (relative to your current directory or use `--out`):

```bash
ohman --delete --out results.txt /path/to/search
```

- Inverse deletion: keep only the newest file and delete the rest:

```bash
ohman --inverse --delete /path/to/search
```

- Inverse delete + rename: keep the newest file and rename it to remove the duplicate marker (e.g. `book (2).pdf` -> `book.pdf`):

```bash
ohman --inverse-and-rename --delete --out kept.txt /path/to/search
```

Example:
```shell
$ stat -f "%m%t%Sm %N" * | sort -rn | cut -f2- | grep pdf
Dec 19 16:42:44 2025 The_Tao_of_Microservices (4).pdf
Apr 25 16:40:11 2020 The_Tao_of_Microservices (3).pdf
Mar  8 13:26:45 2020 The_Tao_of_Microservices (2).pdf
Mar  1 01:06:58 2020 The_Tao_of_Microservices (1).pdf
Dec 31 21:43:15 2018 The_Tao_of_Microservices.pdf

$ ohman --inverse-and-rename --delete .
Results written to results.txt

$ cat results.txt 
Deleted /Volumes/jim/Dropbox/Apps/Manning Books/The Tao of Microservices/The_Tao_of_Microservices (3).pdf
Deleted /Volumes/jim/Dropbox/Apps/Manning Books/The Tao of Microservices/The_Tao_of_Microservices (2).pdf
Deleted /Volumes/jim/Dropbox/Apps/Manning Books/The Tao of Microservices/The_Tao_of_Microservices (1).pdf
Deleted /Volumes/jim/Dropbox/Apps/Manning Books/The Tao of Microservices/The_Tao_of_Microservices.pdf
Renamed /Volumes/jim/Dropbox/Apps/Manning Books/The Tao of Microservices/The_Tao_of_Microservices (4).pdf to /Volumes/jim/Dropbox/Apps/Manning Books/The Tao of Microservices/The_Tao_of_Microservices.pdf
```

## Flags
- `--out, -o <file>` — Write results to the specified file. When `--delete` is used and `--out` is omitted, `results.txt` in the current working directory is used.
- `--regex <pattern>` — Custom regular expression for matching duplicate filenames. USE AT YOUR OWN RISK: a poorly chosen regex may match unintended files or cause surprising behavior; test with `--dryrun` first.
- `--delete` — Actually delete matched duplicate files. Omit to perform a dry-run.
- `--dryrun` — Explicit dry-run mode (prints matches only).
- `--inverse` — When deleting, keep the newest file and delete the older/original ones instead.
- `--inverse-and-rename` — Keep the newest and rename it to the canonical original name.

## Default regex

 The default regex used by `ohman` looks for patterns like `name (N).ext` and matches these extensions by default:

```
(.+)\s\((\d+)\)\.(pdf|mobi|mp4|epub|wav|mp3)$
```

(You can override this with `--regex`, but again: MODIFY THIS AT YOUR OWN RISK.)

## Testing

Run the unit tests:

```bash
go test ./...
```

## Contributing
- Please open issues or pull requests on the repository.
- Run the tests and add new tests for bug fixes or features.

## License
- See `LICENSE` for license terms (Apache 2.0).

## Acknowledgements

Built with [Go](https://github.com/golang/go/) and [Kong](https://github.com/alecthomas/kong). Thanks to the OSS ecosystem.
