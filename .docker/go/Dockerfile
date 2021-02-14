# Собираем в гошке
FROM golang:1.15.2 as build

ENV BIN_FILE /opt/gomigrate
ENV CODE_DIR /home/app

WORKDIR ${CODE_DIR}

# Кэшируем слои с модулями
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ${CODE_DIR}

# Собираем статический бинарник Go (без зависимостей на Си API),
# иначе он не будет работать в alpine образе.
ARG LDFLAGS
RUN CGO_ENABLED=0 go build \
        -ldflags "$LDFLAGS" \
        -o ${BIN_FILE} cmd/gomigrate/*

# На выходе тонкий образ
FROM alpine:3.9

ENV BIN_FILE "/opt/gomigrate"
COPY --from=build ${BIN_FILE} ${BIN_FILE}

ENV CONFIG_FILE ${CODE_DIR}/examples/gomigrate.yaml
COPY ${CONFIG_FILE} configs/gomigrate.yaml
#
#CMD ${BIN_FILE} -config=${CONFIG_FILE}