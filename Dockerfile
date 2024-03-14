# global ars
ARG ALPINE_VERSION=3.18
ARG GOLANG_VERSION=1.22

#### GLOBAL #######
###################

# global build base
FROM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION} AS build-base
WORKDIR /app
ENV GOPRIVATE="github.com/carbondmp/*,github.com/brideclick/*"
ARG GITHUB_ACCESS_TOKEN=""
RUN set -eux; \
    apk add git build-base; \
    if [ ! -z "$GITHUB_ACCESS_TOKEN" ]; then\
        git config --global url."https://${GITHUB_ACCESS_TOKEN}@github.com/".insteadOf "https://github.com/"; \
    fi;

# copy external dependencies seporately to allow caching
COPY go.mod go.sum ./
RUN go mod download
COPY . ./

# global exec base
FROM alpine:${ALPINE_VERSION} AS exec-base
RUN apk add curl


#### SERVICES #####
###################

# build-api
FROM build-base AS build-api
RUN cd api && go build -o /app/changelog-api
# api
FROM exec-base AS api
COPY api/config ./config
COPY --from=build-api /app/changelog-api .

EXPOSE 12345
HEALTHCHECK --interval=2s --timeout=2s --start-period=2s --retries=5 CMD curl --fail http://localhost:12345/ping || exit 1
CMD ["./changelog-api"]

# build-emitter
FROM build-base AS build-emitter
RUN cd emitter && go build -o /app/emitter
# emitter
FROM exec-base AS emitter
COPY emitter/config ./config
COPY --from=build-emitter /app/emitter .

EXPOSE 54321
HEALTHCHECK --interval=2s --timeout=2s --start-period=2s --retries=5 CMD curl --fail http://localhost:54321/ping || exit 1
CMD ["./emitter"]

# build-consumer
FROM build-base AS build-consumer
RUN cd consumer && go build -o /app/changelog-consumer
# consumer
FROM exec-base AS consumer
COPY consumer/config ./config
COPY --from=build-consumer /app/changelog-consumer .

CMD ["./changelog-consumer"]