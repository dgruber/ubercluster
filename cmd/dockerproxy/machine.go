package main

import (
	"github.com/dgruber/ubercluster/pkg/types"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"os"
)

func LocalhostToMachine() types.Machine {
	v, _ := mem.VirtualMemory()
	l, _ := load.Avg()
	hostname, _ := os.Hostname()

	var osVersion types.Version
	osVersion.Major = "0"
	osVersion.Minor = "0"

	return types.Machine{
		Name:           hostname,
		Available:      true,
		Sockets:        1,
		CoresPerSocket: 1,
		ThreadsPerCore: 1,
		Load:           l.Load1,
		PhysicalMemory: int64(v.Total),
		VirtualMemory:  int64(v.Total),
		Architecture:   types.X64,
		OS:             types.OtherOS,
		OSVersion:      osVersion,
	}
}
