# l'oGGo: Rich Terminal User Interface Logging App
![release](https://img.shields.io/github/v/release/aurc/loggo)
![release workflow](https://github.com/jimbertools/loggo/actions/workflows/go.yml/badge.svg)
![license](https://img.shields.io/github/license/aurc/loggo)
![go_version](https://img.shields.io/github/go-mod/go-version/aurc/loggo)
![downloads](https://img.shields.io/github/downloads/aurc/loggo/total)
## Introduction

*For the impatient, go to [Getting Started](#getting-started) or [Using as a Package](#using-as-a-package)*
<p align="center">
<img src="img/loggo_sm.png">
</p>

l'oGGo or Log & Go is a rich Terminal User Interface app written in [golang](https://go.dev/) that harness the
power of your terminal to digest log streams based on JSON based logs.

This can be used against applications running locally, on a Kubernetes cluster (see [K8S Cheatsheet](#k8s)), GCP
Stack Driver ([Google Logs](https://cloud.google.com/logging), see [GCP-Stream Command](#gcp-stream-command)) and many others.

<img src="img/compare.png">
<table>
<tr>
<td>
<p>Without l`oGGo</p>
<img src="img/mov/term.gif">
</td>
<td>
<p>With l`oGGo</p>
<img src="img/mov/loggo.gif">
</td>
</tr>
</table>

Loggo App leveraged [tview](https://github.com/rivo/tview/) and [tcell](https://github.com/gdamore/tcell) projects for rich Terminal User
Interface (TUI).

## Getting Started

### macOS/Linux with Homebrew/Linuxbrew:
The easiest way is to utilise [Homebrew](https://brew.sh/) **(macOS)** or [LinuxBrew](https://docs.brew.sh/Homebrew-on-Linux) **(Linux)**
package management system. Once installed simply issue the following command:

````
brew tap aurc/loggo
brew install aurc/loggo/loggo
````

To update:
````
brew upgrade aurc/loggo/loggo
````

### All Systems

### Install with Go
Assuming you have [go](https://go.dev/) installed in your system (best choice if not homebrew) and bin packages
are in your `PATH`, just run:

````
go install github.com/jimbertools/loggo@latest
````

### Build from Source:
Including **macOS**, build from source.
Pre-Reqs:
- [Golang](https://go.dev/) v1.8+
- Checkout the project from [https://github.com/jimbertools/loggo/](https://github.com/jimbertools/loggo/)

````
go build -o loggo
````
> Install the `loggo` binary in your system's binary bucket or add `loggo` binary to your
path.

### Download Pre-Compiled binary

This option might be suitable for you if you don't have Home/Linux-Brew or golang
readly available for you. In this case you can download a pre-compiled exectuable binary
for the following platforms:

Latest Release [pre-build binaries here](https://github.com/jimbertools/loggo/releases/latest)!

- **Windows 11**:

   Download a `tar.gz` file from [pre-build binaries](https://github.com/jimbertools/loggo/releases/latest) that match your platform:
   - Most of PCs will be `amd64`, so file format will be like `loggo_x.y.z_windows_amd64.tar.gz`. Note: *If you own a Surface Pro, or
virtualised windows on a Mac M series, you'll probably want to grab the `arm64` version.*
   - Open a terminal window (`cmd.exe`) and issue the following command in the folder you downloaded the file:
     ````
     tar -xvzf loggo_x.y.z_windows_amd64.tar.gz
     ````
   - A file `loggo.exe` will be extracted to the current folder.
   - You can place this file in a directory that's on your %PATH% so the command `loggo` is accessible from anywhere.
   -
- **Linux** (both arm64 and amd64 plat):
  - ubuntu:  `tar.gz` file from releases (e.g. loggo_vx.y.z_linux_<plat>.tar.gz)
  - debian:  `*.deb` file from releases
  - redhat:  `*.rpm` file from releases
  - alpine:  `*.apk` file from releases

## Using l'oGGo

Loggo can be used to stream parsed logs from a persisted file and from a
piped input and also provides a tool for creating log templates.

### Some Features
- Local Log filtering/search
  - Main log stream remains unaffected regardless of the source (gcp, pipe, file, etc...)
  - Display only log entries that match search/filter criteria
  - Convenient key finder and operators for filter expression crafting
  ![](img/loggo_filter.png)
- Drill down onto each log entry
  ![](img/log_entry.png)
- Copy Log-Entry to Clipboard
  - Note: Linux requires X11 dev package. For instance, install `libx11-dev` or `xorg-dev` or `libX11-devel` to access X window system.
    ![](img/copy_clipboard.png)
- Navigate Left-Right-Up-Down on Large Grids
  - Select a Line
  - Use the arrow keys (`↓ ↑ ← →`)
    ![](img/mov/nav_right_left.gif)
- Select on screen text
  - Horizontally based selection (`Alt` + Mouse `Click/Drag`)
  - Block/Vertical based selection (`Cmd`+`Opt`+ Mouse `Click/Drag` - macOS)
  - Copy the selected text to clipboard (`Cmd`+`C` - macOS/`Ctrl`+`C` - other systems)
    ![](img/mov/selection.gif)
- Configure Rendering Templates:
  ![](img/render_template.png)
- Fine Tune how columns are displayed (Template):
  - Note that single Value Matches are REGEX expressions.
    ![](img/how_to_display.png)

### `help` Command

To gain fine grained insight of each `loggo` command params, use
the `help` command, e.g.:
````
loggo help
loggo help stream
loggo help template
loggo help gcp-stream
````

### `stream` Command

![](img/loggo_log.png)

**From File:**
````
loggo stream --file <my file>
````
*With Template:*
````
loggo stream --file <my file> --template <my template yaml>
````

**From Pipe:**
````
tail -f <my file> | loggo stream
````
Kubernetes example (See [K8S Cheatsheet](#k8s-cheatsheet))
````
kubectl logs -f -n <namespace> <pod> | loggo stream
````
*With Template:*
````
tail -f <my file> | loggo stream --template <my template yaml>
````

Note that you can pipe to anything that produces an output to the `stdin`.

### `gcp-stream` Command
l`oGGo natively supports GCP Logging but in order to use this feature, there are a few caveats:
- Your personal account has the required permissions to access the logging resources.


Note: `gcp-stream` **does not** support piped commands. If you want to use piped
commands (e.g. chaining K8S output) use the `stream` command instead.

Example:
````
loggo gcp-stream \
    --filter 'resource.labels.namespace_name="some-namespace" resource.labels.container_name="some-container"' \
    --project some-project-ID \
    --from 10m
````
Where:
````
Usage:
  loggo gcp-stream [flags]

Flags:
  -p, --project string       GCP Project ID (required)

  ------------------- Optional Below ------------------

  -f, --filter string        Standard GCP filters
      --force-auth           Only effective if combined with gcloud flag. Force re-authentication even
                             if you may have a valid authentication file.
  -d, --from string          Start streaming from:
                               Relative: Use format "1s", "1m", "1h" or "1d", where:
                                         digit followed by s, m, h, d as second, minute, hour, day.
                               Fixed:    Use date format as "yyyy-MM-ddH24:mm:ss", e.g. 2022-07-30T15:00:00
                               Now:      Use "tail" to start from now (default "tail")
      --gcloud-auth          Use the existing GCloud CLI infrastructure installed on your system for GCP
                             authentication. You must have gcloud CLI installed and configured. If this
                             flag is not passed, it uses l'oggo native connector.
  -h, --help                 help for gcp-stream
      --params-list          List saved gcp connection/filtering parameters for convenient reuse.
      --params-load string   Load the parameters for reuse. If any additional parameters are
                             provided, it overrides the loaded parameter with the one explicitly provided.
      --params-save string   Save the following parameters (if provided) for reuse:
                               Project:   The GCP Project ID
                               Template:  The rendering template to be applied.
                               From:      When to start streaming from.
                               Filter:    The GCP specific filter parameters.
  -t, --template string      Rendering Template
````

For convenience, you can build a list of frequently used command parameters/flags and reuse them without
having to rewrite lengthy list of parameters, for example:

````
loggo gcp-stream \
    --filter 'resource.labels.namespace_name="some-namespace" resource.labels.container_name="some-container"' \
    --project some-project-ID \
    --from 10m
    --template /tmp/myTemplate.yaml
    --params-save mySavedParams1
````

Then you simply issue:
````
loggo gcp-stream --params-load mySavedParams1
````

If you want to review all saved params buckets, issue the following command:

````
loggo gcp-stream --params-list
````

Additionally, you might want to overwrite some parameters. The example command uses `--from 10m`, and
say you want to `tail` instead:

````
loggo gcp-stream --params-load mySavedParams1 --from tail
````

Any additional parameter provided will overwrite the loaded params at runtime.

### `template` Command
The template command opens up the template editor without the
need to stream logs. This is convenient if you want to craft
templates prior using the loggo command.

**Blank Canvas:**

![](img/mov/template.gif)
````
loggo template
````

**Edit Existing Template:**
````
loggo template --file <my template yaml>
````

## K8S Cheatsheet

Combined logs of all pods of an application.
````
kubectl -n <some-namespace> logs -f deployment/<application-name> \
  --all-containers=true \
  --since=10m | loggo stream
````

Logs of a pod.
````
kubectl logs -f -n <some-namespace> <pod-name> | loggo stream
````

## Current Limitations

Most of the items listed here are slated for development in the near future,
prior the first release.
- Browse/Load new log templates on the fly.
- Create template with keys whose name contains `/` as it uses slashes to navigate to nested json branches.

## Feedback

## Using as a Package

Loggo can also be used as an importable package in your Go applications. This allows you to integrate Loggo's powerful log viewing and processing capabilities directly into your applications.

### Installation

```
go get github.com/jimbertools/loggo
```

### Basic Usage

```go
package main

import (
	"github.com/jimbertools/loggo/pkg"
)

func main() {
	// Create a new loggo app that reads from a file
	app := pkg.NewLoggoApp("path/to/logfile.log", "path/to/template.yaml")

	// Run the app
	app.Run()
}
```

### Using the Reader Directly

```go
package main

import (
	"fmt"

	"github.com/jimbertools/loggo/pkg"
)

func main() {
	// Create a reader that reads from stdin
	reader := pkg.NewReader("", nil)

	// Start streaming
	err := reader.StreamInto()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer reader.Close()

	// Process each line
	for line := range reader.ChanReader() {
		fmt.Printf("Received: %s\n", line)
	}
}
```

For more examples, see the `examples/` directory.

---

Please let us know your **thoughts**, **feature requests** and **bug reports**! Use the issues report
link here: https://github.com/jimbertools/loggo/issues
