.PHONY: build

docker:
	docker build -t vidcall:latest -f ./deployment/Dockerfile .
