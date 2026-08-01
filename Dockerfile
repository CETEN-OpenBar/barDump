# First stage: Build Svelte App
FROM node:20-alpine AS svelte-builder

WORKDIR /app

# Copy only package files first to cache them
COPY wrapped-app/package*.json /app/wrapped-app/
WORKDIR /app/wrapped-app
# Install dependencies
RUN npm install

# Then copy the rest of the app source
COPY wrapped-app /app/wrapped-app
RUN npm run build

# Second stage: Build Go API
FROM golang:alpine AS go-builder

WORKDIR /app/go-api
# Copy only dependencies first to cache them
COPY backend/go-api/go.mod backend/go-api/go.sum ./
RUN go mod download

# Then copy source and build
COPY backend/go-api .
RUN go build -o bardump-api ./cmd/server

# Third stage: final runtime image
FROM node:20-alpine

WORKDIR /app

# We no longer need python!

# Copy the built Svelte app and necessary Node modules
COPY --from=svelte-builder /app/wrapped-app/build /app/wrapped-app/build
COPY --from=svelte-builder /app/wrapped-app/package.json /app/wrapped-app/package.json
COPY --from=svelte-builder /app/wrapped-app/node_modules /app/wrapped-app/node_modules

# Copy the built Go API
COPY --from=go-builder /app/go-api/bardump-api /app/backend/go-api/bardump-api

# Create data directories so they exist, even if not mounted
RUN mkdir -p /app/data/raw /app/data/processed /app/wrapped-app/static

ENV BODY_SIZE_LIMIT=52428800

EXPOSE 3000
EXPOSE 8080

# Set the working directory where the Node process expects to run
WORKDIR /app/wrapped-app
# Run both the Go API in the background and Node API in the foreground
CMD ["sh", "-c", "cd /app/backend/go-api && ./bardump-api & cd /app/wrapped-app && cp -an static/* build/client/ 2>/dev/null || true && exec node build/index.js"]
