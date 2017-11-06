# Project Update #1

Ethan Ransom

Sunday November 5th, 2017

[View Project GitHub](https://github.com/schnauzer/synbioblast)

Screenshot: 

![screenshot](https://github.com/schnauzer/synbioblast/raw/master/screenshot_2017-11-05.png "Screenshot as of November 5th, 2017")


## Progress

* **Attempted to build BLAST from source**, but I ran into a compilation error (this 
was supposed to be the stable release) and had to revert to downloading the official 
linux binaries instead. Perhaps not the ideal solution from a security standpoint. 
Possibly worth looking into in the future. (Added the `blastn` and `makeblastdb` binaries to the 
repo. These programs run queries and build custom databases, respectively.)

* **Downloaded the sample BLAST database “16SMicrobial”** to run sample queries on. This 
database stores the 16S ribosomal RNA sequence for a large number of Bacteria and 
Archaea. (This sequence rarely evolves due to its importance and is often used for 
determining prokaryote phylogenies as a result.) [View the database.](https://github.com/schnauzer/synbioblast/tree/master/16SMicrobial)

 * **Wrote a simple web server that displays a form** for users to construct their queries. 
Upon form submission, the server spawns a blastn child process to perform the query 
and displays the results to the user. [View the webserver.](https://github.com/schnauzer/synbioblast/blob/master/synbioblast.go)

## Next Steps

Write some sort of long-running process that queries the Virtuoso database to generate a stream of sequences to be fed into the `makeblastdb` tool. 

Deploy the code.