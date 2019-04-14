# Google Ads Doctor - oauthdoctor
This program will verify your Google Ads client OAuth2 environment and report
anomalies. Where possible, it will guide you through correcting the problems.
This assumes that you have already completed the steps in your client library
[README.md](https://developers.google.com/google-ads/api/docs/first-call/get-client-lib)
file.

The program vets your client library configuration file, verifies connectivity
to the Google Ads host, and simulates either the installed application or web
flow OAuth2 process.

If it discovers errors it will attempt to guide you to correct them. For example,
if your refresh token is invalid, it will walk you through the process of
creating a valid token which is then written to your client library configuration
file.

# Requirements
Go minimum version 1.11. We require 1.11 or greater because we are using
[Go modules](https://github.com/golang/go/wiki/Modules) for dependency management.

# Setup
To run the program, you need to install the Go programming language. Download
the [latest version](https://golang.org/dl/) and follow the installation
instructions.

Clone the code into a directory that is not in your GOPATH.

# Running the program
Once you have verified your Go installation, in a terminal, change to the
directory where you cloned the repository. Change to the directory
oauthdoctor.

At the command line, type:

```
go build ./oauthdoctor.go
```

This will produce a binary called oauthdoctor.

```
./oauthdoctor -help
```

This will display the available command line options. Two of them, -language and
-oauthtype are required. So for an installation using Python and the installed
application OAuth flow, you would type:

```
go ./oauthdoctor.go -language python -oauthtype installed_app
```

If your configuration file is not in your home directory (the default location),
then you will want to specify the location with the --configpath option.

```
go run ./oauthdoctor.go -language python -oauthtype installed_app -configpath /my/path
```

-sysinfo prints the system information to stdout. This is
primarily of use if you need to send the output of the program when contacting
support.

-verbose is for debugging. It will print the complete JSON response.

-hidePII is for when you are sending the output to someone and you want to
mask sensitive information like your client secret.

# Sending output to someone else

If you want to send the output to someone else to assist you with a problem,
consider using the --hidePII option. This will mask sensitive information such
as your client secret and refresh token.

# Where do I submit bug reports or feature requests?

If you have issues directly related to `oauthdoctor`, use the
[issue tracker](https://github.com/googleads/googleads-doctor/issues).

For issues with the Google Ads API, visit [Google Ads API Support](https://developers.google.com/google-ads/api/support).

# Authors

Authors:

 - Bob Hancock
 - Poki Chui
