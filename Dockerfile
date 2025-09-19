FROM golang:1.23-bullseye AS builder
WORKDIR /src
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    ca-certificates \
    libfreetype6-dev \
    libfontconfig1-dev \
    libgl1-mesa-dev \
    libxi-dev \
    libxcursor-dev \
    libxinerama-dev \
    libxkbcommon-dev \
    libxkbcommon-x11-dev \
    libx11-dev \
    libxrandr-dev \
    libxxf86vm-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
# Build with CGO enabled since Fyne/GLFW requires native libs
RUN CGO_ENABLED=1 GOOS=linux go build -o /out/password-manager ./cmd

FROM debian:bullseye-slim
# Install minimal runtime deps required by the UI (libGL, X11, fonts)
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libfontconfig1 \
    libfreetype6 \
    libgl1-mesa-glx \
    libxi6 \
    libxcursor1 \
    libxinerama1 \
    libxkbcommon0 \
    libxkbcommon-x11-0 \
    libxrandr2 \
    libxxf86vm1 \
 && rm -rf /var/lib/apt/lists/*
COPY --from=builder /out/password-manager /usr/local/bin/password-manager
# Include minimal project files so runtime config loader can find configs
COPY --from=builder /src/go.mod /app/go.mod
COPY --from=builder /src/configs /app/configs
WORKDIR /app
ENV GO_PASSWORD_MANAGER_PROJECT_ROOT=/app
VOLUME ["/data"]
EXPOSE 39000/udp
EXPOSE 39000/tcp
ENTRYPOINT ["/usr/local/bin/password-manager"]
