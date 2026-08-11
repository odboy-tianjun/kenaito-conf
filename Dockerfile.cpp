# Build stage
FROM alpine:3.20 AS build

WORKDIR /src

# Install build dependencies
RUN apk add --no-cache \
    build-base \
    cmake \
    curl-dev \
    git

# Copy client library
COPY kenaito-conf-cpp-client/ ./kenaito-conf-cpp-client/

# Copy demo project
COPY kenaito-conf-cpp-client-demo/ ./kenaito-conf-cpp-client-demo/

# Build
WORKDIR /src/kenaito-conf-cpp-client-demo
RUN mkdir build && cd build && \
    cmake .. && \
    make

# Runtime stage
FROM alpine:3.20

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache libcurl libstdc++ libgcc

# Copy binary
COPY --from=build /src/kenaito-conf-cpp-client-demo/build/demo /app/demo

ENTRYPOINT ["./demo"]
