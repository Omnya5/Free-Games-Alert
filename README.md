# Free-Games-Alert

Free Games Alert is a simple application which periodically check if a new game from Epic Game Store became free. 

## Table of contents
* [General info](#general-info)
* [Technologies](#technologies)
* [Setup](#setup)

### General info
Each week, Epic Games Store makes one non free-to-play game available for free. Information about which game is free this week is posted on: https://www.epicgames.com/store/en-US/free-games.

The Free-Game-Alert check which one is discounted this way and send a notification as a Slack message to inform the user about it.

### Technologies
Application use the following technologies:
- Go 1.13
- MySQL 5.7.22
- phpMyAdmin
- Docker

### Setup
Project can be started in command line. As a argument user should provide webhool url from Slack API setup.

First we need to start docker containers:

```
$ cd project-folder
$ docker-compose up
```

Next go to golang conteiner and build application:

```
$ docker exec -it golang-test-container bash
$ cd src
$ go build main.go
$ ./main https://hooks.slack.com/webhook-url-from-slack-api-setup
```

Author: Paulina Strzygowska.
