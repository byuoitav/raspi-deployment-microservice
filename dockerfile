FROM byuoitav/amd64-alpine
MAINTAINER Daniel Randall <danny_randall@byu.edu>

ARG NAME
ENV name=${NAME}


RUN sed -i -e 's/v3\.7/v3.8/g' /etc/apk/repositories
RUN apk update
RUN apk add imagemagick

COPY ${name}-bin ${name}-bin 
COPY version.txt version.txt

# add any required files/folders here
run mkdir /public
COPY /public /public

ENTRYPOINT ./${name}-bin
