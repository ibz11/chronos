# chronos
A tool to keep track of your server's time and  health


## Build the docker image
``docker build -t chronos .``


## Run the container
``
docker run -d \
  --name chronos \
  --restart unless-stopped \
  --env-file .env \
  -p 8080:8080 \
  chronos
  ``
## Stopping the container
