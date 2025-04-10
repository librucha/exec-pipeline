# exec-pipeline

`exec-pipeline` is a Go module designed to simplify the execution of complex workflows or pipelines. It provides utilities to chain and manage the execution of tasks in a structured and efficient manner.

## Installation

To install this module, run:

```bash
go get github.com/librucha/exec-pipeline
```

## Usage

Import the module into your project:

```go
import "github.com/librucha/exec-pipeline"
```

Here's an example of how to use it:

```go
package main

import (
	"github.com/librucha/exec-pipeline"
	"fmt"
)

func main() {
	pipeline := execpipeline.NewPipeline()

	err := pipeline.AddStep(func() error {
		fmt.Println("Step 1 executed")
		return nil
	}).AddStep(func() error {
		fmt.Println("Step 2 executed")
		return nil
	}).Run()

	if err != nil {
		fmt.Printf("Pipeline execution failed: %v\n", err)
	}
}
```

## Features

- Easy-to-use interfaces for defining and running pipelines.
- Error handling for each step in the pipeline.
- Extensible for different execution patterns.

## Contributing

Contributions are welcome! Please submit any issues, feature requests, or pull requests through the GitHub repository.

## License

This project is licensed under the [MIT License](LICENSE).