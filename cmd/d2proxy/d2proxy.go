/*
   Copyright 2014 Daniel Gruber, Univa

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"errors"
	"github.com/dgruber/drmaa2"
	"github.com/dgruber/ubercluster"
	"gopkg.in/alecthomas/kingpin.v1"
	"io/ioutil"
	"log"
	"os"
)

var verbose bool = false
var JobSessionName = "ubercluster"

func init() {
	if verbose == false {
		log.SetOutput(ioutil.Discard)
	}
}

var (
	app        = kingpin.New("d2proxy", "A proxy server for DRMAA2 compatible cluster schedulers (like Univa Grid Engine).")
	cliVerbose = app.Flag("verbose", "Enables enhanced logging for debugging.").Bool()
	cliPort    = app.Flag("port", "Sets address and port on which proxy is listening.").Default(":8888").String()
	certFile   = app.Flag("certFile", "Path to certification file for secure connections (TLS).").Default("").String()
	keyFile    = app.Flag("keyFile", "Path to key file for secure connections (TLS).").Default("").String()
	otp        = app.Flag("otp", "One time password settings (currently \"yubikey\" or otherwise a fixed secret)").Default("").String()
)

type drmaa2proxy struct {
	ms *drmaa2.MonitoringSession
	js *drmaa2.JobSession
}

// implement neccessary methods to fulfill the ProxyImplemeter interface

func (d2p *drmaa2proxy) GetJobInfosByFilter(filtered bool, filter ubercluster.JobInfo) []ubercluster.JobInfo {
	var f *drmaa2.JobInfo
	if filtered {
		convertedFilter := ConvertUCJobInfo(filter)
		f = &convertedFilter
	}
	if ji, err := d2p.ms.GetAllJobs(f); err != nil {
		log.Println("Error during GetAllJobs(): ", err)
		return nil
	} else {
		jis := make([]ubercluster.JobInfo, 0, 0)
		for _, j := range ji {
			jobinfo, _ := j.GetJobInfo()
			d2ji := ConvertD2JobInfo(*jobinfo)
			jis = append(jis, d2ji)
		}
		return jis
	}
}

func (d2p *drmaa2proxy) GetJobInfo(jobid string) *ubercluster.JobInfo {
	filter := drmaa2.CreateJobInfo()
	filter.Id = jobid
	if ji, err := d2p.ms.GetAllJobs(&filter); err == nil {
		if len(ji) == 1 {
			jobinfo, _ := ji[0].GetJobInfo()
			ucJobInfo := ConvertD2JobInfo(*jobinfo)
			return &ucJobInfo
		}
	}
	return nil
}

func (d2p *drmaa2proxy) GetAllMachines(machines []string) ([]ubercluster.Machine, error) {
	if m, err := d2p.ms.GetAllMachines(machines); err != nil {
		return nil, err
	} else {
		return ConvertD2Machine(m), nil
	}
}

func (d2p *drmaa2proxy) GetAllQueues(queues []string) ([]ubercluster.Queue, error) {
	if q, err := d2p.ms.GetAllQueues(queues); err != nil {
		return nil, err
	} else {
		return ConvertD2Queue(q), nil
	}
}

func (d2p *drmaa2proxy) GetAllCategories() ([]string, error) {
	return d2p.js.GetJobCategories()
}

func (d2p *drmaa2proxy) DRMSVersion() string {
	var sm drmaa2.SessionManager
	if version, err := sm.GetDrmsVersion(); err == nil {
		return version.String()
	}
	return ""
}

func (d2p *drmaa2proxy) DRMSName() string {
	var sm drmaa2.SessionManager
	if name, err := sm.GetDrmsName(); err == nil {
		return name
	}
	return ""
}

func (d2p *drmaa2proxy) RunJob(template ubercluster.JobTemplate) (string, error) {
	if job, err := d2p.js.RunJob(ConvertUCJobTemplate(template)); err != nil {
		return "", err
	} else {
		return job.GetId(), nil
	}
}

func (d2p *drmaa2proxy) JobOperation(jobsessionname, operation, jobid string) (string, error) {
	// The filter is missing in GetJobs() hence until this is
	// fixed in Go DRMAA2 we use a non-scaling method and do
	// filtering on our own.
	if jobs, err := d2p.js.GetJobs(); err != nil {
		log.Println("Error while DRMAA2 GetJobs()")
		return "", err
	} else {
		log.Println("Got following jobs in job session: ", jobs)
		for _, job := range jobs {
			log.Println("Job id: ", job.GetId())
			if job.GetId() == jobid || job.GetId() == jobid+".1" {
				switch operation {
				case "suspend":
					if err := job.Suspend(); err != nil {
						return "", err
					} else {
						return "success", nil
					}
				case "resume":
					if err := job.Resume(); err != nil {
						return "", err
					} else {
						return "success", nil
					}
				case "terminate":
					if err := job.Terminate(); err != nil {
						return "", err
					} else {
						return "success", nil
					}
				default:
					return "", errors.New("unsupported operation")
				}
			}
		}
	}
	return "", errors.New("job not found")
}

// implementation specific methods
func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *cliVerbose {
		log.SetOutput(os.Stdout)
	}

	// read-in config
	initializeD2Proxy()

	// Open MonitoringSession and create a JobSession with the given name
	var proxy drmaa2proxy
	proxy.initializeDRMAA2(JobSessionName)
	defer proxy.js.Close()
	defer proxy.ms.CloseMonitoringSession()

	ubercluster.ProxyListenAndServe(*cliPort, *certFile, *keyFile, *otp, &proxy)
}
