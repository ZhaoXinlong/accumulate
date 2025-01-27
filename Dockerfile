FROM golang:1.17 as build

# Build
WORKDIR /root
COPY . .
RUN CGO_ENABLED=0 make

FROM alpine:3

# Copy binary
COPY --from=build /root/accumulated /bin

WORKDIR /root

ENTRYPOINT ["accumulated"]
CMD ["run"]