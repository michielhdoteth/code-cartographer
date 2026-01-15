#!/usr/bin/env python3
"""
Simple web server to serve the Code Map infinite canvas
"""

import http.server
import socketserver
import os
import sys

PORT = 8080

class CodeMapHandler(http.server.SimpleHTTPRequestHandler):
    def end_headers(self):
        # Allow CORS for local development
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Access-Control-Allow-Methods', 'GET, POST, OPTIONS')
        self.send_header('Access-Control-Allow-Headers', 'Content-Type')
        super().end_headers()

    def do_GET(self):
        if self.path == '/':
            self.path = '/infinite-canvas.html'
        return super().do_GET()

def main():
    # Change to the directory containing the files
    script_dir = os.path.dirname(os.path.abspath(__file__))
    os.chdir(script_dir)

    print("[+] Starting Code Map Server")
    print(f"[-] Serving from: {script_dir}")
    print(f"[-] Open: http://localhost:{PORT}/infinite-canvas.html")
    print("[-] Interactive infinite canvas with knowledge graph visualization")
    print("[-] Press Ctrl+C to stop")

    try:
        with socketserver.TCPServer(("", PORT), CodeMapHandler) as httpd:
            httpd.serve_forever()
    except KeyboardInterrupt:
        print("\n[+] Server stopped")
        sys.exit(0)

if __name__ == "__main__":
    main()
