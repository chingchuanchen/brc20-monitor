FROM rust:1.75-bookworm as rustbuilder

WORKDIR /build
ADD . .
RUN cargo build --release

FROM golang:1.20-bookworm as gobuilder

WORKDIR /build
COPY --from=rustbuilder /build /build
RUN go build -o target/server cmd/server.go
RUN go build -o target/main api/main.go
RUN go build -o target/robot robot/robot.go

FROM debian:bookworm 

WORKDIR /opt
COPY --from=rustbuilder /build/target/release/libplatform.so /usr/lib/libplatform.so
RUN apt update && apt install -y libssl-dev ca-certificates
COPY --from=gobuilder /build/target/server .
COPY --from=gobuilder /build/target/main .
COPY --from=gobuilder /build/target/robot .

