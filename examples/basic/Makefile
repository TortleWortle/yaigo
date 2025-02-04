dev:
	APP_ENV=local make -j2 watch-server watch-web

watch-server: generate
	air .

watch-web:
	npm --prefix web run dev

generate:
	go generate ./...

build: generate
	go build -o basic_example .

install:
	npm --prefix web install

.PHONY: watch-web watch-server dev generate build
