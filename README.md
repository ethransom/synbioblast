# SynBioBLAST

SynBioBLAST is a standalone website that adds support for NCBI's BLAST to SynBioHub.

## Setup

1. Clone the repo
2. Make sure you have the [`go`](https://golang.org) compiler installed.
3. Install the dependencies
   ```
   $ go get github.com/foobar/barfoo
   $ go get github.com/other/other
   ```
4. Build the slurper
   ```
   $ go build slurper.go
   ```
5. Build the queryserver
   ```
   $ go build synbioblast.go
   ```
6. If necessary, copy the binaries and `builddb.sh` to the computer you want to run SynBioBLAST from.
7. (On the SynBioBLAST computer.) Make sure redis is installed.
8. Run the slurper (Should only take a few minutes to complete.)
    ```
    $ ./slurper
    ```
9. Build the blast db
    ```
    $ ./builddb.sh
    ```
10. Run the query server
    ```
    $ ./server
    ```
11. Navigate to the appropriate address and port to see synbioblast.

## Overview

![](https://github.com/schnauzer/synbioblast/raw/master/architecture.svg "Overview of architecture")

### Slurper (`slurper.go`)

Pulls records from SynBioHub, stores fastas in configured fasta directory, stores deduplication information in Redis.

### DB Builder (`builddb.sh`)

Intended to run occasionally as a cron job.

Builds the fastas in the configured fasta directory into a BLAST database.

### Queryserver (`synbioblast.go`)

Serves HTTP. Spawns a blast child process to run queries against the BLAST database.