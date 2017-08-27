# Docker Proxy

This is a proxy which creates Docker containers for task execution.

## Compilation

Compilation: 

    $ go build

Installation of the binary:

    $ go install

## Usage

### Starting Proxy

The proxy needs to be started. For that all generic parameters for the proxy are availabe:

    $ dockerproxy --help
    usage: dockerproxy [<flags>]
    
	A proxy server for Docker

	Flags:
	  --help               Show help.
	  --verbose            Enables enhanced logging for debugging.
	  --port=":8080"       Sets address and port on which proxy is listening.
	  --certFile=CERTFILE  Path to certification file for secure connections (TLS).
	  --keyFile=KEYFILE    Path to key file for secure connections (TLS).
	  --otp=OTP            One time password settings ("yubikey") or a fixed shared secret.
	  --yubiID=YUBIID      Yubi client ID if otp is set to yubikey.
	  --yubiSecret=YUBISECRET
	                       Yubi secret key if otp is set to yubikey
	  --yubiAllowedIds=YUBIALLOWEDIDS
	                       A list of IDs of yubikeys which are accepted as source for OTPs.

You might want to create a service which starts the proxy, or use monit, 
nohup, or just start the _dockerproxy_ in the commandline for testing.

Example:

    $ dockerproxy --otp "mysupersecretkey"

### Updating Config File

Update or create a ```config.json``` file so that ```uc``` can find the proxy. The config file
can be in the directory where you start ```uc``` or in the home directory ($HOME/.ubercluster/config.json)
or in /etc/ubercluster/config.json.

Example of a config file containing only one proxy (default) which listens on localhost port 8080:

    {"Cluster":[{"Name":"default","Address":"http://localhost:8080/","ProtocolVersion":"v1"}]}

Check if the config file is found by ```uc``` and the configuration is displayed.

    $ uc config list

### Creating Containers

With the ```uc``` commmand line tools containers can be created:

    $ uc --otp "supersecret" run --arg 120 --category ubuntu:latest /bin/sleep
    Jobid:  65eaddd872bc50515c8d1939147e1e42de439e28c3dc7f326e0d427ca8bff136
    Cluster:  default

    
