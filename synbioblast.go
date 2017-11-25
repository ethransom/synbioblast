package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/mediocregopher/radix.v2/redis"
	"github.com/spacemonkeygo/flagfile"
)

// TODO: deduplicate these
var (
	blastdbDir = flag.String("blastdb.path", "/var/synbioblast/blastdbs",
		"directory where blast dbs are stored")
	blastdbName = flag.String("blastdb.name", "SynBioHub", "name of the blast db to use")

	redisURL          = flag.String("redis.url", "localhost:6379", "URL of redis instance storing dedup state")
	redisSeqSetPrefix = flag.String("redis.sequencePrefix", "sequence",
		"Redis key prefix, appended with hash of sequence to store set of matching components")

	fastaDir = flag.String("fastas.path", "/var/synbioblast/fastas", "path to store fasta files in")
)

// BlastResults represents the result of running a blast query
type BlastResults struct {
	Query   string
	Error   string
	Results []*blastResult
}

// http://www.metagenomics.wiki/tools/blast/blastn-output-format-6
type blastResult struct {
	Qseqid   string // query sequence id
	Sseqid   string // subject sequence id
	Pident   string // percentage of identical matches
	Length   string // alignment length
	Mismatch string // number of mismatches
	Gapopen  string // number of gap openings
	Qstart   string // start of alignment in query
	Qend     string // end of alignment in query
	Sstart   string // start of alignment in subject
	Send     string // end of alignment in subject
	Evalue   string // expect value
	Bitscore string // bit score

	URIs []string
}

func (r *blastResult) getURIs() error {
	key := *redisSeqSetPrefix + ":" + r.Sseqid

	uris, err := redisClient.Cmd("SMEMBERS", key).List()
	if err != nil {
		return err
	}

	r.URIs = uris

	return nil
}

func newBlastResult(record []string) *blastResult {
	return &blastResult{
		Qseqid:   record[0],
		Sseqid:   record[1],
		Pident:   record[2],
		Length:   record[3],
		Mismatch: record[4],
		Gapopen:  record[5],
		Qstart:   record[6],
		Qend:     record[7],
		Sstart:   record[8],
		Send:     record[9],
		Evalue:   record[10],
		Bitscore: record[11],
	}
}

func parseResults(b []byte) ([]*blastResult, error) {
	r := csv.NewReader(bytes.NewReader(b))

	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	results := make([]*blastResult, 0, len(records))
	for _, record := range records {
		result := newBlastResult(record)

		err := result.getURIs()
		if err != nil {
			return []*blastResult{}, nil
		}

		results = append(results, result)
	}

	return results, nil
}

// Blast runs a blast query with the given target sequence.
func Blast(seq string) (*BlastResults, error) {
	cmd := exec.Command("./blastn", "-db", *blastdbName, "-outfmt", "10")
	path := os.ExpandEnv("PATH=$PATH:$PWD")
	blastdb := "BLASTDB=" + os.ExpandEnv(*blastdbDir)
	cmd.Env = append(os.Environ(), path, blastdb)
	log.Printf("running command with db %s", blastdb)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, seq)
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return &BlastResults{Error: string(out), Query: seq}, err
	}

	// TODO: this might be redundant to the err != nil above, investigate
	if cmd.ProcessState.Success() {
		log.Printf("executed successfully")
	} else {
		log.Printf("did not execute successfully")
	}

	results, err := parseResults(out)
	if err != nil {
		return nil, err
	}

	return &BlastResults{Results: results, Query: seq}, nil
}

// https://golang.org/doc/articles/wiki/

var templates = template.Must(template.ParseFiles("form.html", "blast.html"))

func indexHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "form.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func blastHandler(w http.ResponseWriter, r *http.Request) {
	seq := r.FormValue("seq")

	result, err := Blast(seq)
	if err != nil {
		log.Printf("ERROR blast: %v: %s", err, result.Results)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = templates.ExecuteTemplate(w, "blast.html", *result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

var redisClient *redis.Client

func main() {
	flagfile.Load()

	var err error
	redisClient, err = redis.Dial("tcp", *redisURL)
	if err != nil {
		log.Fatal("couldn't dial redis")
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/blast/", blastHandler)
	http.ListenAndServe(":9090", nil)
}
