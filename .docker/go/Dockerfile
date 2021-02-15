FROM golang:alpine

ENV BIN_FILE /opt/gomigrate
RUN apk add --no-cache make curl gcc libc-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Собираем статический бинарник Go (без зависимостей на Си API),
# иначе он не будет работать в alpine образе.
ARG LDFLAGS
RUN CGO_ENABLED=0 go build \
        -ldflags "$LDFLAGS" \
        -o ${BIN_FILE} cmd/gomigrate/*

RUN ls -lh
ENV CONFIG_FILE /configs/gomigrate.yaml
COPY ./examples/gomigrate.yaml ${CONFIG_FILE}
CMD make test-integration