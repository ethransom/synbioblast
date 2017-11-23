package main

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/knakk/sparql"
)

// paginated with a scollable cursor as per:
// http://blog.mynarz.net/2016/06/on-generating-sparql.html
const query = `
# tag: fetch
PREFIX dcterms: <http://purl.org/dc/terms/>
PREFIX sbol: <http://sbols.org/v2#> 

SELECT
	?uri 
	?elements
	?created 
WHERE { 
	{
		SELECT 
			?uri 
			?elements
			?created 
		WHERE { 
			?uri a sbol:ComponentDefinition . 
			?uri sbol:sequence ?sequenceUri . 
			?sequenceUri sbol:elements ?elements . 
			?uri dcterms:created ?created .
		} ORDER BY ASC(str(?created))
	}
}
LIMIT {{.Limit}} OFFSET {{.Offset}}
`

type queryParams struct {
	Limit, Offset int
}

const (
	synbiohubURL = "https://synbiohub.org/sparql"
	resultLimit  = 1000
)

// I couldn't find a way to match an element with an attribute
// with a given value, otherwise we could parse directly
// into a []sequence
type sparqlResult struct {
	XMLName   xml.Name   `xml:"sparql"`
	Variables []variable `xml:"head>variable"`
	Results   []result   `xml:"results>result"`
}

type variable struct {
	Name string `xml:"name,attr"`
}

type result struct {
	Bindings []binding `xml:"binding"`
}

func (r *result) getValue(name string) string {
	for _, b := range r.Bindings {
		if b.Name == name {
			return b.Value
		}
	}

	return ""
}

type binding struct {
	Name     string `xml:"name,attr"`
	Value    string `xml:",any"`
	Datatype string `xml:",any,attr"`
}

type sequence struct {
	URI      string
	Sequence string
	Created  time.Time
}

func parseSparqlTime(s string) (time.Time, error) {
	// this is way less complicated than I thought it would be
	return time.Parse(time.RFC3339, s)
}

func main() {
	bytes := fetch(0)

	println("fetched, parsing response...")

	seqs := parse(bytes)

	println("fetched, processing")

	process(seqs)
}

func parse(bytes []byte) []sequence {
	result := &sparqlResult{}
	err := xml.Unmarshal(bytes, &result)
	if err != nil {
		log.Fatal("couldn't parse xml: ", err)
	}

	// TODO: check if result.variables is correct?

	sequences := make([]sequence, len(result.Results))
	for i, result := range result.Results {
		sequences[i].URI = result.getValue("uri")
		sequences[i].Sequence = result.getValue("elements")
		t, err := parseSparqlTime(result.getValue("created"))
		if err != nil {
			log.Fatal("couldn't parse time: ", result.getValue("created"))
		}
		sequences[i].Created = t
	}

	return sequences
}

func fetch(offset int) []byte {
	config := &queryParams{
		Limit:  resultLimit,
		Offset: offset,
	}

	buf := bytes.NewBufferString(query)
	bank := sparql.LoadBank(buf)

	q, err := bank.Prepare("fetch", config)
	if err != nil {
		log.Fatal("couldn't prepare query: ", err)
	}

	vals := url.Values{}
	vals.Add("query", q)
	vals.Add("graph", "public")

	body := strings.NewReader(vals.Encode())

	req, err := http.NewRequest("POST", synbiohubURL, body)
	if err != nil {
		log.Fatal("couldn't prepare request: ", err)
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("couldn't make request: ", err)
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("couldn't read xml: ", err)
	}

	return bytes
}

func process(seqs []sequence) {

}
