 
FROM alpine:3.10.2

WORKDIR /code

COPY differ .

USER 1000

ENTRYPOINT [ "./differ" ]