FROM denoland/deno:2.5.1 AS builder

WORKDIR /app
COPY . .
RUN deno compile -A -o ./bin/write-server ./write-server/main.ts

FROM debian:stable-slim
WORKDIR /app
COPY --from=builder /app/bin/write-server ./write-server

CMD ["/app/write-server"]