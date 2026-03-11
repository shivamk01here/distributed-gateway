
# -------------------------STAGE 1: The Builder-------------------------

# We will use lightweight Go Alpine image 
FROM golang:1.21-alpine AS builder
WORKDIR /app

COPY go.mod ./
COPY . .

RUN go build -o gateway ./cmd/gateway


# -----------------------------STAGE 2: The Production Image-----------------------------
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/gateway .

EXPOSE 8080

CMD ["./gateway"]