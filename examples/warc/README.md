This is a work-in-progress demo for https://cloudnativeday.ch/an-electric-automation-engine/

There is significant amount of background context in this repository: https://github.com/gerhard/cnd .
While this repository is still private, we expect to make it public sometimes in September 2022.
I intend to update this when that happens.

## Demo

- [x] `go install ./cmd/cloak`
- [x] `cloak generate`
- [x] 💥 `go run main.go`

```sh
$ go run main.go
#1 local://__cloak_workdir
#1 transferring __cloak_workdir: 24.05kB done
#1 DONE 0.0s

#2 copy / /
#2 DONE 0.1s

#3 docker-image://docker.io/library/alpine:3.15
#3 resolve docker.io/library/alpine:3.15
#3 resolve docker.io/library/alpine:3.15 0.9s done
#3 CACHED

#4 mkfile /Dockerfile
#4 DONE 0.0s

#5 mkfile /main.go
#5 CACHED

#6 [internal] load metadata for docker.io/library/golang:1.18.2-alpine
#6 DONE 0.4s

#7 [build 1/6] FROM docker.io/library/golang:1.18.2-alpine@sha256:4795c5d21f01e0777707ada02408debe77fe31848be97cf9fa8a1462da78d949
#7 resolve docker.io/library/golang:1.18.2-alpine@sha256:4795c5d21f01e0777707ada02408debe77fe31848be97cf9fa8a1462da78d949 0.0s done
#7 DONE 0.0s

#8 [build 3/6] RUN apk add --no-cache file git
#8 CACHED

#9 [stage-1 1/1] COPY --from=build /_shim /_shim
#9 CACHED

#10 [build 2/6] WORKDIR /src
#10 CACHED

#11 [build 4/6] RUN go mod init github.com/dagger/cloak/shim/cmd
#11 CACHED

#12 [build 5/6] COPY . .
#12 CACHED

#13 [build 6/6] RUN CGO_ENABLED=0 go build -o /_shim -ldflags '-s -d -w' .
#13 CACHED

#14 apk add -U --no-cache wget bash
#14 CACHED

#6 [internal] load metadata for docker.io/library/golang:1.18.2-alpine
#6 DONE 0.5s

#9 [stage-1 1/1] COPY --from=build /_shim /_shim
#9 CACHED

#15 ./run.sh
#15 0.259 + URL=https://cloudnativeday.ch/an-electric-automation-engine
#15 0.259 + wget --mirror --warc-file=cloudnative.ch --warc-cdx --page-requisites --html-extension --convert-links --execute robots=off --directory-prefix=. --span-hosts --domains=cloudnative.ch,js.tito.io --wait=1 --random-wait https://cloudnativeday.ch/an-electric-automation-engine
#15 0.260 WARC output does not work with timestamping, timestamping will be disabled.
#15 0.260 Opening WARC file 'cloudnative.ch.warc'.
#15 0.260 
#15 0.260 --2022-09-08 16:38:35--  https://cloudnativeday.ch/an-electric-automation-engine
#15 0.266 Resolving cloudnativeday.ch (cloudnativeday.ch)... 94.126.18.174
#15 0.302 Connecting to cloudnativeday.ch (cloudnativeday.ch)|94.126.18.174|:443... connected.
#15 0.372 HTTP request sent, awaiting response... 301 Moved Permanently
#15 1.188 Location: https://cloudnativeday.ch/an-electric-automation-engine/ [following]
#15 1.188 
#15 1.188      0K                                                        0.00 =0s
#15 1.188 
#15 2.104 --2022-09-08 16:38:37--  https://cloudnativeday.ch/an-electric-automation-engine/
#15 2.104 Reusing existing connection to cloudnativeday.ch:443.
#15 2.104 HTTP request sent, awaiting response... 200 OK
#15 2.312 Length: unspecified [text/html]
#15 2.312 Saving to: './cloudnativeday.ch/an-electric-automation-engine.html'
#15 2.312 
#15 2.312      0K .......... .......... .......... .......... .......... 1.42M
#15 2.347     50K .........                                              65.5M=0.03s
#15 2.347 
#15 2.347 2022-09-08 16:38:37 (1.68 MB/s) - './cloudnativeday.ch/an-electric-automation-engine.html' saved [60848]
#15 2.347 
#15 2.347 FINISHED --2022-09-08 16:38:37--
#15 2.347 Total wall clock time: 2.1s
#15 2.347 Downloaded: 1 files, 59K in 0.03s (1.68 MB/s)
#15 2.348 Converting links in ./cloudnativeday.ch/an-electric-automation-engine.html... 29.
#15 2.348 17-12
#15 2.348 Converted links in 1 files in 0.001 seconds.
#15 DONE 2.5s
panic: failed to solve: failed to run script: input:6: core.filesystem.exec.mount missing mount path


goroutine 1 [running]:
main.main()
        /Users/gerhard/github.com/gerhard/cloak/examples/warc/main.go:57 +0x65
exit status 2
```
