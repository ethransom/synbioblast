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

SynBioHub uses the Virtuoso database to store application state. It exposes an
endpoint for readonly queries at [https://synbiohub.org/sparql](https://synbiohub.org/sparql). This endpoint returns XML if the `Accept` header
is not present. An example response can be found in `virtuosooutput.xml`.

With these records, the slurper performs some simple deduplication. Sequences are hashed with SHA1. This hash becomes the primary identifier for the unique sequence.

The sequences are written to fasta files named and identified with their hash. These files are stored in a configurable fasta directory.

Redis is used to store the deduplication information. For each sequence processed,
its uri is added to a set keyed with the hash of the sequence. This set then 
becomes a list of urls for each sequence with this hash encountered.

### DB Builder (`builddb.sh`)

Intended to run occasionally (perhaps nightly or hourly) as a cron job.

Builds the fastas in the configured fasta directory into a BLAST database.

### Queryserver (`synbioblast.go`)

Serves HTTP. Spawns a blast child process to run queries against the BLAST database.

## Future Work

 * The DB Builder does not operate atomically. It should build a new database
   in a temporary location and then atomically rename it so as not to disrupt any
   incoming queries.

 * Deduplication information takes up less room than originally anticipated. The 
   slurper could perform dedup in-memory and write this information out along with the fasta files, eliminating the need for a redis request to translate hashes
   into sequences.

 * Rewrite into JavaScript and merge into SynBioHub.