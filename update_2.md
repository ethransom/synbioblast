# Project Update #2

Ethan Ransom

Wednesday November 22nd, 2017

[View Project GitHub](https://github.com/schnauzer/synbioblast)

## Progress

Work since the last update has focused on the code responsible for scraping components
and associated sequences out of SynBioHub/Virtuoso and giving them to `./makeblastdb`,
the program responsible for creating a blast database readable by the blast algorithm.

Tasks accomplished:

* **Architected a log to store components in order of creation time.** My assumption
  was that the dataset would be large and slow to fetch from Virtuoso, so I considered
  designs for a local persistent store that would be optimized for the access patterns
  of the indexing code.

    Some options I considered were a flat file, postgresql, sqlite, and Google Datastore.

* **Explored sparql queries to fetch components from the database.** Some strategies explored:

   1. Sort the components by their creation time (oldest -> newest) and use `LIMIT` and `OFFSET`
   to fetch them in pages. If the returned list is less than the specified `LIMIT` then we 
   must have fetched all components currently in the database and the scraper should sleep
   for a while before trying again.

        ```sparql
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
        LIMIT 100 OFFSET 1000
        ```

  2. Pick a cutoff time less than the current time. Select all components older than that time
    and sort by uri to page through them. Periodically reset the cutoff time to check
    for new records.

        ```
        PREFIX dcterms: <http://purl.org/dc/terms/>
        PREFIX sbol: <http://sbols.org/v2#>
        PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>

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
                    ?uri dcterms:created ?created . FILTER (xsd:dateTime(?created) > "2010-03-01T00:00:00"^^xsd:dateTime) 
                } ORDER BY ?uri
            }
        }
        LIMIT 100 OFFSET 1000
        ```

 * It was at this point that I realized that Virtuoso was handling both queries surprisingly
  well, even with large `OFFSET`s. This means that a log to cache the values locally might not be necessary.

 * **Wrote the beginnings of the scraper program.** The scraper sends the sparql query to `https://synbiohub.org/sparql` (although the specific SynBioHub instance needs to be configurable) and parses the resulting XML to get a list of (URI, sequence, created) tuples. [View the scraper.](https://github.com/schnauzer/synbioblast/blob/master/slurper.go)

   It was at this point I found and [fixed a minor bug](https://github.com/SynBioHub/synbiohub/pull/526) in the 
   `/sparql` endpoint on SynBioHub.

 * **Realized that duplicate sequences were very common** due to the 
   many-to-many nature of SBOL components and sequences as well as a 
   few duplications due to human error. I discussed this with Zach
   and decided that making SynBioBLAST aware of these duplicates would 
   make for a better UX.

   At first I considered rewriting my Virtuoso query to group the components
   that share a sequence together. This list of uris would be turned into a comma-separated string and reported to `makeblastdb` as the name
   of the sequence. When the webserver was collating results from the
   `blastn` program it would separate the comma separated list and format
   the uris as a list of links.

   However, this would mean that we would need to refetch the list
   of uris for a component periodically--we couldn't fetch it once
   and then know that it wouldn't change out from under us. On the other hand I was discovering that rebuilding the blastdb from the 
   database every night or so wouldn't be as painful as I had first assumed, so I might revisit this option as it does allow the scraping program to be stateless, which simplifies deployment.

   For the time being, I'm going to use the original queries and
   have the scraping program keep state on which sequences it has
   already added to the index. It will also create a record for each
   sequence of the components that utilize it. When the results from `blastn` are being collated each sequence will be looked up to 
   generate the list of links.

 * **I chose [Redis](https://redis.io/) to store the state of the
   scraping program.** Redis is an in-memory key value data store
   that supports a number of useful data structures. In particular, 
   a large set could be used to store the list of seen sequences 
   (or the `sha1` hash of said sequences), and sets for every 
   sequence used to store the uris referencing that sequence.

   Redis has optional persistence, which could be used so that 
   we don't have to regenerate data after crashes, restarts,
   re-deploys, etc. But as discussed earlier these regenerations
   probably aren't too painful, so this could be left off for 
   the performance increase. (Redis is supposed to be Fast!)

   Rejected options included a flat file (too slow
   for looking up query results), sqlite/postgres (would require
   a complex schema for the operations required), Google BigTable (way too expensive, lol), and Google Datastore (learning curve
   and would introduce an organizational dependency for SynBioHub
   instances).

## Next Steps

 * Make the scraper add hashes of the sequences to a Redis set 
   and add components to a set keyed by the hash of their sequence.

   Write a loop that runs the query repeatedly, adding the number of fetched sequences to a redis variable. (That variable is used as 
   `OFFSET` of next query.) When less results are returned than the `LIMIT`, sleep for a while because we're caught up.

   Have the scraper write each sequence to one or more fasta files,
   naming each sequence with its hash.

 * Write code that periodically (nightly?) rebuilds the blast 
   database with `makeblastdb`. (Turns out `makeblastdb` can't add
   to an existing database, I was wrong on that account.) 

 * If querying is slow, look into sharding the fasta files into
   multiple blastdbs. (Based on their hash.) This would take 
   advantage of multiple CPUs on one server and/or multiple servers.

 * Modify the query server to look up the hash of the results
   and generate a list of links to components that reference that
   sequence.

 * Deploy the code.