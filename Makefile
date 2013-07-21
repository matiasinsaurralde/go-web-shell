all: web-shell

arm: arm-web-shell

web-shell:	web-shell.go	
		GOOS=linux GOARCH=386 go build web-shell.go

arm-web-shell:	web-shell.go
		GOOS=linux GOARCH=arm go build web-shell.go
