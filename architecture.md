# Architecture

Because things are getting complicated?

## Backing DB

Virtuoso

## DB Cache

(For testing and not knocking over Virtuoso if we need to rebuild.)

 - Filesystem (local or Google Storage?)

## De-Dupe Layer

Provide a layer that tracks duplicated sequences.

 - **Redis**

   Example schema: ZSet/List: `"seqs" -> ZSet/List<hash(seq)>`, sets for uris keyed by hash(seq).
   
   Indexers pull ranges from `"seqs"`, fetch corresponding sequences, push both to BLASTDB.
   
   Simple, fast because Redis, slow because indirect lookups, persistent or not depending on what's actually better here.

 - **SQLite** or **Postgres**

   Example schema: tables: components: `(hash, uri, created_at)`, sequences: `(hash, seq, created_at)`. 
   
   Indexers `SELECT (hash, uri) FROM sequences ORDER BY created_at LIMIT 100 OFFEST <POS>` to get batches, feed to BLASTDB.

   Queryservers, get results from BLASTDB, for each returned hash `SELECT uri FROM components WHERE hash == '<hash>'`.

   Advantages:

    - Could also use as DB cache!
    - If postgres, supports multiple Queryservers
    - Can do fun things with complex SQL queries

   Disadvantages:
    
    - Indirect multistep lookups (TODO: could be avoided)
    - SQL is harder to administer

 - **Flat File**

   Probably couldn't do dedup without weird reimplement-a-database-but-with-files shenanigans.

 - **Google Datastore**

   Advantages:

    - No management
    - Autoscaling for freeee
   
   Disadvantages:

    - Adds organizational dependency
    - Can't throw it in a docker and hand it out to synbiohub people
    - Less conducive to throwing it out and starting over