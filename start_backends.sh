#!/bin/bash
# ─────────────────────────────────────────────────────────────
#  Round Robin Load Balancer — Start all 5 Flask backend servers
# ─────────────────────────────────────────────────────────────

set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo ""
echo "  ⚙️  Round Robin Load Balancer — Backend Launcher"
echo "  ─────────────────────────────────────────────────"

# Check Python & Flask
if ! command -v python3 &>/dev/null; then
  echo "  ✗ python3 not found. Please install Python 3."
  exit 1
fi

if ! python3 -c "import flask" &>/dev/null; then
  echo "  ⚠  Flask not found. Installing..."
  pip install -q -r requirements.txt
fi

echo "  Starting 5 backend Flask servers..."
echo ""

python3 server.py "Server-1" 5000 &
sleep 0.1
python3 server.py "Server-2" 5001 &
sleep 0.1
python3 server.py "Server-3" 5002 &
sleep 0.1
python3 server.py "Server-4" 5003 &
sleep 0.1
python3 server.py "Server-5" 5004 &

echo "  ✓ Server-1  →  http://127.0.0.1:5000"
echo "  ✓ Server-2  →  http://127.0.0.1:5001"
echo "  ✓ Server-3  →  http://127.0.0.1:5002"
echo "  ✓ Server-4  →  http://127.0.0.1:5003"
echo "  ✓ Server-5  →  http://127.0.0.1:5004"
echo ""
echo "  ─────────────────────────────────────────────────"
echo "  All backends running! Now start the load balancer:"
echo ""
echo "    go run server.go"
echo ""
echo "  Then open:  http://localhost:8000/dashboard"
echo "  ─────────────────────────────────────────────────"
echo ""

wait
