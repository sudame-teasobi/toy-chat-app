FROM denoland/deno:2.5.1 AS builder

WORKDIR /app
COPY . .
RUN deno cache ./write-server/main.ts
