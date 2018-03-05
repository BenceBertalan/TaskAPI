A TaskManager API written in GO for testing/learning purpose.

It will accept, run, and store tasks (OS commands/apps) and their results.

All the Tasks are stored as Documents (objects) in a configured MongoDB database.

This app creates a webservice which exposes the following API endpoints:

- /api/tasks (GET) - get all Tasksk from the App
- /api/tasks/{id} (GET) - get the Task with id {id} from the App 
- /api/tasks (POST) - Post a new Task

The tasks have the following required JSON fields:

- Name: The name of the task
- Command: The command which the user wants to run

The app can be run standalone or using Docker Compose.

Standalone:
- For standalone run, a configured and running MongoDB instance is needed (authentication is not supported)
- Then, the config.json should be edited to have the hostname of the MongoDB server
- Finally, You have to start the API, AND the TaskAPIWorker aswell to process the requests. (The Worker has to be started with the apiendpoint parameter in order to be able to query the Tasks)

Docker-Compose method:

For Docker Compose method, You'll need a docker container host with docker-compose tool installed
- Clone this repo using:
`git clone https://github.com/BenceBertalan/TaskAPI.git` 
- CD to the TaskAPI directory, then use this command to run the app:
`docker-compose up -d`

In order to configure the listening port of the app and the DB name to which the app has to connect, 
the following parameters are taken into account:
- Enviroment Variables: APP_PORT and APP_DB_HOSTNAME (mainly used for Docker container dynamic config)
- config.json: APP_PORT and APP_DB_HOSTNAME parameters are used

If no config method can be used for any reasons, the app will panic.