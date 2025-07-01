package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/TheEntropyCollective/randomfs-core"
	"github.com/gorilla/mux"
)

var (
	httpPort  = flag.Int("port", 8080, "HTTP server port")
	ipfsAPI   = flag.String("ipfs", "http://localhost:5001", "IPFS API endpoint")
	dataDir   = flag.String("data", "./data", "Data directory")
	cacheSize = flag.Int64("cache", 500*1024*1024, "Cache size in bytes")
	webDir    = flag.String("web", "./web", "Web interface directory")
)

type Server struct {
	rfs *randomfs.RandomFS
}

func NewServer(rfs *randomfs.RandomFS) *Server {
	return &Server{rfs: rfs}
}

func (s *Server) handleStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	randomURL, err := s.rfs.StoreFile(header.Filename, data, contentType)
	if err != nil {
		http.Error(w, "Failed to store file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"url":      randomURL.String(),
		"hash":     randomURL.RepHash,
		"size":     randomURL.FileSize,
		"filename": randomURL.FileName,
	})
}

func (s *Server) handleRetrieve(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	data, rep, err := s.rfs.RetrieveFile(hash)
	if err != nil {
		http.Error(w, "Failed to retrieve file: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", rep.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", rep.FileName))
	w.Header().Set("Content-Length", strconv.FormatInt(rep.FileSize, 10))
	w.Write(data)
}

func (s *Server) handleRandomURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	encodedURL := vars["encodedURL"]

	// Decode the URL
	decodedURL, err := base64.URLEncoding.DecodeString(encodedURL)
	if err != nil {
		http.Error(w, "Invalid encoded URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	randomURL, err := randomfs.ParseRandomURL(string(decodedURL))
	if err != nil {
		http.Error(w, "Invalid rd:// URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	data, rep, err := s.rfs.RetrieveFile(randomURL.RepHash)
	if err != nil {
		http.Error(w, "Failed to retrieve file: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", rep.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", rep.FileName))
	w.Header().Set("Content-Length", strconv.FormatInt(rep.FileSize, 10))
	w.Write(data)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := s.rfs.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"stats":   stats,
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"service": "randomfs-server",
		"version": randomfs.ProtocolVersion,
	})
}

func main() {
	flag.Parse()

	log.Printf("Starting RandomFS HTTP Server")
	log.Printf("IPFS API: %s", *ipfsAPI)
	log.Printf("Data Dir: %s", *dataDir)
	log.Printf("Cache Size: %d bytes", *cacheSize)
	log.Printf("Web Dir: %s", *webDir)
	log.Printf("HTTP Port: %d", *httpPort)

	// Create RandomFS instance
	rfs, err := randomfs.NewRandomFS(*ipfsAPI, *dataDir, *cacheSize)
	if err != nil {
		log.Fatalf("Failed to initialize RandomFS: %v", err)
	}

	// Create server
	server := NewServer(rfs)

	// Setup router
	router := mux.NewRouter()

	// API endpoints
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/store", server.handleStore).Methods("POST")
	api.HandleFunc("/retrieve/{hash}", server.handleRetrieve).Methods("GET")
	api.HandleFunc("/stats", server.handleStats).Methods("GET")
	api.HandleFunc("/health", server.handleHealth).Methods("GET")

	// rd:// URL handler
	router.HandleFunc("/rd/{encodedURL:.*}", server.handleRandomURL).Methods("GET")

	// Serve web interface
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(*webDir + "/")))

	// CORS middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Start server
	addr := fmt.Sprintf(":%d", *httpPort)
	log.Printf("Server starting on http://localhost%s", addr)
	log.Printf("API endpoints:")
	log.Printf("  POST /api/v1/store       - Store a file")
	log.Printf("  GET  /api/v1/retrieve/{hash} - Retrieve a file")
	log.Printf("  GET  /api/v1/stats       - Get system stats")
	log.Printf("  GET  /api/v1/health      - Health check")
	log.Printf("  GET  /rd/{encoded-url}   - Access via rd:// URL")

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
