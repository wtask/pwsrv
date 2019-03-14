# :latest based on Debian
# inspired by:
# https://pierreprinetti.com/blog/2018-the-go-1.11-web-service-dockerfile/
# http://blog.wrouesnel.com/articles/Totally%20static%20Go%20builds/
#
# RUN MySQL for the first time before server app:
# 	docker run --name mysql-server --detach -e MYSQL_ROOT_PASSWORD=rootpwd -e MYSQL_DATABASE=pwsrv -e MYSQL_USER=pwsrv -e MYSQL_PASSWORD=pwsrv mysql --default-authentication-plugin=mysql_native_password
# Check MySQL has started (Ctrl+C to stop following logs): 
# 	docker logs --follow mysql-server
# Stop MySQL:
# 	docker container stop mysql-server
# Start prepared container:
# 	docker container start mysql-server
# Build your own pwsrv image or pull it from hub:
#	docker build -t wtask/pwsrv .
#	or
# 	docker pull wtask/pwsrv (~12 Mb)
# RUN pwsrv interactively (image will be pulled if you are not built it):
# 	docker run -it --rm --name pwsrv -p 8000:8000 --link mysql-server wtask/pwsrv
# After that you can use pwsrv-API on localhost:8000.
# Stop the server by Ctrl+C
FROM golang:latest as builder

ENV \
# app config env
APP_PORT="8000" \
APP_STORAGE_DSN="mysql://pwsrv:pwsrv@tcp(mysql-server:3306)/pwsrv" \
APP_STORAGE_CONNECT_TIMEOUT="3m" \
APP_SECRET_USER_PASSWORD="user_password_secret" \
APP_SECRET_AUTH_BEARER="auth_bearer_secret" \
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

EXPOSE 8000

#STOPSIGNAL SIGTERM

ENTRYPOINT [ "./pwsrv" ]
CMD [ "-config=/app/pwsrv.config.json" ]