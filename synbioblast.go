package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

// TODO: ENV or flag or flagfile these
const (
	blastdbDir = "blastdbs"
	// blastdb    = "16SMicrobial"
	blastdb = "SynBioHub"
)

// BlastResult represents the result of running a blast query
type BlastResult struct {
	Query   string
	Results string
}

// Blast runs a blast query with the given target sequence.
func Blast(seq string) (*BlastResult, error) {
	cmd := exec.Command("./blastn", "-db", blastdb)
	path := os.ExpandEnv("PATH=$PATH:$PWD")
	blastdb := os.ExpandEnv("BLASTDB=$PWD/" + blastdbDir)
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
		return &BlastResult{Results: string(out), Query: seq}, err
	}
	if cmd.ProcessState.Success() {
		log.Printf("executed successfully")
	} else {
		log.Printf("did not execute successfully")
	}
	return &BlastResult{Results: string(out), Query: seq}, nil
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

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/blast/", blastHandler)
	http.ListenAndServe(":9090", nil)
}
