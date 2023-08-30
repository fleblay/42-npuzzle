FROM golang:1.21.0-alpine3.18

RUN apk update && apk add bash make

WORKDIR /www/back
COPY ./launch.sh .
CMD ["./launch.sh"]
