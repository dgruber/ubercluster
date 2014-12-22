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

    $ firefox http://localhost:8888/v1/monitoring?jobs=all

### Update config.json 

The *config.json* file in **uc** directory needs to point to your cluster proxies. The *default* entry is the cluster/proxy which is used when no other is specified as parameter of **uc**.

### Examples

#### List all jobs of your default cluster

    $ uc show jobstate all

#### List all running jobs of cluster "cluster1" (from config)

    $ uc show jobstate r

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

#### Let a simple process run in default cluster

    $ uc run --arg=123 /bin/sleep

#### ...and now in the "cluster1" cluster, adding a job name and selecting a queue (default is "all.q"):

    $ uc --cluster=cluster1 run --queue=all.q --name=MyName --arg=123 /bin/sleep

#### List all hosts of default cluster:

    $ uc show machine
    
    HOSTNAME ARCH NSOC NCOR NTHR LOAD MEMTOT SWAPTO
    u1010 x64 1 4 4 0.080000 504184 911731
    ...

#### Get full command description...

    $ uc --help

```sh
usage: uc [<flags>] <command> [<flags>] [<args> ...]

A tool which can interact with multiple compute clusters.

Flags:
 --help               Show help.
 --verbose            Enables enhanced logging for debugging.
 --cluster="default"  Cluster name to interact with.

Commands:
  help [<command>]
    Show help for a command.

  show job [<id>]
    Information about a particular job.

  show jobstate [<state>]
    All jobs in a specific state (r/p/all).

  show machine [<name>]
    Information about compute hosts.

  show queue [<name>]
    Information about queues.

  run [<flags>] <command>
    Submits an application to a cluster.

  config list
    Lists all configured cluster proxies.
```
