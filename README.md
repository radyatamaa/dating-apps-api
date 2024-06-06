# dating-apps-api

## Introduction
this application for dating apps api base
## Package list

Packages which use in this project

#### 👨‍💻 Full list what has been used:
1. [swag](https://github.com/swaggo/swag) Swagger for Go<br/>
2. [Beego](https://github.com/beego/beego) Framework fro Go<br/>
3. [logging](go.uber.org/zap) - logging
4. [jwt](https://github.com/golang-jwt/jwt) - JWT Token
5. [go-playground](https://github.com/go-playground) - Handling Custom Validation, and Translations
6. [gorm](https://gorm.io/) - ORM query builder database 

## Architecture Code
### Clean Architecture
This project has  4 Domain layer :

 * Models Layer
 * Repository Layer
 * Usecase Layer  
 * Delivery Layer

#### The diagram:

![golang clean architecture](https://github.com/bxcodec/go-clean-arch/raw/master/clean-arch.png)

The explanation about this project's structure clean arch can read from this medium's post : https://medium.com/@imantumorang/golang-clean-archithecture-efd6d7c43047

### Go Standar Project For Layout Pkgs
more the explanation : https://github.com/golang-standards/project-layout

## Installation
#### 1. Requirements

##### a. Golang Language SDK minimal 1.16 https://golang.org/dl/
##### b. gomod Golang Package Management https://go.dev/doc/modules/managing-dependencies
##### c. Database (MYSQL)

#### 2.Clone The Projects
```$xslt
    cd ~YOUR/GO/FOLDER/DIRECTORY
``` 
then clone the project
```$xslt
    git clone https://github.com/radyatamaa/dating-apps-api.git
```

#### 3. install all the dependencies
a. Go to the project folder
```$xslt
    cd ~YOUR/GO/FOLDER/DIRECTORY/dating-apps-api
```
b. then run `gomod tidy` command.

## app.ini (environment variable)
Make sure `gomod tidy` is successful, and then make .env file with command
```$xslt
    cp .app.conf.example .app.ini
```

Update the content of `app.ini` value, like the database name, the `initDataDummyProfileSeeder` you can enable `true` if you want to add the seeder dummy data for profile & user, 
the `redisBeegoConConfig` you can adjust as your configuration redis example with host only you can put `"{"conn":"127.0.0.1:6379"}"` and with all object setting 
`"{"key":"datingAppsAPI","conn":"127.0.0.1:6379","dbNum":"1","password":"redispassword"}"`
```$xslt
appname = dating_apps_api
appUrl = http://localhost:8082
version = 1.1.0
serverTimeout=120
executionTimeout=30
httpport = 8082
runmode = dev
autorender = false
copyrequestbody = true
EnableDocs = true
lang="en|id"
logPath="./logs/api.log"
initDataDummyProfileSeeder=true
redisBeegoConConfig="{"conn":"127.0.0.1:6379"}"

[database]
# debug=true
driver="mysql"
host="localhost"
username=root
password=
name=dating_apps
port=3306
options="charset=utf8mb4&parseTime=True&loc=Local"
maxOpenConn = 25
maxIdleConn = 25
maxLifeTimeConn = 300
maxIdleTimeConn = 300

```

### How To Run This Project in local use Docker With Redis and Mysql installation

```bash
# move to directory
cd $GOPATH/src/github.com/radyatamaa

#move to project
cd dating-apps-api

# deploy the app use docker
docker compose -f "docker-compose.yml" up -d --build

# Open at browser this url
http://localhost:8082/swagger/index.html
```

### How To Run This Project in local use Golang

```bash
# move to directory
cd $GOPATH/src/github.com/radyatamaa

#move to project
cd dating-apps-api

# Run app 
go run main.go

# if you not installed yet golang can use docker compose 
docker compose -f "docker-compose-app-only.yml" up -d --build

# Open at browser this url
http://localhost:8082/swagger/index.html
```

## Notes
this app using auto migration by `gorm` , so you dont need create table as manually or anything , you only do need to run the app then the tables will be migrated by the app

## Commands
- run unit test : go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out
- run unit test with makefile : make test
- generate mock : mockgen -source='internal/user/repository.go' MysqlRepository -destination='internal/domain/mocks/RestRepository.go' -package=mocks
- generate mock package external : mockgen -destination='internal/domain/mocks/event.go' -package=mocks github.com/KB-FMF/platform-library/event Event

### Swagger UI:
http://localhost:8082/swagger/index.html
![swagger-image](https://github.com/radyatamaa/dating-apps-api/blob/dev/swagger-image.png)

### More about app details test & additional information:
[documentation](https://github.com/radyatamaa/dating-apps-api/blob/dev/document-app.pdf)
