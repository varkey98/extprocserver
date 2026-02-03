FROM golang:1.25
LABEL authors="varkeychanjacob"

COPY . /app
WORKDIR /app

EXPOSE 5441
EXPOSE 5442

RUN go build -o /app.o
CMD ["/app.o"]