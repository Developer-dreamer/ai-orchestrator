# AI-orchestration system

## Deployment

Docker deployment files located in [deployment/docker](deployment/docker) folder. 
To run the compose file you are supposed to create `.env` file in the root directory of the project, with the
next variables:
```dotenv
PORT=8080
REDIS_URI=redis:6379
CACHE_TTL_MINUTES=5
REDIS_STREAM_ID=tasks
NUMBER_OF_WORKERS=5
```
After this run: `docker compose -f deployment/docker/docker-compose.yml up --build -d`

To check whether everything is correct check the logs of the container inside docker application,
the app should start with appropriate message. 
Make `GET` request via **Postman** or any other tool you like to the `http://localhost:8080/health` endpoint. 
If response **200** everything is fine.

## Redis integration
Redis used as a message broker between microservices. I have chose it, because of its easy integration to project, compared to RabbitMQ. Also it is already
used for caching, so there is no sense at this stage to integrate another complicated service. Moreover, Redis Streams, which were used in this implementation, 
gives you in this concrete case faster processing than basic RabbitMQ. So consider all these aspects, I decided to not overengineer my project and stand by Redis Stream.

Here is the underlying architecture of my approach:
![](doc/diagram/microservice_communication.png)