# Usage

```go
package main
import "github.com/tanopwan/server-template/server"
import "os"

func main() {
    s := server.NewInstance("example-app", "1", nil)
    err := s.Start()
    if err != nil {
        os.Exit(1)
    }
}
```