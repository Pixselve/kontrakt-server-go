build:
	env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bin/server

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose


generateResolvers:
	go run github.com/99designs/gqlgen generate

generatePrisma:
	go run github.com/prisma/prisma-client-go generate