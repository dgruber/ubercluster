/*
   Copyright, 2014, Daniel Gruber, info@gridengine.eu
*/

/*
The d2proxy tool is a proxy server which serves compute cluster
information data based on open DRMAA2 (see http://www.ogf.org)
standard. The implementation is based on the library which comes
with Univa Grid Engine.

Usage (on Linux):
   - Be sure Univa Grid Engine (>= 8.2.0) is in path (i.e. "qstat" works)
   - Set the LD_LIBRARY_PATH to the C DRMAA2 library
   - Start the d2proy in background
   - Check with browser if it is working (http://localhost:8888/monitoring?jobs=all)
   - Use other clients (like d2stat) to access it from any host in the network

    source path/to/gridengine/default/common/settings.sh
    export LD_LIBRARY_PATH=$SGE_ROOT/lib/lx-amd64
    ./d2proxy &
*/

package main
