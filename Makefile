Cmd = github.com/electricface/my-go-gir-generator/cmd

cmds: cmd-girgen cmd-trial cmd-missing cmd-write-debian-files cmd-test-one-func cmd-try-error

cmd-girgen:
	go build -i -v $(Cmd)/girgen

cmd-trial:
	go build -i -v $(Cmd)/trial

cmd-missing:
	go build -i -v $(Cmd)/missing

cmd-write-debian-files:
	go build -i -v $(Cmd)/write-debian-files

cmd-test-one-func:
	go build -i -v $(Cmd)/test-one-func

cmd-try-error:
	go build -i -v $(Cmd)/try-error
