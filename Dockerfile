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
COPY musicbrainz/*.go musicbrainz/
COPY plex/*.go plex/
COPY spotify/*.go spotify/
COPY types/*.go types/
COPY utils/*.go utils/

COPY web/movies/*.go web/movies/
COPY web/movies/*.html web/movies/
COPY web/music/*.go web/music/
COPY web/music/*.html web/music/
COPY web/tv/*.go web/tv/
COPY web/tv/*.html web/tv/
COPY web/settings/*.go web/settings/
COPY web/settings/*.html web/settings/
COPY web/server.go web/index.html web/static/ /web/

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /plex-lookup

# Final stage
FROM scratch
COPY --from=builder /plex-lookup /plex-lookup
# Import the root certificate for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/plex-lookup", "web"]