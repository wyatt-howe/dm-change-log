FROM alpine:3.18
ENV ENV="prod"
WORKDIR /app
COPY config/ ./config
COPY main ./
CMD ["/app/main"]
EXPOSE 8780