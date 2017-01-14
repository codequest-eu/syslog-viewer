FROM golang:1.7.4
RUN go get github.com/julienschmidt/httprouter  \
           gopkg.in/mcuadros/go-syslog.v2       \
           golang.org/x/net/websocket
ENV APP_HOME=/syslog
WORKDIR $APP_HOME
ADD syslog.go $APP_HOME/
RUN go build
ADD frontend $APP_HOME/frontend
EXPOSE 4321/udp
CMD '/syslog/syslog'
