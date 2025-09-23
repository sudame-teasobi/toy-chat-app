FROM denoland/deno:2.5.1 AS builder

WORKDIR /app
COPY . .
RUN deno compile -A -o ./bin/read-server ./read-server/main.ts

FROM debian:stable-slim
WORKDIR /app
COPY --from=builder /app/bin/read-server ./read-server

CMD ["/app/read-server"]