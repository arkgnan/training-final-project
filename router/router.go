package router

import (
	"log"
	"mygram-api/controllers"
	"mygram-api/database"
	"mygram-api/middlewares"
	"os"
	"time"

	_ "mygram-api/docs" // Import generated swagger docs

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const MaxRequests = 10
const RateWindow = time.Minute

func SetupRouter() *gin.Engine {
	appLogger := log.Default()
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	r := gin.Default()
	r.Use(gin.Recovery())
	r.Use(middlewares.CORSConfig())

	// Add Swagger UI endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// --- Public Endpoints (No Auth Required) ---
	userController := controllers.NewUserController(database.GetDB(), appLogger)
	r.POST("/auth/register", userController.Register) // POST /users/register
	r.POST("/auth/login", userController.Login)       // POST /users/login

	// --- Authenticated Endpoints (Auth Required) ---
	authRouter := r.Group("/")
	authRouter.Use(middlewares.Authentication()) // Apply JWT Auth to all routes in this group

	// Users (PUT/DELETE require only Auth, no Authorization needed as it's self-modifying)
	authRouter.PUT("/users", userController.Update)    // PUT /users
	authRouter.DELETE("/users", userController.Delete) // DELETE /users

	// Photos
	photoController := controllers.NewPhotoController(database.GetDB(), appLogger)
	authRouter.POST("/photos", middlewares.RateLimiterConfig(MaxRequests, RateWindow), photoController.Create) // POST /photos
	authRouter.GET("/photos", photoController.GetAll)                                                          // GET /photos

	// Photos (PUT/DELETE require Auth AND Authorization)
	photoAuthRouter := authRouter.Group("/photos")
	photoAuthRouter.Use(middlewares.Authorization("photo"))
	{
		photoAuthRouter.PUT("/:photoID", photoController.Update)    // PUT /photos/:photoID
		photoAuthRouter.DELETE("/:photoID", photoController.Delete) // DELETE /photos/:photoID
	}

	// Comments
	commentController := controllers.NewCommentController(database.GetDB(), appLogger)
	authRouter.POST("/comments", middlewares.RateLimiterConfig(MaxRequests, RateWindow), commentController.Create) // POST /comments
	authRouter.GET("/comments", commentController.GetAll)                                                          // GET /comments

	authRouter.POST("/comments/reply/:parentCommentID", middlewares.RateLimiterConfig(MaxRequests, RateWindow), commentController.CreateReply)
	authRouter.GET("/comments/:parentCommentID/replies", commentController.GetReplies)

	// Comments (PUT/DELETE require Auth AND Authorization)
	commentAuthRouter := authRouter.Group("/comments")
	commentAuthRouter.Use(middlewares.Authorization("comment"))
	{
		commentAuthRouter.PUT("/:commentID", commentController.Update)    // PUT /comments/:commentID
		commentAuthRouter.DELETE("/:commentID", commentController.Delete) // DELETE /comments/:commentID
	}

	// SocialMedias
	socialMediaController := controllers.NewSocialMediaController(database.GetDB(), appLogger)
	authRouter.POST("/socialmedias", socialMediaController.Create) // POST /socialmedias
	authRouter.GET("/socialmedias", socialMediaController.GetAll)  // GET /socialmedias

	// SocialMedias (PUT/DELETE require Auth AND Authorization)
	smAuthRouter := authRouter.Group("/socialmedias")
	smAuthRouter.Use(middlewares.Authorization("socialmedia"))
	{
		smAuthRouter.PUT("/:socialmediaID", socialMediaController.Update)    // PUT /socialmedias/:socialMediaID
		smAuthRouter.DELETE("/:socialmediaID", socialMediaController.Delete) // DELETE /socialmedias/:socialMediaID
	}

	return r
}
