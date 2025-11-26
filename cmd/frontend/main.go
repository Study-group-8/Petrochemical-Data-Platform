package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const defaultFrontendPort = "3000"

func main() {
	var (
		dirFlag  = flag.String("dir", "", "Каталог со статическими файлами фронтенда")
		portFlag = flag.String("port", "", "Порт HTTP сервера (по умолчанию 3000)")
	)
	flag.Parse()

	assetsSource, err := newAssetSource(*dirFlag)
	if err != nil {
		log.Fatal(err)
	}
	port := resolvePort(*portFlag)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if served := tryServeStatic(w, r, assetsSource); served {
			return
		}
		serveIndex(w, r, assetsSource)
	})

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Frontend server listening on :%s (assets: %s)", port, assetsSource.description)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("frontend server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal received, stopping frontend server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Graceful shutdown failed: %v", err)
	}
}

type assetSource struct {
	fsys        fs.FS
	fileServer  http.Handler
	description string
}

func newAssetSource(flagValue string) (*assetSource, error) {
	dir, err := resolveRootDir(flagValue)
	if err != nil {
		return nil, err
	}

	dirFS := os.DirFS(dir)
	if _, err := fs.Stat(dirFS, "index.html"); err != nil {
		return nil, fmt.Errorf("в каталоге %s нет index.html: %w", dir, err)
	}
	return &assetSource{
		fsys:        dirFS,
		fileServer:  http.FileServer(http.Dir(dir)),
		description: dir,
	}, nil
}

func resolveRootDir(flagValue string) (string, error) {
	for _, candidate := range frontendDirCandidates(flagValue) {
		if dir, ok := validateFrontendDir(candidate); ok {
			return dir, nil
		}
	}
	return "", fmt.Errorf("не найден каталог фронтенда — укажите путь через флаг --dir или переменную FRONTEND_DIR")
}

func frontendDirCandidates(flagValue string) []string {
	unique := make(map[string]struct{})
	order := make([]string, 0)
	appendUnique := func(path string) {
		path = strings.TrimSpace(path)
		if path == "" {
			return
		}
		cleaned := filepath.Clean(path)
		if _, ok := unique[cleaned]; ok {
			return
		}
		unique[cleaned] = struct{}{}
		order = append(order, cleaned)
	}

	appendUnique(flagValue)
	appendUnique(os.Getenv("FRONTEND_DIR"))

	if cwd, err := os.Getwd(); err == nil {
		appendUnique(filepath.Join(cwd, "frontend"))
		for _, candidate := range searchUpwardsForFrontend(cwd) {
			appendUnique(candidate)
		}
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		appendUnique(filepath.Join(exeDir, "frontend"))
		for _, candidate := range searchUpwardsForFrontend(exeDir) {
			appendUnique(candidate)
		}
	}

	if src := sourceFrontendPath(); src != "" {
		appendUnique(src)
	}

	return order
}

func searchUpwardsForFrontend(start string) []string {
	var candidates []string
	current := filepath.Clean(start)
	for {
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		candidates = append(candidates, filepath.Join(parent, "frontend"))
		current = parent
	}
	return candidates
}

func sourceFrontendPath() string {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "frontend"))
}

func validateFrontendDir(path string) (string, bool) {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return "", false
	}
	indexPath := filepath.Join(path, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		return "", false
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", false
	}
	return abs, true
}

func resolvePort(flagValue string) string {
	if port := strings.TrimSpace(flagValue); port != "" {
		return port
	}
	if env := strings.TrimSpace(os.Getenv("FRONTEND_PORT")); env != "" {
		return env
	}
	return defaultFrontendPort
}

func tryServeStatic(w http.ResponseWriter, r *http.Request, assets *assetSource) bool {
	rel := strings.TrimPrefix(r.URL.Path, "/")
	rel = path.Clean(rel)
	if rel == "." || rel == "" {
		return false
	}
	if strings.HasPrefix(rel, "..") {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return true
	}

	if _, err := fs.Stat(assets.fsys, rel); err != nil {
		return false
	}

	assets.fileServer.ServeHTTP(w, r)
	return true
}

func serveIndex(w http.ResponseWriter, r *http.Request, assets *assetSource) {
	data, err := fs.ReadFile(assets.fsys, "index.html")
	if err != nil {
		http.Error(w, "index file not found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		log.Printf("failed to write index.html response: %v", err)
	}
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(recorder, r)
		log.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, recorder.status, time.Since(start).Round(time.Millisecond))
	})
}
