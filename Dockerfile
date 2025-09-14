# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.25.0 AS build
WORKDIR /src

# Copy go.mod and go.sum files and download dependencies
# This leverages Docker's layer caching. Dependencies will be re-downloaded only if these files change.
COPY go.mod go.sum ./
RUN go mod download -x

# Copy the rest of the source code
COPY . .

# Build the application
# We use CGO_ENABLED=0 to create a statically linked binary.
ARG TARGETARCH
RUN CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /bin/server .


# Final stage
FROM alpine:latest AS final



# Update package repositories and install runtime dependencies
RUN apk update && apk add --no-cache \
    chromium \
    chromium-chromedriver \
    nss \
    freetype \
    harfbuzz \
    ca-certificates \
    ttf-freefont \
    xvfb \
    dbus \
    ttf-dejavu \
    fontconfig \
    dumb-init

# Create directories for Chrome and set permissions
RUN mkdir -p /tmp/chrome-user-data /tmp/.X11-unix /tmp/.chromium && \
    chmod 777 /tmp/chrome-user-data /tmp/.X11-unix /tmp/.chromium

# Create startup script to set up virtual display
RUN echo '#!/bin/sh' > /app/start.sh && \
    echo 'export DISPLAY=:99' >> /app/start.sh && \
    echo 'Xvfb :99 -screen 0 1024x768x24 -ac +extension GLX +render -noreset &' >> /app/start.sh && \
    echo 'sleep 2' >> /app/start.sh && \
    echo 'exec "$@"' >> /app/start.sh && \
    chmod +x /app/start.sh

# Set environment variables for Chrome
ENV DISPLAY=:99
ENV CHROME_BIN=/usr/bin/chromium-browser
ENV CHROME_PATH=/usr/bin/chromium-browser
ENV CHROME_DISABLE_CRASH_REPORTER=1
ENV CHROME_DISABLE_CRASHPAD=1
ENV CHROME_NO_CRASH_UPLOAD=1
ENV CHROME_DISABLE_BREAKPAD=1
ENV CHROME_DISABLE_CRASHPAD_HANDLER=1
ENV XDG_CONFIG_HOME=/tmp/.chromium
ENV XDG_CACHE_HOME=/tmp/.chromium
ENV XDG_DATA_HOME=/tmp/.chromium
ENV HOME=/tmp

# Create a non-privileged user
ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser

# Set working directory
WORKDIR /app

# Copy the executable from the "build" stage
COPY --from=build /bin/server /bin/

# Copy Chrome wrapper script
COPY chrome-wrapper.sh /usr/local/bin/chrome-wrapper
RUN chmod +x /usr/local/bin/chrome-wrapper

# Test Chrome installation
RUN /usr/bin/chromium-browser --version || echo "Chrome version check failed"
RUN ls -la /usr/bin/chromium* || echo "Chrome binary listing failed"


USER appuser

# What the container should run when it is started
ENTRYPOINT [ "/app/start.sh", "dumb-init", "--", "/bin/server" ]
