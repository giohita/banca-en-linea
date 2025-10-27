# TigerBeetle Dockerfile
FROM ubuntu:22.04

# Install dependencies
RUN apt-get update && apt-get install -y \
    curl \
    unzip \
    netcat \
    && rm -rf /var/lib/apt/lists/*

# Download and install TigerBeetle
RUN curl -Lo tigerbeetle.zip https://linux.tigerbeetle.com && \
    unzip tigerbeetle.zip && \
    chmod +x tigerbeetle && \
    mv tigerbeetle /usr/local/bin/ && \
    rm tigerbeetle.zip

# Create data directory
RUN mkdir -p /data

# Copiar script de inicialización centralizado
# Nota: El contexto de construcción debe ser la raíz del proyecto
COPY scripts/init-db.sh /usr/local/bin/init-db.sh
RUN chmod +x /usr/local/bin/init-db.sh

# Create non-root user
RUN groupadd -r tigerbeetle && useradd -r -g tigerbeetle tigerbeetle
RUN chown -R tigerbeetle:tigerbeetle /data

# Switch to non-root user
USER tigerbeetle

# Set working directory
WORKDIR /data

# Environment variables for Docker compatibility
ENV TIGERBEETLE_IO_MODE=blocking
ENV TIGERBEETLE_DISABLE_IO_URING=1

# Expose TigerBeetle port
EXPOSE 3002

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD nc -z localhost 3002 || exit 1

# Comando por defecto
CMD ["/usr/local/bin/init-db.sh"]