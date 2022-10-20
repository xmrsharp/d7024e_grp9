# D7024E - Kademlia - Group 9

Repository for lab work in the course Mobile and distributed computing systems (D7024E) at Luleå University of Technology (LTU)

## System requirements

This is tested on Linux systems running latest versions of golang and docker, both have to manually be installed on your system before use.

See [Build your Go image](https://docs.docker.com/language/golang/build-images/) to get started with docker using go.

If you are new to go as a language the developers have a good [Get started with Go](https://go.dev/doc/tutorial/getting-started) tutorial.

## Running using make file

To manage the system use the built in shortcut commands from the make file. Use the commands go to the root folder of the project and run

```
make <option>
```

### Build

The docker-compose file dictates the number of nodes and which version of the image you want to build.

Simply make desired changes in docker-compose.yml file and run

```
make build
```

to setup the containers.

### Run

After containers are built you are ready to go to spin up the containers on your system, run

```
make run
```

Depending on the number of nodes it may take some time for the system to initialize.

### Down

Stops containers and removes containers, networks, volumes, and images created by run command.

```
make down
```

### Stop

Stops and removes running containers.

```
make stop
```

### Remove all previously created containers (Nuclear option)

Manual script for listing and removing all docker containers created using the commands above.
If any changes are made to the name of the containers the script has to be changed accordingly.

```
sudo docker rm $(sudo docker ps -aq --filter name="d7024e") | wc -l
```

### Testing

To run tests and generate a cover.html file for more detailed breakdown of code coverage run

```
make cov
```

### Changes to make file

If something is missing that is needed for the project or simply changes want to be made it is good to check the documentation mainly for docker compose beforehand to understand them.
See [Docker compose documentation](https://docs.docker.com.xy2401.com/compose/) for
available commands.

## Contributing

This is a student project for a course in distributed systems and will not be maintained after the course completion, however feel free to create issues with suggested improvements and we will look into it when available.
