# go-srv
Interface for implementing servers in go with graceful shutdown on OS signals.

## Example server implementation
gin-gonic/zerolog/gorm
```go
import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/rs/zerolog"

	server "github.com/balazshorvath/go-srv"
)

const defaultConfig = "config.yaml"

type httpServer struct {
	server.BasicHttpServer

	logger *zerolog.Logger
	config *config
	db     *gorm.DB
}

func New(ctx context.Context, group *sync.WaitGroup) server.Server {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = defaultConfig
	}
	config := parseConfig(path)
	return &httpServer{
		BasicHttpServer: server.BasicHttpServer{
			BasicServer: server.BasicServer{
				Ctx:   ctx,
				Group: group,
			},
		},
		config: config,
	}
}

func (h *httpServer) Init() {
	router := gin.Default()
	h.Srv = &http.Server{
		Addr:    fmt.Sprintf("%s:%s", h.config.server.Host, h.config.server.Port),
		Handler: router,
	}
	// Init dependencies
	h.initDatabase(&h.config.database)
	// Setup http
	router.GET("/api/path", handlePath(h.db))
}

func handlePath() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"hello": "there",
		})
	}
}
```
