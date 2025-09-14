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

#Install chromium
RUN apt-get update && apt-get install -y chromium chromium-driver

# Install runtime dependencies
RUN apk --update add \
    ca-certificates \
    chromium \
    tzdata \
    && \
    update-ca-certificates

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


USER appuser

# What the container should run when it is started
ENTRYPOINT [ "/bin/server" ]
