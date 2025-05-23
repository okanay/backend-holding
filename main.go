package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	c "github.com/okanay/backend-holding/configs"
	db "github.com/okanay/backend-holding/database"
	ch "github.com/okanay/backend-holding/handlers/content"
	fh "github.com/okanay/backend-holding/handlers/file"
	mh "github.com/okanay/backend-holding/handlers/globals"
	jh "github.com/okanay/backend-holding/handlers/job"
	uh "github.com/okanay/backend-holding/handlers/user"

	"github.com/okanay/backend-holding/middlewares"
	mw "github.com/okanay/backend-holding/middlewares"
	air "github.com/okanay/backend-holding/repositories/ai"
	cr "github.com/okanay/backend-holding/repositories/content"
	fr "github.com/okanay/backend-holding/repositories/file"
	jr "github.com/okanay/backend-holding/repositories/job"
	r2r "github.com/okanay/backend-holding/repositories/r2"
	tr "github.com/okanay/backend-holding/repositories/token"
	ur "github.com/okanay/backend-holding/repositories/user"

	"github.com/okanay/backend-holding/services/cache"
)

type Repositories struct {
	User    *ur.Repository
	Token   *tr.Repository
	AI      *air.Repository
	File    *fr.Repository
	R2      *r2r.Repository
	Job     *jr.Repository
	Content *cr.Repository
}

type Services struct {
	Cache cache.CacheService
}
type Handlers struct {
	Main    *mh.Handler
	User    *uh.Handler
	File    *fh.Handler
	Job     *jh.Handler
	Content *ch.Handler
}

func main() {
	// 1. Çevresel Değişkenleri Yükle
	if err := godotenv.Load(".env"); err != nil {
		log.Println("[ENV]: .env file could not be loaded, environment variables will be used")
	} else {
		log.Println("[ENV]: .env file loaded successfully.")
	}

	// 2. Veritabanı Bağlantısı Kur
	sqlDB, err := db.Init(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("[DATABASE]: Error connecting to the database: %v", err)
	} else {
		log.Println("[DATABASE]: Database connection established successfully")
		defer sqlDB.Close()
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// 3. Servisleri ve Handler'ları Başlat
	repos := initRepositories(sqlDB)
	services := initServices()
	handlers := initHandlers(repos, services)

	// 4. Router ve Middleware Yapılandırması
	router := gin.Default()
	router.Use(c.CorsConfig())
	router.Use(c.SecureConfig)
	router.MaxMultipartMemory = 10 << 20

	// 4.0 Global Middlewares
	router.Use(middlewares.TimeoutMiddleware())

	// 4.1 Middlewares Initialize
	Recaptcha := middlewares.NewRecaptchaMiddleware(os.Getenv("RECAPTCHA_SECRET_KEY"), 0.7)
	defer Recaptcha.Close()

	// 4.2 Routes
	publicAPI := router.Group("/public")
	publicFileAPI := router.Group("/public/files")
	internalAPI := router.Group("/internal")
	authAPI := router.Group("/auth")

	// 4.3 Middlewares
	publicAPI.Use(mw.RateLimiterMiddleware(60, time.Minute))
	internalAPI.Use(mw.RateLimiterMiddleware(1000, time.Minute))

	authAPI.Use(mw.RateLimiterMiddleware(120, time.Minute))
	authAPI.Use(mw.AuthMiddleware(repos.User, repos.Token))

	publicFileAPI.Use(mw.RateLimiterMiddleware(4, 120*time.Minute))

	// `start with /`
	router.GET("/", handlers.Main.Index)
	router.NoRoute(handlers.Main.NotFound)

	// `start with /public`
	publicAPI.POST("/login", handlers.User.Login)
	publicAPI.POST("/register", handlers.User.Register)

	publicAPI.GET("/jobs", handlers.Job.ListPublishedJobs)
	publicAPI.GET("/jobs/:id", handlers.Job.GetJobBySlug)
	publicAPI.POST("/jobs/:id", handlers.Job.CreateJobApplication)

	publicAPI.GET("/contents", handlers.Content.ListPublishedContents)
	publicAPI.GET("/contents/:lang/:slug", handlers.Content.GetContentBySlug)

	// `start with /auth`
	authAPI.GET("/logout", handlers.User.Logout)
	authAPI.GET("/get-me", handlers.User.GetMe)

	authAPI.GET("/jobs", handlers.Job.ListJobs)
	authAPI.GET("/job/:id", handlers.Job.GetJobByID)
	authAPI.POST("/create-new-job", handlers.Job.CreateJob)
	authAPI.PATCH("/job/:id", handlers.Job.UpdateJob)
	authAPI.DELETE("/job/:id", handlers.Job.DeleteJob)
	authAPI.PATCH("/job/status/:id", handlers.Job.UpdateJobStatus)

	authAPI.GET("/applicants", handlers.Job.ListJobApplications)
	authAPI.PATCH("/applicant/status/:id", handlers.Job.UpdateJobApplicationStatus)

	authAPI.GET("/contents", handlers.Content.ListContents)
	authAPI.GET("/content/:id", handlers.Content.GetContentByID)
	authAPI.PATCH("/content/restore/:id", handlers.Content.RestoreContent)

	authAPI.POST("/content", handlers.Content.CreateContent)
	authAPI.PATCH("/content/:id", handlers.Content.UpdateContent)
	authAPI.DELETE("/content/:id", handlers.Content.DeleteContent)
	authAPI.PATCH("/content/status/:id", handlers.Content.UpdateContentStatus)

	// `start with /public/files`
	publicFileAPI.POST("/presigned-url", handlers.File.CreatePresignedURL)
	publicFileAPI.POST("/confirm-upload", handlers.File.ConfirmUpload)

	// `start with /auth`
	authAPI.GET("/files/category", handlers.File.GetFilesByCategory)
	authAPI.POST("/files/presigned-url", handlers.File.CreatePresignedURL)
	authAPI.POST("/files/confirm-upload", handlers.File.ConfirmUpload)
	authAPI.DELETE("/files/:id", handlers.File.DeleteFile)

	// 5. Sunucuyu Başlat
	startServer(router)

}

// Repository'lerin başlatılması
func initRepositories(sqlDB *sql.DB) Repositories {
	return Repositories{
		User:    ur.NewRepository(sqlDB),
		Token:   tr.NewRepository(sqlDB),
		AI:      air.NewRepository(os.Getenv("OPENAI_API_KEY")),
		File:    fr.NewRepository(sqlDB),
		Job:     jr.NewRepository(sqlDB),
		Content: cr.NewRepository(sqlDB),
		R2: r2r.NewRepository(
			os.Getenv("R2_ACCOUNT_ID"),
			os.Getenv("R2_ACCESS_KEY_ID"),
			os.Getenv("R2_ACCESS_KEY_SECRET"),
			os.Getenv("R2_BUCKET_NAME"),
			os.Getenv("R2_FOLDER_NAME"),
			os.Getenv("R2_PUBLIC_URL_BASE"),
			os.Getenv("R2_ENDPOINT"),
		),
	}
}

// initServices fonksiyonunu da güncelle
func initServices() Services {
	// Cache oluştur
	cacheService := cache.NewCacheService(1 * time.Hour)

	return Services{
		Cache: cacheService, // İşaretçi dönüştürme yapmadan doğrudan atama
	}
}

// Handler'ların başlatılması
func initHandlers(repos Repositories, services Services) Handlers {
	return Handlers{
		Main:    mh.NewHandler(),
		User:    uh.NewHandler(repos.User, repos.Token),
		File:    fh.NewHandler(repos.File, repos.R2),
		Job:     jh.NewHandler(repos.File, repos.R2, repos.Job, services.Cache),
		Content: ch.NewHandler(repos.Content, services.Cache),
	}
}

// Sunucuyu başlatır
func startServer(router *gin.Engine) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Varsayılan port
	}

	log.Printf("[SERVER]: %s portu üzerinde dinleniyor...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("[SERVER]: Sunucu başlatılırken hata: %v", err)
	}
}
