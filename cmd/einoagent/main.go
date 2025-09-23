package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"otel-eino-example/cmd/einoagent/agent"
	"otel-eino-example/cmd/einoagent/task"
	"otel-eino-example/pkg/env"
)

func init() {
	env.MustHasEnvs("OPENAI_EMBEDDING_MODEL", "OPENAI_API_KEY", "OPENAI_EMBEDDING_BASE_URL", "OPENAI_CHAT_MODEL", "OPENAI_CHAT_BASE_URL")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	h := server.Default(server.WithHostPorts(":" + port))

	h.Use(LogMiddleware())

	taskGroup := h.Group("/task")
	if err := task.BindRoutes(taskGroup); err != nil {
		log.Fatal("failed to bind task routes:", err)
	}

	agentGroup := h.Group("/agent")
	if err := agent.BindRoutes(agentGroup); err != nil {
		log.Fatal("failed to bind agent routes:", err)
	}

	h.GET("/", func(ctx context.Context, c *app.RequestContext) {
		c.Redirect(302, []byte("/agent"))
	})

	h.Spin()
}

func LogMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		path := string(c.Request.URI().Path())
		method := string(c.Request.Method())
		c.Next(ctx)
		latency := time.Since(start)
		statusCode := c.Response.StatusCode()
		log.Printf("[HTTP] %s %s %d %v\n", method, path, statusCode, latency)
	}
}
