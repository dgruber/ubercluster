ubercluster
===========

Simple multi-clustering tool based on DRMAA2 (e.g. for Univa Grid Engine).

## Compilation

Make sure you have the DRMAA2 Go binding (adding godep later...).
Make sure you have a cluster scheduler supporting DRMAA2 C API (like Univa Grid Engine)
installed.

Go to cmd/d2proxy
 
    $ source path/to/your/GE/installation
    $ ./build


Go to cmd/d2stat

    $ source path/to/your/GE/installation
    $ ./build

## Example usage

### Start your proxy - one per cluster:

    $ source path/to/your/GE/installation

    $ d2proxy &
    
or for listening on port 8282
    
    $ d2proxy -port=":8282" &

### Test the proxies by opening the address in the webbrowser.

Example:

    $ firefox http://localhost:8888/monitoring?jobs=all

### Update config.json 

The config.json file in *d2stat* directory needs to point to your cluster proxies. The **default** entry is the cluster/proxy which is used when no other is specified as parameter of *d2stat*.

### Examples

#### List all jobs of your default cluster

    $ d2stat -s=all

#### List all running jobs of cluster "cluster1" (from config)

    $ d2stat -c=cluster1 -s=r

    job_number:		3000000003
    state:			Running
    submission_time:	2014-12-06 18:02:59 +0100 CET
    dispatch_time:		2014-12-06 18:03:00 +0100 CET
    finish_time:		-
    owner:			daniel
    slots:			1
    allocated_machines:	u1010
    exit_status:		-1

    job_number:		3000000004
    state:			Running
    submission_time:	2014-12-06 18:03:01 +0100 CET
    dispatch_time:		2014-12-06 18:03:10 +0100 CET
    finish_time:		-
    owner:			daniel
    slots:			1
    allocated_machines:	u1010
    exit_status:		-1

#### Let a simple process running in cluster "cluster1"

    $ d2stat -c=cluster1 -submit=sleep -name=MySleeperJob -arg=77 -queue=all.q

#### ..and now in the default cluster:

    $ d2stat -submit=sleep -name=MySleeperJob -arg=77 -queue=all.q

#### List all hosts of default cluster:

    $ d2stat -m=all
    
    HOSTNAME ARCH NSOC NCOR NTHR LOAD MEMTOT SWAPTO
    u1010 x64 1 4 4 0.080000 504184 911731
    ...
