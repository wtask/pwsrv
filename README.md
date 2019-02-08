# pwsrv

Demo http-server to support API backend for Parrot Wings web application.

## System requirements

Go 1.11 (amd64), MySQL 5.7 (64-bit) or MariaDB (equivalent version, 64-bit).

## Dependencies

All used packages are managed by `dep` and included in this repository:

```
PROJECT                         CONSTRAINT     VERSION        REVISION  LATEST   PKGS USED
github.com/go-sql-driver/mysql  v1.4.1         v1.4.1         72cd26f   v1.4.1   1
github.com/gorilla/mux          v1.7           v1.7           a796238   v1.7.0   1
github.com/jinzhu/gorm          ^1.9.2         v1.9.2         472c70c   v1.9.2   2
github.com/jinzhu/inflection    branch master  branch master  0414036   0414036  1
golang.org/x/net                branch master  branch master  65e2d4e   65e2d4e  1
golang.org/x/text               v0.3.0         v0.3.0         f21a4df   v0.3.0   14
google.golang.org/appengine     v1.4.0         v1.4.0         e9657d8   v1.4.0   1
```

## Installing

Project IS NOT dockerized. Only manual installation is possible:

* Install Go and set up standard golang environment (under {GOPATH})
* Install and run DB server, create database and user to use with app
* Clone or download this repository into local folder under `{GOPATH}/src/...` (project root): `{GOPATH}/src/github.com/wtask/pwsrv`
* Copy example config `pwsrv.config.dist.json` from project root into any dir, for example `{GOPATH}\etc\pwsrv\`. If you have a plan to support several environments, make copy of config per-environment: `pwsrv.config.dev.json`.
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

## API

The server supports calling its methods in REST-style via http.

See more in [API documentation](https://documenter.getpostman.com/view/6496185/Rztpq7Wy)
