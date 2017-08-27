# Proxy for Cloud Foundry Tasks

## Running the Proxy

The proxy can be pushed as a Cloud Foundry application ('cf push'). In order to
create Cloud Foundry tasks it needs access to the CF API. API address, user,
password have to be configured as environment variables.

Example _manifest.yml_

    ---
	applications:
	  - name: cf-tasks
	    command: cf-tasks --otp=mysupersecretpassword
	    env:
	      GOVERSION: go1.8
	      GOPACKAGENAME: github.com/dgruber/uc/proxy/cf-tasks
	      CF_TARGET: https://api.run.pivotal.io
	      NAME: user@xyz.io
	      PASSWORD: password

In order to protect access to the proxy you can use as start command in the manifest
something like _cf-tasks --verbose --otp=yubikey --yubiID=clientIDFromYubico --yubiSecret=generatedSecretFromYubico --yubiAllowedIds=first12digitsOfYourYubikeys_ 

## Running Tasks

Cloud Foundry tasks can be created by using 'uc'.

    uc --otp=yubikey run --arg="120" --name=mytask --category=mycfappimage myCommand

    <press yubikey>

    cf tasks mycfappimage

