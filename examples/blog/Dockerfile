FROM alpine:latest

RUN apk --no-cache add \
        ca-certificates \
        coreutils \
        tzdata \
        && \
        update-ca-certificates

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

# Copy the executable from the "build" stage of Github Action.
COPY  ./bin/console /app/bin/
# Copy environment variables
COPY .env .env
#COPY .env.test .env.test
COPY .env.dev .env.dev
COPY dotenvx dotenvx
RUN chown -R appuser:appuser /app

USER appuser

WORKDIR /app

RUN ls -la /app

# Expose the port that the application listens on.
EXPOSE 8001

# What the container should run when it is started.
ENTRYPOINT [ "/app/dotenvx", "run", "-f", ".env.dev","--", "/app/bin/console", "serve" ]
