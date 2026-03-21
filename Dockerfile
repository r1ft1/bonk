# ---- Build stage ----
FROM node:22-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .

# PUBLIC_ env vars are inlined at build time — pass them as build args in Coolify
ARG PUBLIC_SERVER_HTTP_URL
ARG PUBLIC_SERVER_WS_URL
ENV PUBLIC_SERVER_HTTP_URL=$PUBLIC_SERVER_HTTP_URL
ENV PUBLIC_SERVER_WS_URL=$PUBLIC_SERVER_WS_URL

RUN npm run build

# ---- Runtime stage ----
FROM node:22-alpine

WORKDIR /app

COPY --from=builder /app/build ./build
COPY --from=builder /app/package*.json ./

RUN npm ci --omit=dev

EXPOSE 3000

CMD ["node", "build/index.js"]
