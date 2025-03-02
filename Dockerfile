FROM golang:alpine AS build

RUN apk add --no-cache gcc musl-dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go .

ENV CGO_ENABLED=1
RUN go build -tags goolm -o product-price-scraper

# -----------------------
FROM alpine

COPY --from=build /src/product-price-scraper /bin/

ENTRYPOINT ["/bin/product-price-scraper"]
