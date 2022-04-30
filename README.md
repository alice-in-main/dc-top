# dc-top

## Intro
Linux terminal based tool for monitoring local docker and docker-compose containers.

Requires docker API version 1.41

![main-help](https://user-images.githubusercontent.com/44703928/165945935-679c50d0-669f-451c-afcc-d604b556adc1.png)

## Features
* List of all the containers with their basic data (ID, Name, Image, CPU Usage, Memory Usage and State)
* Remote shell inside containers
* Print container logs with a search feature
* Launch `dc-top` with `-f` flag for docker-compose mode that allows to edit the docker-compose yaml file and send compose commands
* Inspect containers
* and more...

## docker-compose mode
This tool contains a feature that allows the user to edit the docker-compose yaml file and update the compose containes on save, similar to kubernetes.

Run `./dc-top -f <path/to/docker-compose.yaml>` and press 'v' to enter edittor.

Press `Ctrl+h` inside edittor to view controls.

![edittor-help](https://user-images.githubusercontent.com/44703928/165941771-1a742e34-d093-4db0-838c-3e1c16e1e0b1.png)

## Platforms
* Works in WSL2 & Ubuntu (Other linuxes not tested)
* Partial Windows 10 functionality (Some visual bugs)

## Credits
* https://github.com/gdamore/tcell - Library for creating terminal based applications
* https://pkg.go.dev/github.com/docker/docker - Docker client API
* https://go.uber.org/goleak - Goroutine leak detector
* https://pkg.go.dev/gopkg.in/yaml.v2 - YAML unmarshaller
* https://github.com/acarl005/stripansi - Useful library for removing ansi escapse sequences from strings

## Screenshots

An example of the container filtering feature:
![search](https://user-images.githubusercontent.com/44703928/165943008-927c78de-2aea-42d6-99df-a80c5754781c.png)

Container logs:
![logs](https://user-images.githubusercontent.com/44703928/165942597-a869de45-a6fe-4f60-a6f1-b750a04bf591.png)

Container inspect:
![inspect](https://user-images.githubusercontent.com/44703928/165942593-30b755cd-d415-4776-b66a-dad1244f5eb1.png)

docker-compose mode enabled:
![docker-compose](https://user-images.githubusercontent.com/44703928/165942589-2a2b917c-1607-4eeb-893c-58882254ff9b.png)

## Installation

### Build from source
* [Install golang](https://go.dev/doc/install)
* Clone this repository
* Go to the main directory
* Run `go build`
* Run `./dc-top`

