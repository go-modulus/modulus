ARG GO_VERSION=1.24.0
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS builder
WORKDIR /src

COPY . .

RUN go mod download -x

ARG TARGETARCH

RUN CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /bin/console ./cmd/console/main.go

FROM alpine:latest

RUN apk --no-cache add \
        ca-certificates \
        tzdata \
        && \
        update-ca-certificates

RUN apk add --no-cache curl

# Create a non-privileged user that the app will run under.
# See https://docs.docker.com/go/dockerfile-user-best-practices/
ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser
USER appuser

WORKDIR /app

# Copy the executable from the "build" stage.
COPY --from=builder /bin/console /app/bin/

# Copy environment variables
COPY .env .env
#COPY .env.test .env.test
COPY .env.dev .env.dev

# Expose the port that the application listens on.
EXPOSE 8001

RUN curl -sfS https://dotenvx.sh/install.sh?directory=. | sh
#CMD ["dotenvx", "run", "--", "/bin/console", "serve"]

# What the container should run when it is started.
ENTRYPOINT [ "./dotenvx", "run", "-f", ".env.dev","--", "bin/console", "serve" ]
#ENTRYPOINT [ "bin/console", "serve" ]