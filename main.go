package main

import (
	"flag"
	"fmt"
	"web/internal/config"
	"web/internal/handler"
	"web/internal/middleware"
	"web/internal/svc"

	core "github.com/ynhuu/gin-core"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	Version     = "develop"
	mVer        = flag.Bool("v", false, "version")
	mHost       = flag.String("l", "127.0.0.1:8000", "listen addr")
	mConfigFile = flag.String("f", "conf/conf.yml", "the config file")
)

func main() {
	flag.Parse()

	if *mVer {
		fmt.Println(Version)
		return
	}

	c := config.MustLoad(*mConfigFile)
	ctx := svc.NewServiceContext(c)
	defer ctx.Close()

	server := core.MustNewServer(*mHost)
	server.Use(middleware.NewCors(), middleware.NewPrometheus())
	server.GET(c.Metrics, gin.WrapH(promhttp.Handler()))

	handler.RegisterHandlers(server, ctx)

	zap.S().Info("Starting server at ", *mHost)
	server.Start()
}
