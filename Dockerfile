# syntax=docker/dockerfile:1
# Brawlhalla's own online stack (Blue Mammoth user server) for the Switch build.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build
ARG TARGETOS TARGETARCH
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -buildvcs=false -ldflags="-s -w" -o /out/brawlhallad ./cmd/brawlhallad

FROM gcr.io/distroless/static
COPY --from=build /out/brawlhallad /brawlhallad
USER nonroot:nonroot
EXPOSE 22690 22692
ENTRYPOINT ["/brawlhallad"]
