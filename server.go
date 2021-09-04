package main

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/cors"
	"kontrakt-server/dataloader"
	"kontrakt-server/graph"
	"kontrakt-server/graph/auth"
	"kontrakt-server/graph/generated"
	"kontrakt-server/graph/model"
	"kontrakt-server/prisma/db"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	"github.com/gorilla/mux"
)

var muxRouter *mux.Router
var prismaClient *db.PrismaClient

func init() {
	muxRouter = mux.NewRouter()

	if prismaClient == nil {
		client := db.NewClient()
		if err := client.Prisma.Connect(); err != nil {
			panic(err)
		}
		prismaClient = client
	}

	config := generated.Config{Resolvers: &graph.Resolver{
		Prisma: prismaClient,
	}}

	config.Directives.HasRole = func(ctx context.Context, obj interface{}, next graphql.Resolver, role model.Role) (interface{}, error) {
		forContext := auth.ForContext(ctx)
		if forContext == nil || string(forContext.Role) != role.String() {
			// block calling the next resolver
			return nil, fmt.Errorf("Access denied")
		}

		// or let it pass through
		return next(ctx)
	}

	config.Directives.IsLoggedIn = func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
		forContext := auth.ForContext(ctx)
		if forContext == nil {
			// block calling the next resolver
			return nil, fmt.Errorf("Access denied")
		}

		// or let it pass through
		return next(ctx)
	}

	schema := generated.NewExecutableSchema(config)
	server := handler.NewDefaultServer(schema)
	muxRouter.Handle("/query", dataloader.Middleware(prismaClient, server))
	muxRouter.Handle("/", playground.Handler("GraphQL playground", "/query"))
	muxRouter.Use(auth.Middleware(prismaClient))
	muxRouter.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://kontrakt.maelkerichard.com", "https://kontrakt.ecolepontpean.fr"},
		AllowCredentials: true,
		Debug:            true,
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
	}).Handler)

}

func lambdaHandler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	muxAdapter := gorillamux.New(muxRouter)

	rsp, err := muxAdapter.Proxy(req)
	if err != nil {
		log.Println(err)
	}
	return rsp, err
}

func main() {
	isRunningAtLambda := strings.Contains(os.Getenv("AWS_EXECUTION_ENV"), "AWS_Lambda_")

	if isRunningAtLambda {
		lambda.Start(lambdaHandler)
	} else {
		defaultPort := "7010"
		port := os.Getenv("PORT")

		if port == "" {
			port = defaultPort
		}

		log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
		log.Fatal(http.ListenAndServe(":"+port, muxRouter))
	}
}
