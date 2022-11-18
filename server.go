package main

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
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

	// create default user
	ctx := context.Background()
	userCount, err := prismaClient.User.FindMany().Exec(ctx)
	if err != nil {
		panic(err)
	}
	if len(userCount) == 0 {
		// no users

		username := os.Getenv("USERNAME")
		password := os.Getenv("PASSWORD")
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			panic(err)
		}
		createdUser, err := prismaClient.User.CreateOne(db.User.Username.Set(username), db.User.Password.Set(string(hashedPassword)), db.User.Role.Set(db.RoleTEACHER)).Exec(ctx)
		if err != nil {
			panic(err)
		}
		_, err = prismaClient.Teacher.CreateOne(db.Teacher.Owner.Link(db.User.Username.Equals(createdUser.Username)), db.Teacher.FirstName.Set("admin"), db.Teacher.LastName.Set("admin")).Exec(ctx)
		if err != nil {
			panic(err)
		}
	}

	schema := generated.NewExecutableSchema(config)
	server := handler.NewDefaultServer(schema)
	muxRouter.Handle("/query", dataloader.Middleware(prismaClient, server))
	muxRouter.Handle("/", playground.Handler("GraphQL playground", "/query"))
	muxRouter.Use(auth.Middleware(prismaClient))
	muxRouter.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Content-Type", "X-Amz-Date", "Authorization", "X-Api-Key", "X-Amz-Security-Token"},
		AllowedMethods:   []string{"DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"},
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
