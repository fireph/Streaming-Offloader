FROM nvidia/cuda:12.4.0-runtime-ubuntu22.04

# Install dependencies and gosu for privilege drop
RUN apt-get update && apt-get install -y \
    git build-essential pkg-config libssl-dev yasm nasm \
    libx264-dev libfdk-aac-dev autoconf automake libtool \
    bash curl wget ca-certificates \
    gosu \
  && rm -rf /var/lib/apt/lists/*

# Install NVENC headers (nv-codec-headers)
RUN git clone https://git.videolan.org/git/ffmpeg/nv-codec-headers.git /nv-codec-headers && \
    cd /nv-codec-headers && \
    make install PREFIX=/usr && \
    cd / && rm -rf /nv-codec-headers

# Build FFmpeg with NVENC support
RUN git clone https://git.ffmpeg.org/ffmpeg.git /ffmpeg && \
    cd /ffmpeg && \
    ./configure --prefix=/usr/local \
      --enable-gpl --enable-nonfree \
      --enable-libx264 --enable-libfdk-aac \
      --enable-cuda --enable-nvenc \
      --extra-cflags="-I/usr/local/cuda/include -I/usr/include/ffnvcodec" \
      --extra-ldflags="-L/usr/local/cuda/lib64" && \
    make -j$(nproc) && make install && \
    cd / && rm -rf /ffmpeg

# Install Go
RUN apt-get update && apt-get install -y golang && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY cmd ./cmd

# Include default config template
COPY default-config.yaml /app/default-config.yaml

# Build the Go binary
RUN go build -o streamer cmd/streamer/main.go

# Copy entrypoint
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["./streamer"]
