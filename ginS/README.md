# Gin Default Server

This is API experiment for Gin.

```go
package main

import (
	"github.com/numbergroup/gin-light"
	"github.com/numbergroup/gin-light/ginS"
)

func main() {
	ginS.GET("/", func(c *gin.Context) { c.String(200, "Hello World") })
	ginS.Run()
}
```
