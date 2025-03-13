# Caution Companion

## Description
A backend service for an emergency alert app that notifies users in the vicinity when an emergency occurs. It leverages real-time location data to broadcast alerts, allowing users to view the exact location of incidents on a map.

## Getting Started

### Prerequisites

- Go 1.22 or higher
- Docker (for containerization)
- Make (for using the Makefile)


## Installation

* Clone this repo

  ```bash
   git clone https://github.com/olad5/caution-companion.git
   cd caution-companion
  ```
* Insall dependencies

  ```bash
   go mod download
  ```


## Build and Start the Server

```bash
    go build -v cmd/main.go  && ./main
```


## Run tests

```bash
  make test.verbose
```

### Swagger docs
http://localhost:8000/docs/index.html

![](./public/uploads/sal-er-diagram.png)

##  Resources that were helpful 
https://betterstack.com/community/guides/scaling-go/dockerize-golang/

### Built with
#### Backend

- [Golang](https://www.go.dev/) - Golang
- [Chi Router](https://go-chi.io/) - Router
- [Postgresql](https://www.postgresql.org/) - Database
- [Redis](https://redis.io/) - For holding onto JWT tokens and refresh tokens
- [SMTP express](https://smtpexpress.com/) - Mailing service
