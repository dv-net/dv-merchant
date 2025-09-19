FROM golang:1.24-alpine AS build

WORKDIR /app

RUN apk add --no-cache curl jq tar

COPY . .

RUN go mod download

ARG FRONTEND_PROJECT_NAME=dv-frontend
ARG GITHUB_OWNER=dv-net

RUN DV_FRONTEND_TAG=$(curl -s https://api.github.com/repos/${GITHUB_OWNER}/${FRONTEND_PROJECT_NAME}/releases/latest | jq -r '.tag_name') && \
    DV_FRONTEND_ARCHIVE_FILENAME="${FRONTEND_PROJECT_NAME}.${DV_FRONTEND_TAG}.tar.gz" && \
    echo "Downloading frontend ${DV_FRONTEND_TAG} ..." && \
    curl -sSL -o ${DV_FRONTEND_ARCHIVE_FILENAME} "https://github.com/${GITHUB_OWNER}/${FRONTEND_PROJECT_NAME}/releases/download/${DV_FRONTEND_TAG}/${DV_FRONTEND_ARCHIVE_FILENAME}" && \
    mkdir -p frontend/dist && \
    tar -xf ${DV_FRONTEND_ARCHIVE_FILENAME} -C frontend/dist && \
    rm ${DV_FRONTEND_ARCHIVE_FILENAME}

RUN go build -o ./.bin/dv-merchant ./cmd/app/

FROM alpine:latest

RUN apk add --no-cache tzdata

WORKDIR /app

RUN mkdir -p /app/configs

COPY --from=build /app/.bin/dv-merchant /app/
COPY ./configs/rbac_model.conf /app/configs
COPY ./configs/rbac_policies.csv /app/configs

COPY ./artifacts/scripts/start.sh /app/start.sh
RUN chmod +x /app/start.sh

EXPOSE 80

CMD ["/app/start.sh"]
