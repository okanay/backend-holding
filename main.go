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
	fh "github.com/okanay/backend-holding/handlers/file"
	mh "github.com/okanay/backend-holding/handlers/main"
	uh "github.com/okanay/backend-holding/handlers/user"
	"github.com/okanay/backend-holding/middlewares"
	mw "github.com/okanay/backend-holding/middlewares"
	air "github.com/okanay/backend-holding/repositories/ai"
	fr "github.com/okanay/backend-holding/repositories/file"
	r2r "github.com/okanay/backend-holding/repositories/r2"
	tr "github.com/okanay/backend-holding/repositories/token"
	ur "github.com/okanay/backend-holding/repositories/user"
	"github.com/okanay/backend-holding/services/cache"
)

type Repositories struct {
	User  *ur.Repository
	Token *tr.Repository
	AI    *air.Repository
	File  *fr.Repository
	R2    *r2r.Repository
}

type Services struct {
	BlogCache *cache.Cache
}

type Handlers struct {
	Main *mh.Handler
	User *uh.Handler
	File *fh.Handler
}

func main() {
	// 1. Çevresel Değişkenleri Yükle
	loadEnvironmentVariables()

	// 2. Veritabanı Bağlantısı Kur
	sqlDB := setupDatabase()
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	defer sqlDB.Close()

	// 3. Servisleri ve Handler'ları Başlat
	repos := initRepositories(sqlDB)
	handlers := initHandlers(repos)
	_ = initServices()

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

	publicFileAPI.Use(mw.RateLimiterMiddleware(10, 15*time.Minute))
	publicFileAPI.Use(Recaptcha.Middleware())

	// `start with /`
	router.GET("/", handlers.Main.Index)
	router.NoRoute(handlers.Main.NotFound)

	// `start with /public`
	publicAPI.POST("/login", handlers.User.Login)
	publicAPI.POST("/register", handlers.User.Register)

	// `start with /auth`
	authAPI.GET("/logout", handlers.User.Logout)
	authAPI.GET("/get-me", handlers.User.GetMe)

	// `start with /public/files`
	publicFileAPI.POST("/presigned-url", handlers.File.CreatePresignedURL)
	publicFileAPI.POST("/confirm-upload", handlers.File.ConfirmUpload)

	// `start with /auth`
	authAPI.GET("/files/category", handlers.File.GetFilesByCategory)
	authAPI.DELETE("/files/:id", handlers.File.DeleteFile)

	// 5. Sunucuyu Başlat
	startServer(router)
}

// Çevresel değişkenleri yükler
func loadEnvironmentVariables() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("[ENV]: .env dosyası yüklenemedi, ortam değişkenleri kullanılacak")
	}
}

// Veritabanı bağlantısını kurar ve yapılandırır
func setupDatabase() *sql.DB {
	sqlDB, err := db.Init(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("[DATABASE]: Veritabanına bağlanırken hata: %v", err)
	}

	log.Println("[DATABASE]: Veritabanı bağlantısı başarıyla kuruldu")
	return sqlDB
}

// Repository'lerin başlatılması
func initRepositories(sqlDB *sql.DB) Repositories {
	return Repositories{
		User:  ur.NewRepository(sqlDB),
		Token: tr.NewRepository(sqlDB),
		AI:    air.NewRepository(os.Getenv("OPENAI_API_KEY")),
		File:  fr.NewRepository(sqlDB),
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

// Servislerin başlatılması
func initServices() Services {
	// Cache ve servis oluştur
	blogCache := cache.NewCache(30 * time.Minute)
	return Services{
		BlogCache: blogCache,
	}
}

// Handler'ların başlatılması
func initHandlers(repos Repositories) Handlers {
	return Handlers{
		Main: mh.NewHandler(),
		User: uh.NewHandler(repos.User, repos.Token),
		File: fh.NewHandler(repos.File, repos.R2),
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
