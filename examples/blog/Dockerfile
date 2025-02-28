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
USER appuser

WORKDIR /app

# Copy the executable from the "build" stage of Github Action.
COPY  ./bin/console /app/bin/
RUN ls -lah /app/bin/
RUN /bin/chmod +x /app/bin/console
RUN /bin/chmod +x /app/dotenvx

# Copy environment variables
COPY .env .env
#COPY .env.test .env.test
COPY .env.dev .env.dev
COPY dotenvx dotenvx

# Expose the port that the application listens on.
EXPOSE 8001

# What the container should run when it is started.
ENTRYPOINT [ "./dotenvx", "run", "-f", ".env.dev","--", "bin/console", "serve" ]
