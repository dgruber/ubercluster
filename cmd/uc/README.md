uc
==

A simple tool to interact with ubercluster proxies.

The entire tool is written in pure Go and therefore can be compiled for all platforms 
supported by the **Go** compiler.

The **uc** tool communicates with compute clusters and compute backends through
JSON encoded [DRMAA2](http://www.drmaa.org) like data-structures based on the [Go DRMAA2 port](https://github.com/dgruber/drmaa2). DRMAA2 is an open standard by the [Open Grid Forum](http://www.ogf.org).

_uc_ requires a config.json file for configuring the endpoints of the proxies for the clusters.

_uc_ is used for starting and monitoring jobs which are executed remotely.

## Compilation and Installation

    go install


