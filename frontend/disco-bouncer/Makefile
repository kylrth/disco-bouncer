all: build-debug

.PHONY: build-debug
build-debug: Dockerfile package*.json public/ src/
	docker build -t kylrth/disco-bouncer-web .

.PHONY: run-debug
run-debug: build-debug
	docker run -p 3000:3000 -it kylrth/disco-bouncer-web
