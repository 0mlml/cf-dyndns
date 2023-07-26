# Cloudflare Dynamic DNS Update Tool

This tool allows you to update a DNS record on Cloudflare with your current IP address, effectively providing dynamic DNS functionality. It's useful if you have a domain on Cloudflare and your home IP address changes often.

## Features

- Supports flags for setting options from the command line.
- Can read options from a configuration file.
- Debug logging and optional suppression of output.

## Usage

Because it is a [Go](https://go.dev/dl) program, you can run it directly with the `go run` command:
```
go run main.go parser.go 
```

You can also build it and run the executable:
```
go build
./cf-dyndns
```

## Options

- -optfile: Path to the options file. Default is "dyndns.cfg".
- -quiet: Suppresses all output except errors when set. Default is false.
- -cfapikey: Your Cloudflare API key.
- -cfemail: The email address associated with your Cloudflare account.
- -domain: The domain to update.
- -record: The record to update.
- -ipapi: Your preferred IP API URL. It must return a plain text IP address. Default is "https://api.ipify.org".

## Configuration File Format

The configuration file uses a simple format with sections for different types of options (bool, string, int). For example:

```
[bool]
quiet=false 
[string]
cfapikey=YOUR_API_KEY
cfemail=YOUR_EMAIL_ADDRESS
domain=YOUR_DOMAIN
record=YOUR_RECORD
ipapi=https://api.ipify.org/
```
