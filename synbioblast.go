package main

import (
	"encoding/xml"
	"flag"
	"fmt"
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
	XMLName   xml.Name `xml:"BlastOutput"`
	Version   string   `xml:"BlastOutput_version"`
	Reference string   `xml:"BlastOutput_reference"`

	// TODO: parameters?

	Results []blastResult `xml:"BlastOutput_iterations>Iteration>Iteration_hits>Hit"`

	DBNum int `xml:"BlastOutput_iterations>Iteration>Iteration_stat>Statistics>Statistics_db-num"`

	Query string
	Error string
}

type blastResult struct {
	SeqHash string `xml:"Hit_def"`

	BitScore float64 `xml:"Hit_hsps>Hsp>Hsp_bit-score"`
	Score    int     `xml:"Hit_hsps>Hsp>Hsp_score"`
	EValue   string  `xml:"Hit_hsps>Hsp>Hsp_evalue"`

	QuerySeq string `xml:"Hit_hsps>Hsp>Hsp_qseq"`
	Midline  string `xml:"Hit_hsps>Hsp>Hsp_midline"`
	HitSeq   string `xml:"Hit_hsps>Hsp>Hsp_hseq"`

	URIs []string
}

func (r *blastResult) getURIs() error {

	key := *redisSeqSetPrefix + ":" + r.SeqHash

	uris, err := redisClient.Cmd("SMEMBERS", key).List()
	if err != nil {
		return err
	}

	r.URIs = uris

	return nil
}

func parseResults(b []byte) (*BlastResults, error) {
	results := &BlastResults{}
	err := xml.Unmarshal(b, &results)
	if err != nil {
		return nil, err
	}

	for i := range results.Results {
		err = results.Results[i].getURIs()
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

// Blast runs a blast query with the given target sequence.
func Blast(seq string) (*BlastResults, error) {
	cmd := exec.Command("./blastn", "-db", *blastdbName, "-outfmt", "5")
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
		println("MARK")
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

	fmt.Printf("URIs: %+v\n", results.Results[0].URIs)

	results.Query = seq

	return results, nil
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
		log.Printf("ERROR blast: %v: %+v", err, result)
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
