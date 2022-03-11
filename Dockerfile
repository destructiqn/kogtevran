FROM golang:latest
WORKDIR /home/kogtevran
ADD messages .
EXPOSE 25565
EXPOSE 8080
CMD [ "./kogtevran" ]