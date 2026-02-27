package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

// ─── ANSI Colors ────────────────────────────────────────────────────────────
const (
	colorReset  = "\033[0m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorPurple = "\033[35m"
	colorBold   = "\033[1m"
)

func logInfo(format string, a ...interface{}) {
	fmt.Printf(colorCyan+"[INFO]  "+colorReset+format+"\n", a...)
}
func logRoute(format string, a ...interface{}) {
	fmt.Printf(colorGreen+"[ROUTE] "+colorReset+format+"\n", a...)
}
func logWarn(format string, a ...interface{}) {
	fmt.Printf(colorYellow+"[WARN]  "+colorReset+format+"\n", a...)
}
func logError(format string, a ...interface{}) {
	fmt.Printf(colorRed+"[ERROR] "+colorReset+format+"\n", a...)
}

// ─── Server Configuration ───────────────────────────────────────────────────
type Backend struct {
	URL      string
	Name     string
	Requests atomic.Int64
	Healthy  atomic.Bool
}

var (
	backends = []*Backend{
		{URL: "http://127.0.0.1:5000", Name: "Server-1"},
		{URL: "http://127.0.0.1:5001", Name: "Server-2"},
		{URL: "http://127.0.0.1:5002", Name: "Server-3"},
		{URL: "http://127.0.0.1:5003", Name: "Server-4"},
		{URL: "http://127.0.0.1:5004", Name: "Server-5"},
	}
	currentIndex  int
	mu            sync.Mutex
	startTime     = time.Now()
	totalRequests atomic.Int64
)

func init() {
	// Mark all backends healthy initially
	for _, b := range backends {
		b.Healthy.Store(true)
	}
}

// ─── Health Check Poller ─────────────────────────────────────────────────────
func startHealthChecker() {
	go func() {
		for {
			for _, b := range backends {
				go func(backend *Backend) {
					resp, err := http.Get(backend.URL + "/health")
					if err != nil || resp.StatusCode != 200 {
						if backend.Healthy.Load() {
							logWarn("%s went %sOFFLINE%s", backend.Name, colorRed, colorReset)
						}
						backend.Healthy.Store(false)
					} else {
						if !backend.Healthy.Load() {
							logInfo("%s is back %sONLINE%s", backend.Name, colorGreen, colorReset)
						}
						backend.Healthy.Store(true)
					}
				}(b)
			}
			time.Sleep(5 * time.Second)
		}
	}()
}

// ─── Round Robin Selector ────────────────────────────────────────────────────
func getNextBackend() *Backend {
	mu.Lock()
	defer mu.Unlock()
	total := len(backends)
	for i := 0; i < total; i++ {
		idx := currentIndex % total
		currentIndex++
		if backends[idx].Healthy.Load() {
			return backends[idx]
		}
	}
	return nil // all backends down
}

// ─── Handlers ────────────────────────────────────────────────────────────────
func forwardRequest(res http.ResponseWriter, req *http.Request) {
	backend := getNextBackend()
	if backend == nil {
		logError("All backends are down!")
		http.Error(res, `{"error":"All backends are unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	parsedURL, err := url.Parse(backend.URL)
	if err != nil {
		logError("Failed to parse backend URL: %v", err)
		http.Error(res, "Internal error", http.StatusInternalServerError)
		return
	}

	backend.Requests.Add(1)
	totalRequests.Add(1)
	logRoute("→ %s%-10s%s [Total: %d]", colorBold, backend.Name, colorReset, totalRequests.Load())

	proxy := httputil.NewSingleHostReverseProxy(parsedURL)
	proxy.ServeHTTP(res, req)
}

func metricsHandler(res http.ResponseWriter, req *http.Request) {
	type ServerStat struct {
		Name     string `json:"name"`
		URL      string `json:"url"`
		Requests int64  `json:"requests"`
		Healthy  bool   `json:"healthy"`
	}
	type Metrics struct {
		TotalRequests int64        `json:"total_requests"`
		UptimeSeconds float64      `json:"uptime_seconds"`
		Servers       []ServerStat `json:"servers"`
	}

	stats := Metrics{
		TotalRequests: totalRequests.Load(),
		UptimeSeconds: time.Since(startTime).Seconds(),
	}
	for _, b := range backends {
		stats.Servers = append(stats.Servers, ServerStat{
			Name:     b.Name,
			URL:      b.URL,
			Requests: b.Requests.Load(),
			Healthy:  b.Healthy.Load(),
		})
	}

	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(res).Encode(stats)
}

func dashboardHandler(res http.ResponseWriter, req *http.Request) {
	http.ServeFile(res, req, "static/dashboard.html")
}

func staticHandler() http.Handler {
	return http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
}

// ─── Main ────────────────────────────────────────────────────────────────────
func main() {

	fmt.Println(colorPurple + colorBold + `
  ██████╗  ██████╗ ██╗   ██╗███╗   ██╗██████╗
  ██╔══██╗██╔═══██╗██║   ██║████╗  ██║██╔══██╗
  ██████╔╝██║   ██║██║   ██║██╔██╗ ██║██║  ██║
  ██╔══██╗██║   ██║██║   ██║██║╚██╗██║██║  ██║
  ██║  ██║╚██████╔╝╚██████╔╝██║ ╚████║██████╔╝
  ╚═╝  ╚═╝ ╚═════╝  ╚═════╝ ╚═╝  ╚═══╝╚═════╝
  ██████╗  ██████╗ ██████╗ ██╗███╗   ██╗
  ██╔══██╗██╔═══██╗██╔══██╗██║████╗  ██║
  ██████╔╝██║   ██║██████╔╝██║██╔██╗ ██║
  ██╔══██╗██║   ██║██╔══██╗██║██║╚██╗██║
  ██║  ██║╚██████╔╝██████╔╝██║██║ ╚████║
  ╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚═╝╚═╝  ╚═══╝
  ██╗      ██████╗  █████╗ ██████╗
  ██║     ██╔═══██╗██╔══██╗██╔══██╗
  ██║     ██║   ██║███████║██║  ██║
  ██║     ██║   ██║██╔══██║██║  ██║
  ███████╗╚██████╔╝██║  ██║██████╔╝
  ╚══════╝ ╚═════╝ ╚═╝  ╚═╝╚═════╝
  ██████╗  █████╗ ██╗      █████╗ ███╗   ██╗ ██████╗███████╗██████╗
  ██╔══██╗██╔══██╗██║     ██╔══██╗████╗  ██║██╔════╝██╔════╝██╔══██╗
  ██████╔╝███████║██║     ███████║██╔██╗ ██║██║     █████╗  ██████╔╝
  ██╔══██╗██╔══██║██║     ██╔══██║██║╚██╗██║██║     ██╔══╝  ██╔══██╗
  ██████╔╝██║  ██║███████╗██║  ██║██║ ╚████║╚██████╗███████╗██║  ██║
  ╚═════╝ ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝╚══════╝╚═╝  ╚═╝
` + colorReset)

	logInfo("Starting with %d backend servers", len(backends))
	logInfo("Dashboard  → %shttp://localhost:8000/dashboard%s", colorCyan, colorReset)
	logInfo("Metrics    → %shttp://localhost:8000/metrics%s", colorCyan, colorReset)
	logInfo("Balancer   → %shttp://localhost:8000/%s", colorCyan, colorReset)
	fmt.Println()

	startHealthChecker()

	http.HandleFunc("/dashboard", dashboardHandler)
	http.HandleFunc("/metrics", metricsHandler)
	http.Handle("/static/", staticHandler())
	http.HandleFunc("/", forwardRequest)

	logInfo("Listening on %s:8000%s ...", colorBold, colorReset)
	if err := http.ListenAndServe(":8000", nil); err != nil {
		logError("Server failed: %v", err)
	}
}
