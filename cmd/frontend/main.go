package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	// Get the directory of the executable
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	exeDir := filepath.Dir(exePath)

	// Serve static files from the frontend directory
	fs := http.FileServer(http.Dir(filepath.Join(exeDir, "frontend")))

	// Handle all routes by serving the index.html for client-side routing
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check if the requested file exists
		path := filepath.Join(exeDir, "frontend", r.URL.Path)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// If file doesn't exist, serve index.html
			http.ServeFile(w, r, filepath.Join(exeDir, "frontend", "index.html"))
		} else {
			// Serve the requested file
			fs.ServeHTTP(w, r)
		}
	})

	log.Println("Frontend server starting on :3000")
	log.Println("Open http://localhost:3000 in your browser")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
