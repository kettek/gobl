# gobl - Go Build
[![Go Reference](https://pkg.go.dev/badge/github.com/kettek/gobl.svg)](https://pkg.go.dev/github.com/kettek/gobl)

gobl is an experimental build system that uses go to define and run build tasks. Although it is actively used to develop other programs, it is by no means stable and may have breaking API changes if any redesigns are required. However, syntax _should_ largely remain the same, as I find its current syntax to be simple to use, which is one of the primary goals of it.

## Quickstart
In your project root directory, simply create a file called `gobl.go` and fill its contents with the following:

```go
package main

import (
	. "github.com/kettek/gobl"
)

func main() {
	Task("listFiles").
		Exec("ls")

	Go()
}

```

You will likely have to get the gobl dependency by issuing `go get -d github.com/kettek/gobl`.

A more complex example that would allow automatic rebuilding + running, would be:

```go
package main

import (
	. "github.com/kettek/gobl"
)

func main() {
	Task("build").
		Exec("go", "build", "./cmd/mycmd")

	Task("run").
		Exec("mycmd")

	Task("watch").
		Watch("./pkg/*/*.go").
		Signaler(SigQuit).
		Run("build").
		Run("run")

	Go()
}
```

## Task Steps
For a complete rundown of available steps, see the [godoc task reference](https://pkg.go.dev/github.com/kettek/gobl@v0.1.0/pkg/task).

## Visual Studio Code Integration
There is a task provider extension for VSCode that allows using gobl tasks as VSCode tasks. It is available as [Gobl Task Provider](https://marketplace.visualstudio.com/items?itemName=kettek.gobl-task-provider) and has a GitHub repository [here](https://github.com/kettek/vscode-gobl-task-provider).
