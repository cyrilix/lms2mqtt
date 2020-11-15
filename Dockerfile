FROM --platform=$BUILDPLATFORM golang:alpine AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM

WORKDIR /go/src
ADD . .

RUN GOOS=$(echo $TARGETPLATFORM | cut -f1 -d/) && \
    GOARCH=$(echo $TARGETPLATFORM | cut -f2 -d/) && \
    GOARM=$(echo $TARGETPLATFORM | cut -f3 -d/ | sed "s/v//" ) && \
    CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} GOARM=${GOARM} go build -a -mod vendor -tags netgo ./cmd/lms2mqtt/



FROM gcr.io/distroless/static

USER 1234
COPY --from=builder /go/src/lms2mqtt /go/bin/lms2mqtt
ENTRYPOINT ["/go/bin/lms2mqtt"]
