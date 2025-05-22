install:
	mkdir -p ~/.local/bin
	go build -o ~/.local/bin/vllmctl

update:
	go build -o "$(shell command -v vllmctl)"
