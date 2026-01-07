package main

import (
	"crypto/subtle"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tus/tusd/pkg/filestore"
	tusd "github.com/tus/tusd/pkg/handler"
	"golang.org/x/time/rate"
)

// rate limiter
var visitors = make(map[string]*visitor)
var mu sync.Mutex

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if !exists {
		// 2 req per s, 5 burst
		limiter := rate.NewLimiter(2, 5)
		visitors[ip] = &visitor{limiter, time.Now()}
		return limiter
	}

	// for last visit
	v.lastSeen = time.Now()
	return v.limiter
}

// clean memory and limiter
func cleanupVisitors() {
	for {
		time.Sleep(5 * time.Minute)
		mu.Lock()
		for ip, v := range visitors {
			// del after 10min
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}

// middleware
func limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// for cloudflare
		ip := r.Header.Get("CF-Connecting-IP")
		if ip == "" {
			// fallback
			ip = r.RemoteAddr
		}

		limiter := getVisitor(ip)
		if !limiter.Allow() {
			http.Error(w, "Too Many Requests - Calmeaza-te!", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// delete old files
func cleanExpiredFiles(uploadDir string) {
	for {
		time.Sleep(1 * time.Hour)
		files, err := os.ReadDir(uploadDir)
		if err != nil {
			log.Println("eroare citire director", err)
			continue
		}

		for _, file := range files {
			info, err := file.Info()
			if err != nil {
				continue
			}
			if time.Since(info.ModTime()) > 24*time.Hour {
				os.Remove(filepath.Join(uploadDir, file.Name()))
				log.Println("sters fisier expirat", file.Name())
			}
		}
	}
}

// new middleware
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// prevent guess of file
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// iframe guess
		w.Header().Set("X-Frame-Options", "DENY")
        
        // https
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")


        w.Header().Set("Referrer-Policy", "no-referrer")

		path := r.URL.Path
		if len(path) > 7 && path[:7] == "/files/" && r.Method == "GET" {
			w.Header().Set("Content-Disposition", "attachment")
		}

		// csp
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' https://releases.transloadit.com; style-src 'self' https://releases.transloadit.com 'unsafe-inline'; img-src 'self' data:; object-src 'none'; base-uri 'none';")

		next.ServeHTTP(w, r)
	})
}

func main() {
	basePath := "/files/"
	uploadDir := "./uploads"

	// secret verify on start
	adminSecret := os.Getenv("UPLOAD_SECRET")
	if adminSecret == "" {
		log.Fatal("eroare: variabila UPLOAD_SECRET nu este setata")
	}

	// make sure the folder exists
	os.MkdirAll(uploadDir, 0755)

	// background clean
	go cleanExpiredFiles(uploadDir)
	go cleanupVisitors()

	store := filestore.New(uploadDir)
	composer := tusd.NewStoreComposer()
	store.UseIn(composer)

	handler, err := tusd.NewHandler(tusd.Config{
		BasePath:              basePath,
		StoreComposer:         composer,
		NotifyCompleteUploads: true,
		// 1gb upload limit
		MaxSize: 1024 * 1024 * 1024,
	})
	if err != nil {
		log.Fatal(err)
	}

	// middleware for security
	// verify token
	uploadHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPatch || r.Method == http.MethodDelete {
			userToken := r.Header.Get("X-Auth-Token")

			// timing attacks
			if subtle.ConstantTimeCompare([]byte(userToken), []byte(adminSecret)) != 1 {
				http.Error(w, "interzis", http.StatusUnauthorized)
				return
			}
		}
		handler.ServeHTTP(w, r)
	})

	// limit -> security -> handler
	http.Handle(basePath, limit(securityHeaders(uploadHandler)))

	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/", limit(securityHeaders(fileServer)))

	log.Println("port 8080")

	srv := &http.Server{
		Addr: "0.0.0.0:8080",
		Handler: nil,
		// header attacks
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout: 120 * time.Second,
	}
	
	srv.ListenAndServe()
}