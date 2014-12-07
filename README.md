clusterstatus
=============

Simple multiclustering tools based on DRMAA2 (e.g. for Univa Grid Engine).

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

1. Start your proxy - one per cluster:

    $ source path/to/your/GE/installation
    $ d2proxy &
    
or for listening on port 8282
    
    $ d2proxy -port=":8282" &

2. Test the proxies by opening the address in the webbrowser. Example:

    $ firefox http://localhost:8888/monitoring?jobs=all

3. Update config.json file in d2stat directory so that it points to your clusters. The default entry is the cluster which is used when no one is specified.

4. List all jobs of your default cluster:

    $ ./d2stat -s=all

5. List all running jobs of cluster "cluster1" (from config):

    $ ./d2stat -c=cluster1 -s=r

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

6. Let a simple process running in cluster "cluster1":

    $ ./d2stat -c=cluster1 -submit=sleep -name=MySleeperJob -arg=77 -queue=all.q

7. And now in the default cluster:

    $ ./d2stat -submit=sleep -name=MySleeperJob -arg=77 -queue=all.q

8. List all hosts of default cluster:

    $ ./d2stat -m=all
    
    HOSTNAME ARCH NSOC NCOR NTHR LOAD MEMTOT SWAPTO
    u1010 x64 1 4 4 0.080000 504184 911731
    ...