# go-dockercli
(yet another) Docker CLI to monitor services and act if needed. 

The montivation behind it came from the need to be able to verify if a rolling update within the new docker services is successfull or not.
Therefore the tool checks what a service (or all) desire to run as an Docker Image and if all tasks are running the latest one.
Furthermore it checks if all services are healthy or running into a timeout.

If all works fine, the script returns 0 - if not it returns 1, which can be picked up by the CI/CD pipeline.

The second use-case is to watch a service and all it's tasks / container, as this is not yet easily discoverable within the docker toolchain itself.

## Service

First start a simple service...
```
$ docker service create --name httpcheck -p 8080:8080 \
                        --replicas=2 qnib/httpcheck:good-60s
5pqit5ixbjiw6efzdv673ezwq
$
```

This one will be healthy after about 60s.

```
$ docker ps
CONTAINER ID        IMAGE                     COMMAND                  CREATED              STATUS                        PORTS               NAMES
7f47976347e9        qnib/httpcheck:good-60s   "/opt/qnib/supervisor"   About a minute ago   Up About a minute (healthy)   8080/tcp            httpcheck.2.dx1r72f4efeiba91wly3h1whp
6ea93ccc5b10        qnib/httpcheck:good-60s   "/opt/qnib/supervisor"   About a minute ago   Up About a minute (healthy)   8080/tcp            httpcheck.1.cscz2ofk37h2sx1i3ieni1jms
```


## watchSrv

Iterates over services and displays tasks of services.

```
$ export DOCKER_HOST=unix:///var/run/docker.sock
$ go-dockercli watchSrv --loop 0
>> Loop 2s              (2016-07-29 18:52:52.437772768 +0200 CEST)
 Name            Replicas   Image                                    Tag
 httpcheck       2          qnib/httpcheck                           fail-120s
   >> Slot    ID                          Node                      TaskState  SecSince   CntStatus       Image           Tag
   >> 2       0i49b5cdaq2abgziuh746x4o7   moby                      running    54.4       starting        qnib/httpcheck  fail-120s
   >> 1       491mqt1vzvdmxkylqcnqz5f41   moby                      running    54.5       starting        qnib/httpcheck  fail-120s
 httpcheck-v2    2          qnib/httpcheck                           good-60s
   >> Slot    ID                          Node                      TaskState  SecSince   CntStatus       Image           Tag
   >> 1       07fpmjdrtra638qyqwj05phe4   moby                      running    1655.5     healthy         qnib/httpcheck  good-60s
   >> 2       8j3wk0963llgh8gr3gsi9aflg   moby                      running    1655.5     healthy         qnib/httpcheck  good-60s


>> Logs within Loop (flushed afterwards)
$
```

## superRu

### Successful RollingUpdate

```
$ go-dockercli superRu --timeout 90 --services httpcheck
 Name            Replicas   Image                                    Tag
 httpcheck       2          qnib/httpcheck                           good-60s
   >> Slot    ID                          Node                      TaskState  SecSince   CntStatus       Image           Tag                       Updated    Faulty
   >> 1       30vffsek05w6hday3is63ceif   moby                      running    126.4      healthy         qnib/httpcheck  good-60s                  true       false
   >> 2       8ke328t4iudei7kuesgdmu7yq   moby                      running    130.4      healthy         qnib/httpcheck  good-60s                  true       false
 httpcheck-v2    2          qnib/httpcheck                           good-60s
   >> Slot    ID                          Node                      TaskState  SecSince   CntStatus       Image           Tag                       Updated    Faulty
   >> 1       07fpmjdrtra638qyqwj05phe4   moby                      running    1536.5     healthy         qnib/httpcheck  good-60s                  true       false
   >> 2       8j3wk0963llgh8gr3gsi9aflg   moby                      running    1536.5     healthy         qnib/httpcheck  good-60s                  true       false
>>> All Services are updated and healthy -> OK
```

### Failed RollingUpdate

```
$ go-dockercli superRu --timeout 10 --services httpcheck
 Name            Replicas   Image                                    Tag
 httpcheck       2          qnib/httpcheck                           fail-120s
   >> Slot    ID                          Node                      TaskState  SecSince   CntStatus       Image           Tag                       Updated    Faulty
   >> 2       0i49b5cdaq2abgziuh746x4o7   moby                      starting   11.7       starting        qnib/httpcheck  fail-120s                 true       true
   >> 1       491mqt1vzvdmxkylqcnqz5f41   moby                      running    11.7       starting        qnib/httpcheck  fail-120s                 true       true
 httpcheck-v2    2          qnib/httpcheck                           good-60s
   >> Slot    ID                          Node                      TaskState  SecSince   CntStatus       Image           Tag                       Updated    Faulty
   >> 1       07fpmjdrtra638qyqwj05phe4   moby                      running    1612.7     healthy         qnib/httpcheck  good-60s                  true       false
   >> 2       8j3wk0963llgh8gr3gsi9aflg   moby                      running    1612.7     healthy         qnib/httpcheck  good-60s                  true       false
>>> Some services are faulty (timeout reached and not healthy) -> FAIL
exit status 1
```
