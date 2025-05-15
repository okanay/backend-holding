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
	mw "github.com/okanay/backend-holding/middlewares"
	air "github.com/okanay/backend-holding/repositories/ai"
	fr "github.com/okanay/backend-holding/repositories/file"
	r2r "github.com/okanay/backend-holding/repositories/r2"
	tr "github.com/okanay/backend-holding/repositories/token"
	ur "github.com/okanay/backend-holding/repositories/user"
	"github.com/okanay/backend-holding/services/cache"
)

// Uygulama bileşenlerini gruplamak için yapılar
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
	_ = initServices()

	// 5. Handler Katmanını Başlat
	h := initHandlers(r)

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
