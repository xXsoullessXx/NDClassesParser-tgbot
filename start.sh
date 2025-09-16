#!/bin/sh
export DISPLAY=:99
Xvfb :99 -screen 0 1024x768x24 -ac +extension GLX +render -noreset &
sleep 2
exec "$@"
