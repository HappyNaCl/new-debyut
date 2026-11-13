package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/HappyNaCl/new-debyut/api/swagger"
	grpcclient "github.com/HappyNaCl/new-debyut/internal/common/grpcclient"
	l "github.com/HappyNaCl/new-debyut/internal/common/logger"
	"github.com/HappyNaCl/new-debyut/internal/gateway/routes"
	"github.com/HappyNaCl/new-debyut/internal/genproto/auth"
	"github.com/HappyNaCl/new-debyut/internal/genproto/event"
	"github.com/HappyNaCl/new-debyut/internal/genproto/users"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title debyut API Gateway
// @version 1.0
// @description REST API exposed by the debyut api-gateway. The gateway translates these HTTP calls into gRPC requests to the auth, users, and event services.
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	if err := l.Init(); err != nil {
		log.Fatalf("Failed to init zaplogger: %v", err)
	}
	defer l.Sync()

	r := gin.Default()
	r.Use(
		cors.New(cors.Config{
			AllowOrigins:     allowedOrigins(),
			AllowMethods:     []string{"GET", "DELETE", "PUT", "POST", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			AllowCredentials: true,
		}),
	)

	port := flag.String("port", "8080", "Port the API will listen on")
	flag.Parse()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")
	registerRoutes(api)

	addr := fmt.Sprintf(":%s", *port)
	r.Run(addr)
}

func registerRoutes(api *gin.RouterGroup) {
	authConn, err := grpcclient.NewGrpcClient(os.Getenv("AUTH_SERVICE_ADDR"))
	if err != nil {
		panic("Cannot connect to authService")
	}
	authClient := auth.NewAuthServiceClient(authConn)

	userConn, err := grpcclient.NewGrpcClient(os.Getenv("USER_SERVICE_ADDR"))
	if err != nil {
		panic("Cannot connect to userService")
	}
	userClient := users.NewUserServiceClient(userConn)
	applicationClient := users.NewOrganizerApplicationServiceClient(userConn)
	organizationClient := users.NewOrganizationServiceClient(userConn)

	eventConn, err := grpcclient.NewGrpcClient(os.Getenv("EVENT_SERVICE_ADDR"))
	if err != nil {
		panic("Cannot connect to eventService")
	}
	eventClient := event.NewEventServiceClient(eventConn)

	routes.RegisterHealthRoutes(api)
	routes.RegisterAuthRoutes(api, authClient)
	routes.RegisterUserRoutes(api, userClient)
	routes.RegisterOrganizerApplicationRoutes(api, applicationClient)
	routes.RegisterOrganizationRoutes(api, organizationClient)
	routes.RegisterCategoryRoutes(api, eventClient)
}

// allowedOrigins returns the default local frontend origins plus any extra
// comma-separated origins in CORS_ALLOWED_ORIGINS (e.g. a LAN address used to
// open the dev frontend from another device).
func allowedOrigins() []string {
	origins := []string{"http://localhost:5173", "http://frontend:5173"}
	for _, origin := range strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}
