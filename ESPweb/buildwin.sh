#!/bin/bash
mkdir -p bin

GOOS=windows GOARCH=amd64 go build -o bin/esp32.exe .

cat > bin/web.bat << 'EOF'
explorer "http://localhost:8032"
powershell -WindowStyle Minimized ./esp32.exe -enablelog true
EOF

echo "Done: bin/esp32.exe  bin/web.bat"
