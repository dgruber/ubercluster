uc
==

A simple tool to interact with multiple compute clusters and compute backends. 

The entired tool is written in pure Go and therefore can compiled for all platforms 
supported by **Go** compilers.

The **uc** tool communicates with compute clusters and compute backends through
JSON encoded [DRMAA2](http://www.drmaa.org) like data-structures based on the [Go DRMAA2 port](https://github.com/dgruber/drmaa2). DRMAA2 is an open standard by the [Open Grid Forum](http://www.ogf.org).

## Compilation and Installation

    godep go install


