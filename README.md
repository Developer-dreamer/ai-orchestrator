# AI-orchestration system

## Deployment

Docker deployment files located in [deployment/docker](deployment/docker) folder. 
To run the compose file you are supposed to create `.env` file in the root directory of the project, with the
next variables:
```dotenv
PORT=8080
```
After this run: `docker compose -f deployment/docker/docker-compose.yml up --build -d`

To check whether everything is correct check the logs of the container inside docker application,
the app should start with appropriate message. 
Make `GET` request via **Postman** or any other tool you like to the `http://localhost:8080/health` endpoint. 
If response **200** everything is fine.