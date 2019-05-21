# Google Ads Doctor - oauthdoctor
This program will verify your Google Ads client OAuth2 environment and report
anomalies. Where possible, it will guide you through correcting the problems.
This assumes that you have already completed the steps in your client library
[README.md](https://developers.google.com/google-ads/api/docs/first-call/get-client-lib)
file.

The program vets your client library configuration file, verifies connectivity
to the Google Ads host, and executes either the installed application or web
flow OAuth2 process.

If it discovers errors it will attempt to guide you to correct them. For example,
if your refresh token is invalid, it will walk you through the process of
creating a valid token which is then written to your client library configuration
file.

# Downloads
If you are building from source, follow the
[source installation instructions](#source).

If you want to use a binary, fetch the repository and

```
cd <path>/google-ads-doctor/oauthdoctor/bin
```

There are three directories, one each for Darwin (OSX), Linux, and Windows.
You can run the binary in place or copy the file to any location on your system.

```
bin
├── darwin
│   ├── 386
│   └── amd64
├── linux
│   ├── 386
│   └── amd64
└── windows
    ├── 386
    └── amd64
```


#  <a name="running"></a> Running the Program

```
oauthdoctor -help
```

This displays the available command line options. Two of them, -language and
-oauthtype are required. So for an installation using Python and the installed
application OAuth flow, you would type:

```
oauthdoctor -language python -oauthtype installed_app
```

If your configuration file is not in your home directory (the default location),
then you will want to specify the location with the --configpath option.

```
oauthdoctor -language python -oauthtype installed_app -configpath /my/path
```

-sysinfo prints the system information to stdout. This is
primarily of use if you need to send the output of the program when contacting
support.

-verbose is for debugging. It will print the complete JSON responses.

-hidePII is for when you are sending the output to someone and you want to
mask sensitive information like your Client Secret.

# Sending output to someone else

If you want to send the output to someone else to assist you with a problem,
consider using the --hidePII option. This will mask sensitive information such
as your client secret and refresh token.


#  <a name="source"></a> Install from Source

Clone the repository outside of your GOPATH. If you must clone within your GOPATH
then set the environment variable `GO111MODULE=off`.

Download and install the Go programming language runtime
[latest version](https://golang.org/dl/) and follow the installation instructions.

Google Ads Doctor requires Go version 1.11 or greater because we are using
[Go modules](https://github.com/golang/go/wiki/Modules) for dependency management.

Once you have verified your Go installation, in a terminal, change to the
directory where you cloned the repository and

```
cd oauthdoctor
```

At the command line, type:

```
go build ./oauthdoctor.go
```

This produces a binary called oauthdoctor. From here, follow the the
instructions in [Running the Program](#running)


# Where do I submit bug reports or feature requests?

If you have issues directly related to `oauthdoctor`, use the
[issue tracker](https://github.com/googleads/google-ads-doctor/issues)

For issues with the Google Ads API, visit [Google Ads API Support](https://developers.google.com/google-ads/api/support)

# Authors

Authors:

 - Bob Hancock
 - Poki Chui
