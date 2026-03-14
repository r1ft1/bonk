# ---- Build stage ----
FROM node:22-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .

# VITE_ env vars are inlined at build time — pass them as build args in Coolify
ARG VITE_SERVER_HTTP_URL
ARG VITE_SERVER_WS_URL
ENV VITE_SERVER_HTTP_URL=$VITE_SERVER_HTTP_URL
ENV VITE_SERVER_WS_URL=$VITE_SERVER_WS_URL

RUN npm run build

# ---- Runtime stage ----
FROM node:22-alpine

WORKDIR /app

COPY --from=builder /app/build ./build
COPY --from=builder /app/package*.json ./

RUN npm ci --omit=dev

EXPOSE 3000

CMD ["node", "build/index.js"]
