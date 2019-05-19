# go-sprintfjs

A go port of https://github.com/alexei/sprintf.js

## Installation

```bash
go get -u github.com/crazytyper/go-sprintfjs
```

## Usage

```go
import "github.com/crazytyper/go-sprintfjs"

formatted, err := sprintfjs.Format(
    `Hello %(who)s!`,
    map[string]interface{}{
        "who":"world",
    },
);
```

## License

This package uses BSD 3-Clause "New" or "Revised" license.
See file [LICENSE](LICENSE) for details.

This package is based on [sprintf-js](https://github.com/alexei/sprintf.js):
Copyright (c) 2007-present, Alexandru Mărășteanu <hello@alexei.ro>
All rights reserved.
