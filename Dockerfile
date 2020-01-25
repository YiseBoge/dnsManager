FROM golang:1.13
RUN mkdir src/dnsManager
ADD . src/dnsManager
WORKDIR src/dnsManager

RUN go get "gopkg.in/robfig/cron.v3"
RUN go get "gopkg.in/yaml.v2"
RUN go get "github.com/jinzhu/gorm"
RUN go get "github.com/jinzhu/gorm/dialects/sqlite"

RUN go build -o main .
CMD ["/go/src/dnsManager/main"]