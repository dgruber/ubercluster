package main

import ()

func (cp *CFProxy) listApps() ([]string, error) {
	apps, errApps := cp.client.ListApps()
	if errApps != nil {
		return nil, errApps
	}
	appNames := make([]string, 0, len(apps))
	for i := range apps {
		appNames = append(appNames, apps[i].Name)
	}
	return appNames, nil
}
