#!/bin/sh
# Chrome wrapper script to bypass crashpad handler issues

# Set environment variables
export XDG_CONFIG_HOME=/tmp/.chromium
export XDG_CACHE_HOME=/tmp/.chromium
export XDG_DATA_HOME=/tmp/.chromium
export HOME=/tmp
export CHROME_DISABLE_CRASH_REPORTER=1
export CHROME_DISABLE_CRASHPAD=1
export CHROME_NO_CRASH_UPLOAD=1
export CHROME_DISABLE_BREAKPAD=1
export CHROME_DISABLE_CRASHPAD_HANDLER=1
export DISPLAY=:99

# Create directories if they don't exist
mkdir -p /tmp/.chromium /tmp/chrome-user-data

# Log startup
echo "Starting Chrome wrapper script..."
echo "Environment variables set"
echo "Directories created"

# Check if Chrome binary exists
if [ ! -f "/usr/bin/chromium-browser" ]; then
    echo "ERROR: Chrome binary not found at /usr/bin/chromium-browser"
    exit 1
fi

echo "Chrome binary found, launching..."

# Launch Chrome with all crashpad disabling flags
exec /usr/bin/chromium-browser \
    --headless \
    --no-sandbox \
    --disable-setuid-sandbox \
    --disable-dev-shm-usage \
    --disable-accelerated-2d-canvas \
    --no-first-run \
    --no-zygote \
    --single-process \
    --disable-gpu \
    --disable-background-timer-throttling \
    --disable-backgrounding-occluded-windows \
    --disable-renderer-backgrounding \
    --disable-features=TranslateUI,VizDisplayCompositor,Crashpad \
    --disable-ipc-flooding-protection \
    --disable-extensions \
    --disable-plugins \
    --disable-images \
    --disable-web-security \
    --remote-debugging-port=0 \
    --disable-logging \
    --log-level=3 \
    --silent \
    --disable-crash-reporter \
    --disable-in-process-stack-traces \
    --disable-breakpad \
    --disable-crashpad \
    --disable-crashpad-handler \
    --disable-crashpad-handler-database \
    --disable-crashpad-handler-upload \
    --disable-crashpad-handler-reporting \
    --crashpad-handler=false \
    --no-crash-upload \
    --disable-crash-reporter-upload \
    --user-data-dir=/tmp/chrome-user-data \
    --disable-background-networking \
    --disable-default-apps \
    --disable-sync \
    --disable-translate \
    --disable-component-update \
    --disable-domain-reliability \
    "$@"
