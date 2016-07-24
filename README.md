# go-dockercli
(yet anpther) Docker CLI to monitor services and act if needed

## watchSrv

Iterates over services and displays tasks of services.

```
$ export DOCKER_HOST=unix:///var/run/docker.sock
$ docker service create --name httpcheck-v2  --publish 8080:8080 --env NGINX_HTTP_PORT=8080 \
                        --replicas=3 qnib/httpcheck:good-60s
asry7kv1r8n4qv8y444jsetco
$ ./go-dockercli watchSrv --loop=1
>>> Services 		(2016-07-24T12:15:46+02:00)
Name            Replicas   Image
httpcheck-v2    3          qnib/httpcheck:good-60s
                Slot       Node                                TaskStatus           Image
                1          moby                                running              qnib/httpcheck:good-60s
                2          moby                                running              qnib/httpcheck:good-60s
                3          moby                                running              qnib/httpcheck:good-60s
$
```
