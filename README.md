# sarifw (SARIF wrapper)

This project converts the output of `ripgrep` and `ast-grep` into SARIF format.
You can convert the output to SARIF format by prefixing these commands with `sarifw`.

## Motivation

ripgrep and ast-grep are lightweight and very useful tools.
However, triaging a large number of detection results can be challenging.

In such scenarios, using tools like [SARIF Explorer](https://github.com/trailofbits/vscode-sarif-explorer) can help reduce the workload.
Therefore, I created a tool to convert the output of ripgrep and ast-grep into SARIF files.

## Installation

```
go install github.com/lambdasawa/sarifw@latest
```

## Usage

```
sarifw rg 'console.log'
sarifw sg --pattern 'console.log'
sarifw sg run --pattern 'console.log'
sarifw sg scan 'console.log'
sarifw ast-grep scan 'console.log'
```

Internally, it simply executes the `rg` and `sg` commands with the --json option appended.
Therefore, you can use any options implemented in the `rg` and `sg` commands, not just the ones shown in these examples.
