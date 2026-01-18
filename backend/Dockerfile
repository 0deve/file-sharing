# build
FROM golang:1.21-alpine AS builder
WORKDIR /app
# copy files
COPY main.go .
# donwload dep
RUN go mod init odv-share && go get github.com/tus/tusd/pkg/handler && go get github.com/tus/tusd/pkg/filestore && go get golang.org/x/time/rate
# compile
RUN CGO_ENABLED=0 go build -o server main.go

# small image
FROM alpine:3.19

# create non root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# folder for user
RUN mkdir uploads && chown appuser:appgroup uploads

# copy exec
COPY --from=builder /app/server .

# copy static files
COPY static ./static

# perms
RUN chown appuser:appgroup server
RUN chown -R appuser:appgroup ./static

# upload folder
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

CMD ["./server"]