# Build binary ----------------------------------
FROM golang:1.20-alpine AS build
WORKDIR /src
COPY . .
RUN go build -o /bin/proxy .

# Copy to prod image ----------------------------
FROM scratch
COPY --from=build /bin/proxy .
EXPOSE 8080
ENV PORT=8080
CMD ["/proxy"]
