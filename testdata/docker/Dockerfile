FROM openjdk:8-jdk-alpine

VOLUME /tmp
ARG DEPENDENCY=target/dependency
COPY ${DEPENDENCY}/BOOT-INF/lib /app/lib
COPY ${DEPENDENCY}/META-INF /app/META-INF
COPY ${DEPENDENCY}/BOOT-INF/classes /app

RUN apk add --no-cache python3 python3-dev build-base && pip3 install awscli==1.18.1

ENTRYPOINT ["java","-cp","app:app/lib/*","hello.Application"]
