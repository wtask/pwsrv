# :latest based on Debian
# inspired by:
# https://pierreprinetti.com/blog/2018-the-go-1.11-web-service-dockerfile/
# http://blog.wrouesnel.com/articles/Totally%20static%20Go%20builds/
FROM golang:latest as builder

ARG app_port="8000"
ARG app_storage_user="pwsrv"
ARG app_storage_pass="pwsrv"
ARG app_storage_ip="127.0.0.1"
ARG app_dbname="pwsrv"


ENV \
APP_PORT="${app_port}"\
APP_STORAGE_TYPE="mysql"\
APP_STORAGE_USER="${app_storage_user}"\
APP_STORAGE_PASS="${app_storage_pass}"\
APP_STORAGE_IPADDR="${app_storage_ip}"\
APP_STORAGE_MYSQL_PORT="3306"\
APP_DBNAME="${app_dbname}" \
# golang env
CGO_ENABLED=0 \
GOOS=linux \
GOARCH=amd64

WORKDIR /build
COPY ./ ./

RUN \
apt-get update -y \ 
&& apt-get install -y \ 
	gettext-base \ 
&& update-ca-certificates \
&& envsubst < pwsrv.config.docker.json | tee > pwsrv.config.json \
&& go mod download \
&& go build -a -ldflags '-extldflags "-static"' -o pwsrv .

FROM scratch

WORKDIR /app
# Import the Certificate-Authority certificates for enabling HTTPS.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/pwsrv /build/pwsrv.config.json ./

EXPOSE ${app_port}

#STOPSIGNAL SIGTERM

ENTRYPOINT [ "./pwsrv" ]
CMD [ "-config=/app/pwsrv.config.json" ]