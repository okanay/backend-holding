package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	// Proje modülleri

	c "github.com/okanay/backend-holding/configs"
	db "github.com/okanay/backend-holding/database"
	"github.com/okanay/backend-holding/handlers"
	ImageHandler "github.com/okanay/backend-holding/handlers/image"
	UserHandler "github.com/okanay/backend-holding/handlers/user"
	mw "github.com/okanay/backend-holding/middlewares"
	AIRepository "github.com/okanay/backend-holding/repositories/ai"
	ImageRepository "github.com/okanay/backend-holding/repositories/image"
	R2Repository "github.com/okanay/backend-holding/repositories/r2"
	TokenRepository "github.com/okanay/backend-holding/repositories/token"
	UserRepository "github.com/okanay/backend-holding/repositories/user"
	"github.com/okanay/backend-holding/services/cache"
)

// Uygulama bileşenlerini gruplamak için yapılar
type Repositories struct {
	User  *UserRepository.Repository
	Token *TokenRepository.Repository
	AI    *AIRepository.Repository
	Image *ImageRepository.Repository
	R2    *R2Repository.Repository
}

type Services struct {
	BlogCache *cache.Cache
}

type Handlers struct {
	Main  *handlers.Handler
	User  *UserHandler.Handler
	Image *ImageHandler.Handler
}

func main() {
	// 1. Çevresel Değişkenleri Yükle
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("[ENV]: .env dosyası yüklenemedi")
		return
	}

	// 2. Veritabanı Bağlantısı Kur
	sqlDB, err := db.Init(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("[DATABASE]: Veritabanına bağlanırken hata: %v", err)
		return
	}
	defer sqlDB.Close()

	// 3. Repository Katmanını Başlat
	r := initRepositories(sqlDB)

	// 4. Servis Katmanını Başlat
	s := initServices(r)

	// 5. Handler Katmanını Başlat
	h := initHandlers(r, s)

	// 6. Router ve Middleware Yapılandırması
	router := gin.Default()
	router.Use(c.CorsConfig())
	router.Use(c.SecureConfig)
	router.MaxMultipartMemory = 10 << 20 // 10 MB

	// Kimlik doğrulama gerektiren rotalar için grup
	auth := router.Group("/auth")
	auth.Use(mw.AuthMiddleware(r.User, r.Token))

	// Global Routes
	router.GET("/", h.Main.Index)
	router.NoRoute(h.Main.NotFound)

	// Authentication Routes (public)
	router.POST("/login", h.User.Login)
	router.POST("/register", h.User.Register)
	auth.GET("/logout", h.User.Logout)
	auth.GET("/get-me", h.User.GetMe)

	// Image Routes
	imageAuth := auth.Group("/images")
	{
		imageAuth.POST("/presign", h.Image.CreatePresignedURL)
		imageAuth.POST("/confirm", h.Image.ConfirmUpload)
		imageAuth.GET("", h.Image.GetUserImages)
		imageAuth.DELETE("/:id", h.Image.DeleteImage)
	}

	// 7. Sunucuyu Başlat
	port := os.Getenv("PORT")
	log.Printf("[SERVER]: %s portu üzerinde dinleniyor...", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("[SERVER]: Sunucu başlatılırken hata: %v", err)
	}
}

// Repository'lerin başlatılması
func initRepositories(sqlDB *sql.DB) Repositories {
	return Repositories{
		User:  UserRepository.NewRepository(sqlDB),
		Token: TokenRepository.NewRepository(sqlDB),
		AI:    AIRepository.NewRepository(os.Getenv("OPENAI_API_KEY")),
		Image: ImageRepository.NewRepository(sqlDB),
		R2: R2Repository.NewRepository(
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
func initServices(repos Repositories) Services {
	// Cache ve servis oluştur
	blogCache := cache.NewCache(30 * time.Minute)

	return Services{
		BlogCache: blogCache,
	}
}

// Handler'ların başlatılması
func initHandlers(repos Repositories, services Services) Handlers {
	return Handlers{
		Main:  handlers.NewHandler(),
		User:  UserHandler.NewHandler(repos.User, repos.Token),
		Image: ImageHandler.NewHandler(repos.Image, repos.R2),
	}
}
