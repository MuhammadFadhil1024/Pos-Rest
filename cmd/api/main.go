package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	handler "pos-backend/internal/handler/http"
	"pos-backend/internal/handler/http/middleware"
	"pos-backend/internal/repository/mysql"
	"pos-backend/internal/service"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if present
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		secretKey = "FadhilGanteng123"
	}

	dbDsn := os.Getenv("DB_DSN")
	if dbDsn == "" {
		dbDsn = "root@tcp(127.0.0.1:3306)/pos?parseTime=true&loc=Asia%2FJakarta"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, err := sql.Open("mysql", dbDsn)

	if err != nil {
		log.Fatal("Failed to insiasi DB driver", err)
	}

	defer db.Close()

	db.SetMaxOpenConns(25)                 // Batas maksimal koneksi terbuka ke DB
	db.SetMaxIdleConns(25)                 // Jumlah koneksi idle yang bersiap di pool
	db.SetConnMaxLifetime(5 * time.Minute) // Waktu maksimal koneksi sebelum di-recycle
	db.SetConnMaxIdleTime(5 * time.Minute) // Waktu maksimal koneksi idle menganggur

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to connect DB", err)
	}

	log.Print("Success connect to DB")

	var sessionTZ string
	var dbNow, dbUTC time.Time
	if err := db.QueryRow("SELECT @@session.time_zone, NOW(), UTC_TIMESTAMP()").Scan(&sessionTZ, &dbNow, &dbUTC); err != nil {
		log.Printf("Failed to read DB timezone: %v", err)
	} else {
		log.Printf("DB session time_zone=%s, NOW()=%s, UTC_TIMESTAMP()=%s, offset=%s",
			sessionTZ, dbNow.Format(time.RFC3339), dbUTC.Format(time.RFC3339), dbNow.Sub(dbUTC))
		if dbNow.Sub(dbUTC) != 7*time.Hour {
			log.Printf("WARNING: DB timezone is not UTC+7 (Asia/Jakarta). Date filters on ticket.created_at may be shifted.")
		}
	}

	userRepo := mysql.NewUserRepository(db)
	branchRepo := mysql.NewBranchRepository(db)
	ServiceRepo := mysql.NewServiceRepository(db)
	PaymentRepo := mysql.NewPaymentRepository(db, ServiceRepo)
	authService := service.NewAuthService(userRepo, secretKey)
	_= authService
	branchService := service.NewBranchService(branchRepo)
	_= branchService
	serviceTypeService := service.NewServiceTypeService(ServiceRepo, PaymentRepo)
	_= serviceTypeService
	paymentService := service.NewPaymentService(PaymentRepo)
	_= paymentService
	ticketService := service.NewTicketService(mysql.NewTicketRepository(db), mysql.NewMenuRepository(db))
	_= ticketService


	authHandler := handler.NewAuthHandler(authService)
	branchHandler := handler.NewBranchHandler(branchService)
	preloadHandler := handler.NewPreloadHandler(serviceTypeService, branchService, paymentService)
	ticketHandler := handler.NewTicketHandler(ticketService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("GET /api/v1/branches", middleware.RequireAuth(secretKey)(branchHandler.GeMyBranches))
	mux.HandleFunc("GET /api/v1/preload/{id}", preloadHandler.GetPreloadData)
	mux.HandleFunc("POST /api/v1/tickets", ticketHandler.CreateTicket)
	mux.HandleFunc("GET /api/v1/tickets", middleware.RequireAuth(secretKey)(ticketHandler.GetTickets))

	log.Printf("Server POS berjalan di port :%s...\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal("Server error:", err)
	}
}
