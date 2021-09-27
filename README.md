
# Dns resolver

Run the resolver:

```
$ docker run --net=host -v /var/run/docker.sock:/var/run/docker.sock ferranbt/dockeringress:latest --port 53535
```

Run **two** containers with the label 'dnsresolve':

```
$ docker run --label dnsresolve=bar.service nginx
```

Query and resolve the bar.service:

```
$ dig @localhost -p 53535 bar.service.

; <<>> DiG 9.16.1-Ubuntu <<>> @localhost -p 53535 bar.service.
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 15690
;; flags: qr rd; QUERY: 1, ANSWER: 2, AUTHORITY: 0, ADDITIONAL: 0
;; WARNING: recursion requested but not available

;; QUESTION SECTION:
;bar.service.			IN	A

;; ANSWER SECTION:
bar.service.		3600	IN	A	172.17.0.3
bar.service.		3600	IN	A	172.17.0.2

;; Query time: 0 msec
;; SERVER: 127.0.0.1#53535(127.0.0.1)
;; WHEN: lun ago 02 17:21:02 CEST 2021
;; MSG SIZE  rcvd: 83
```
