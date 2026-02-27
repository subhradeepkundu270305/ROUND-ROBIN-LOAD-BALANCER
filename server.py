from flask import Flask, jsonify
import sys
import time

app = Flask(__name__)

if len(sys.argv) < 3:
    print("Usage: python3 server.py <server_name> <port>")
    sys.exit(1)

SERVER_NAME = sys.argv[1]
PORT = int(sys.argv[2])
START_TIME = time.time()
request_count = 0

@app.route('/')
def hello():
    global request_count
    request_count += 1
    return jsonify({
        "server": SERVER_NAME,
        "port": PORT,
        "requests_served": request_count,
        "uptime_seconds": round(time.time() - START_TIME, 2),
        "timestamp": time.strftime("%Y-%m-%dT%H:%M:%S")
    })

@app.route('/health')
def health():
    return jsonify({
        "status": "ok",
        "server": SERVER_NAME,
        "port": PORT,
        "uptime_seconds": round(time.time() - START_TIME, 2)
    }), 200

if __name__ == '__main__':
    print(f"  ✓ {SERVER_NAME} listening on port {PORT}")
    app.run(host='127.0.0.1', port=PORT, debug=False)