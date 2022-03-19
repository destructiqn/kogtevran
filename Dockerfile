FROM golang:latest
WORKDIR /home/kogtevran
ADD kogtevran .
ADD texteria .
EXPOSE 25565
EXPOSE 8080
ENV KV_ENVIRONMENT="production"
CMD [ "./kogtevran" ]
