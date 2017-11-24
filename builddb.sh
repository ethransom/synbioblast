#!/bin/bash

BLASTDB="${BLASTDB:-/var/synbioblast/blastdbs}" 
echo "Storing blastdb files in $BLASTDB"

DBNAME="${DBNAME:-SynBioHub}"
echo "Using db name of $DBNAME"

TITLE="$DBNAME (generated $(date))"

awk 'FNR==1{print ""}1' fastas/*.fasta  | ./makeblastdb -dbtype nucl -title "$TITLE" -out "$BLASTDB/$DBNAME" -in -
