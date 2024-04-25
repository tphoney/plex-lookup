FROM golang:latest as builder
 
# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source from the current directory 
COPY *.go ./
COPY amazon/*.go amazon/
COPY cinemaparadiso/*.go cinemaparadiso/
COPY cmd/*.go cmd/
COPY plex/*.go plex/
COPY types/*.go types/
COPY utils/*.go utils/
COPY web/*.go web/
COPY web/*.html web/
COPY web/static/* web/static/

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /plex-lookup

# Final stage
FROM scratch
COPY --from=builder /plex-lookup /plex-lookup
# Import the root certificate for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/plex-lookup", "web"]