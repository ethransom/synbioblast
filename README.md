# SynBioBLAST

SynBioBLAST is a standalone website that adds support for NCBI's BLAST to SynBioHub.

[Project proposal.](https://github.com/schnauzer/synbioblast/blob/master/Project%20Proposal--SynBioBlast.pdf)

[Update 1.](https://github.com/schnauzer/synbioblast/blob/master/update_1.md)

[Update 2.](https://github.com/schnauzer/synbioblast/blob/master/update_2.md)

[Final Report.](https://github.com/schnauzer/synbioblast/blob/master/final_report.pdf)

## Setup

While Go is multi-platform, the included executables are compiled for Linux, so for the 
time being Linux is the only supported platform for SynBioBLAST.

1. Clone the repo
2. Make sure you have the [`go`](https://golang.org) compiler and Redis installed.
3. Install the dependencies
   ```
   $ go get github.com/knakk/sparql
   $ go get github.com/mediocregopher/radix.v2
   $ go get github.com/spacemonkeygo/flagfile
   ```
4. Build the slurper
   ```
   $ go build slurper.go
   ```
5. Build the queryserver
   ```
   $ go build synbioblast.go
   ```
7. Run the slurper (Should only take a few minutes to complete.)
    ```
    $ ./slurper -flagfile synbioblast.flags
    ```
8. Build the blast db
    ```
    $ SYNBIOBLASTDIR=$PWD ./builddb.sh
    ```
9. Run the query server
    ```
    $ ./synbioblast -flagfile synbioblast.flags
    ```
10. Navigate to SynBioBLAST with your favorite browser. By default it is on port 9090.

## Overview

![](https://github.com/schnauzer/synbioblast/raw/master/actualarchitecture.png "Overview of architecture")

### Slurper ([`slurper.go`](https://github.com/schnauzer/synbioblast/blob/master/slurper.go))

SynBioHub uses the Virtuoso database to store application state. It exposes an
endpoint for readonly queries at [https://synbiohub.org/sparql](https://synbiohub.org/sparql). This endpoint returns XML if the `Accept` header
is not present. An example response can be found in `virtuosooutput.xml`.

With these records, the slurper performs some simple deduplication. Sequences are hashed with SHA1. This hash becomes the primary identifier for the unique sequence.

The sequences are written to fasta files named and identified with their hash. These files are stored in a configurable fasta directory.

Redis is used to store the deduplication information. For each sequence processed,
its uri is added to a set keyed with the hash of the sequence. This set then 
becomes a list of urls for each sequence with this hash encountered.

### DB Builder ([`builddb.sh`](https://github.com/schnauzer/synbioblast/blob/master/builddb.sh))

Intended to run occasionally (perhaps nightly or hourly) as a cron job.

Builds the fastas in the configured fasta directory into a BLAST database.

### Queryserver ([`synbioblast.go`](https://github.com/schnauzer/synbioblast/blob/master/synbioblast.go))

Serves HTTP. Spawns a blast child process to run queries against the BLAST database.

## Future Work

 * The DB Builder does not operate atomically. It should build a new database
   in a temporary location and then atomically rename it so as not to disrupt any
   incoming queries.

 * The overhead of reading hundreds of thousands of small files is a huge
   performance bottleneck in the building of the BLAST database. If the slurper
   concatenated these files together we could cut the time needed by 10x.

 * Deduplication information takes up less room than originally anticipated. The 
   slurper could perform dedup in-memory and write this information out along with the fasta files, eliminating the need for a redis request to translate hashes
   into sequences.

 * Rewrite into JavaScript and merge into SynBioHub.