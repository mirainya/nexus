package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/mirainya/nexus/internal/api"
	"github.com/mirainya/nexus/internal/llm"
	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/internal/processor"
	"github.com/mirainya/nexus/internal/service"
	"github.com/mirainya/nexus/internal/sse"
	"github.com/mirainya/nexus/internal/worker"
	"github.com/mirainya/nexus/pkg/cache"
	"github.com/mirainya/nexus/pkg/config"
	"github.com/mirainya/nexus/pkg/database"
	"github.com/mirainya/nexus/pkg/logger"
	"github.com/mirainya/nexus/pkg/vectordb"
)

func main() {
	mode := flag.String("mode", "all", "run mode: api, worker, all")
	flag.Parse()

	cwd, _ := os.Getwd()
	configPaths := []string{
		filepath.Join(cwd, "configs", "config.yaml"),
		"configs/config.yaml",
	}

	var configPath string
	for _, p := range configPaths {
		if _, err := os.Stat(p); err == nil {
			configPath = p
			break
		}
	}
	if configPath == "" {
		log.Fatalf("config file not found")
	}

	if err := config.Load(configPath); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	if err := logger.Init(); err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}

	if s := config.C.Server.JWTSecret; s == "change-me" || len(s) < 32 {
		log.Fatalf("jwt_secret is insecure: must be at least 32 characters and not 'change-me'")
	}

	db, err := database.Connect()
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	model.SetDB(db)

	devMode := os.Getenv("NEXUS_DEV") == "1"
	if devMode {
		if err := model.AutoMigrate(); err != nil {
			log.Fatalf("failed to migrate: %v", err)
		}
	}

	llm.Init(db)
	cache.Init()
	if err := vectordb.Init(config.C.Milvus.Addr, config.C.Milvus.Collection, config.C.Milvus.Dimension); err != nil {
		log.Fatalf("failed to init milvus: %v", err)
	}
	sse.Init()
	processor.Init()

	runAPI := *mode == "all" || *mode == "api"
	runWorker := *mode == "all" || *mode == "worker"

	var workerSrv *asynq.Server
	var asynqClient *asynq.Client

	if config.C.Redis.Addr != "" {
		asynqClient = worker.NewClient()
		if runWorker {
			jobSvc := service.NewJobService(db, asynqClient, sse.Default(), llm.G)
			workerSrv = startWorker(jobSvc)
			if err := jobSvc.RecoverStalled(); err != nil {
				log.Fatalf("failed to recover stalled jobs: %v", err)
			}
		}
	} else {
		logger.Info("redis not configured, worker disabled")
	}

	var httpSrv *http.Server
	if runAPI {
		r := api.SetupRouter(db, asynqClient, sse.Default(), llm.G)
		addr := fmt.Sprintf(":%d", config.C.Server.Port)
		httpSrv = &http.Server{Addr: addr, Handler: r}

		go func() {
			logger.Info("server starting on " + addr + " (mode: " + *mode + ")")
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("failed to start server: %v", err)
			}
		}()
	}

	if !runAPI && !runWorker {
		log.Fatalf("invalid mode: %s (must be api, worker, or all)", *mode)
	}

	if !runAPI && runWorker {
		logger.Info("worker-only mode started")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		httpSrv.Shutdown(ctx)
	}

	if workerSrv != nil {
		workerDone := make(chan struct{})
		go func() {
			workerSrv.Shutdown()
			close(workerDone)
		}()
		select {
		case <-workerDone:
			logger.Info("worker shutdown completed")
		case <-time.After(120 * time.Second):
			logger.Warn("worker shutdown timed out after 120s")
			workerSrv.Stop()
		}
	}
	if asynqClient != nil {
		asynqClient.Close()
	}
	vectordb.Close()

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}
	logger.Sync()
	logger.Info("server exited")
}

func startWorker(jobSvc *service.JobService) *asynq.Server {
	srv := worker.NewServer()
	mux := asynq.NewServeMux()
	worker.RegisterHandlers(mux, jobSvc)
	go func() {
		logger.Info("worker starting...")
		if err := srv.Run(mux); err != nil {
			log.Printf("worker stopped: %v", err)
		}
	}()
	return srv
}
