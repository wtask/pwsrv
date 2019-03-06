# pwsrv

Demo http-server to support API backend for Parrot Wings web application.

## System requirements

Go 1.11 (amd64), MySQL 5.7 (64-bit) or MariaDB (equivalent version, 64-bit).

## Dependencies

All used packages are managed by `go mod` (and also redundantly by `dep`) and included in this repository. It should not require to use `go get` for building, installing or running server. Using of `dep` is likely to be discontinued in the future, but this is not certain.

## Local installation

Manual installation steps:

* Install Go 1.11 or above and set up standard golang environment.
* Install and run DB server, create database and user to use with app
* Clone or download this repository:
	- into local folder __under__ `{GOPATH}`: `{GOPATH}/src/github.com/wtask/pwsrv`
	- __or__ into any local folder __outside__ `{GOPATH}`
* Copy example config `pwsrv.config.default.json` from project root into any dir, for example `{GOPATH}\etc\pwsrv\`. If you have a plan to support several environments, make copy of config per-environment: `pwsrv.config.dev.json`.
* Modify the configuration file according with the your local settings.

## Running

Your may install server into `{GOPATH}\bin` or build it (`go install` or `go build`) or simply run it quickly from project root in console:

```
{project root}/>go run . -config={config location}
```

Also, you can set `PWSRV_CONFIG` environment variable to hold config location. In that case, you do not need use `-config` option, but it has higher priority.

To stop server press `Ctrl+C`.

## Testing

Not all of project code is covered by tests yet. But some tests are ready. Run testing under project root:

```
{project root}/>go test ./...
```

## Database

When the server has started and after successfully connecting to the database, it checks the necessary tables and creates them if they are missing. All used tables have prefix `pwsrv_`.

## Run with Docker

* RUN MySQL container (or MariaDB if you prefer it) for the first time (slow):
	`docker run --name mysql-server -d -e MYSQL_ROOT_PASSWORD=rootpwd -e MYSQL_DATABASE=pwsrv -e MYSQL_USER=pwsrv -e MYSQL_PASSWORD=pwsrv mysql --default-authentication-plugin=mysql_native_password`

	> If you need to connect your local MySQL-client with this container add `-p 3306:3306` option.
	
* Check DB server is ready to listen on default (3306) port (Ctrl+C to stop following logs):
	`docker logs --follow mysql-server`
* Stop DB server:
	`docker stop mysql-server` or `docker container stop mysql-server`
* Start DB server again (fast):
	`docker start mysql-server` or `docker container start mysql-server`
* Build your own pwsrv-image (if you have pull down this repo):
	`docker build -t wtask/pwsrv .`
* Or simply pull server image from docker hub (~ 12 Mb):
	`docker pull wtask/pwsrv`
* Start server on your localhost:8000 (DB server must started already):
	`docker run -it --rm --name pwsrv -p 8000:8000 --link mysql-server wtask/pwsrv`
* Press Ctrl+C to stop server (and container)

## API

The server supports calling its methods in REST-style via http.

See more in [API documentation](https://documenter.getpostman.com/view/6496185/Rztpq7Wy)
