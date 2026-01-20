package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Mock stream server for port 6501
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Mock Stream - Port 6501</title>
    <style>
        body {
            margin: 0;
            background: linear-gradient(45deg, #667eea 0%%, #764ba2 100%%);
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            font-family: Arial, sans-serif;
            color: white;
        }
        .stream-container {
            text-align: center;
            padding: 40px;
            background: rgba(0,0,0,0.3);
            border-radius: 20px;
        }
        .stream-title {
            font-size: 48px;
            font-weight: bold;
            margin-bottom: 20px;
        }
        .stream-info {
            font-size: 24px;
            margin: 10px 0;
        }
        .pulse {
            animation: pulse 2s infinite;
        }
        @keyframes pulse {
            0%%, 100%% { opacity: 1; }
            50%% { opacity: 0.5; }
        }
        .time {
            font-size: 36px;
            font-family: monospace;
            margin-top: 30px;
        }
    </style>
    <script>
        function updateTime() {
            const now = new Date();
            document.getElementById('time').textContent = now.toLocaleTimeString();
        }
        setInterval(updateTime, 1000);
        updateTime();
    </script>
</head>
<body>
    <div class="stream-container">
        <div class="stream-title pulse">📹 LIVE STREAM</div>
        <div class="stream-info">Port 6501 - VM001</div>
        <div class="stream-info">Status: <span style="color: #00ff00;">● Active</span></div>
        <div class="time" id="time"></div>
        <div style="margin-top: 20px; font-size: 18px; opacity: 0.8;">
            Mock video stream for testing
        </div>
    </div>
</body>
</html>
			`)
		})
		log.Println("Mock stream server started on port 6501")
		if err := http.ListenAndServe(":6501", nil); err != nil {
			log.Printf("Error starting server on 6501: %v", err)
		}
	}()

	// Mock stream server for port 6503
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Mock Stream - Port 6503</title>
    <style>
        body {
            margin: 0;
            background: linear-gradient(45deg, #f093fb 0%%, #f5576c 100%%);
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            font-family: Arial, sans-serif;
            color: white;
        }
        .stream-container {
            text-align: center;
            padding: 40px;
            background: rgba(0,0,0,0.3);
            border-radius: 20px;
        }
        .stream-title {
            font-size: 48px;
            font-weight: bold;
            margin-bottom: 20px;
        }
        .stream-info {
            font-size: 24px;
            margin: 10px 0;
        }
        .pulse {
            animation: pulse 2s infinite;
        }
        @keyframes pulse {
            0%%, 100%% { opacity: 1; }
            50%% { opacity: 0.5; }
        }
        .time {
            font-size: 36px;
            font-family: monospace;
            margin-top: 30px;
        }
    </style>
    <script>
        function updateTime() {
            const now = new Date();
            document.getElementById('time').textContent = now.toLocaleTimeString();
        }
        setInterval(updateTime, 1000);
        updateTime();
    </script>
</head>
<body>
    <div class="stream-container">
        <div class="stream-title pulse">📹 LIVE STREAM</div>
        <div class="stream-info">Port 6503 - VM002</div>
        <div class="stream-info">Status: <span style="color: #00ff00;">● Active</span></div>
        <div class="time" id="time"></div>
        <div style="margin-top: 20px; font-size: 18px; opacity: 0.8;">
            Mock video stream for testing
        </div>
    </div>
</body>
</html>
			`)
		})
		log.Println("Mock stream server started on port 6503")
		if err := http.ListenAndServe(":6503", mux); err != nil {
			log.Printf("Error starting server on 6503: %v", err)
		}
	}()

	// Mock stream server for port 6502
	go func() {
		mux2 := http.NewServeMux()
		mux2.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Mock Stream - Port 6502</title>
    <style>
        body {
            margin: 0;
            background: linear-gradient(45deg, #4facfe 0%%, #00f2fe 100%%);
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            font-family: Arial, sans-serif;
            color: white;
        }
        .stream-container {
            text-align: center;
            padding: 40px;
            background: rgba(0,0,0,0.3);
            border-radius: 20px;
        }
        .stream-title {
            font-size: 48px;
            font-weight: bold;
            margin-bottom: 20px;
        }
        .stream-info {
            font-size: 24px;
            margin: 10px 0;
        }
        .pulse {
            animation: pulse 2s infinite;
        }
        @keyframes pulse {
            0%%, 100%% { opacity: 1; }
            50%% { opacity: 0.5; }
        }
        .time {
            font-size: 36px;
            font-family: monospace;
            margin-top: 30px;
        }
    </style>
    <script>
        function updateTime() {
            const now = new Date();
            document.getElementById('time').textContent = now.toLocaleTimeString();
        }
        setInterval(updateTime, 1000);
        updateTime();
    </script>
</head>
<body>
    <div class="stream-container">
        <div class="stream-title pulse">📹 LIVE STREAM</div>
        <div class="stream-info">Port 6502 - VM001 (Camera 2)</div>
        <div class="stream-info">Status: <span style="color: #00ff00;">● Active</span></div>
        <div class="time" id="time"></div>
        <div style="margin-top: 20px; font-size: 18px; opacity: 0.8;">
            Mock video stream for testing
        </div>
    </div>
</body>
</html>
			`)
		})
		log.Println("Mock stream server started on port 6502")
		if err := http.ListenAndServe(":6502", mux2); err != nil {
			log.Printf("Error starting server on 6502: %v", err)
		}
	}()

	// Keep the program running
	log.Println("Mock stream servers running. Press Ctrl+C to stop.")
	log.Println("Access streams at:")
	log.Println("  - http://localhost:6501")
	log.Println("  - http://localhost:6502")
	log.Println("  - http://localhost:6503")
	
	select {}
}
