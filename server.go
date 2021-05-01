package main

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dgrijalva/jwt-go"
	"kontrakt-server/graph"
	"kontrakt-server/graph/generated"
	"kontrakt-server/graph/model"
	"kontrakt-server/prisma/db"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	"github.com/gorilla/mux"
)

var muxRouter *mux.Router
var prismaClient *db.PrismaClient

var userCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

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
		forContext := ForContext(ctx)
		if forContext == nil || string(forContext.Role) != role.String() {
			// block calling the next resolver
			return nil, fmt.Errorf("Access denied")
		}

		// or let it pass through
		return next(ctx)
	}

	schema := generated.NewExecutableSchema(config)
	server := handler.NewDefaultServer(schema)
	muxRouter.Handle("/query", server)
	muxRouter.Handle("/", playground.Handler("GraphQL playground", "/query"))
	muxRouter.Use(Middleware(prismaClient))

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

func Middleware(prisma *db.PrismaClient) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := r.Header.Get("Authorization")



			// Allow unauthenticated users in
			if len(tokenString) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
			claims, err := verifyToken(tokenString)
			if err != nil {
				http.Error(w, "Error verifying JWT token: " + err.Error(), http.StatusUnauthorized)
				return
			}

			username, ok := claims.(jwt.MapClaims)["username"].(string)
			if !ok {
				http.Error(w, "Invalid user", http.StatusForbidden)
				return
			}
			user, err := prisma.User.FindUnique(db.User.Username.Equals(username)).Exec(r.Context())
			if err != nil {
				http.Error(w, "Invalid user", http.StatusForbidden)
				return
			}

			//// put it in context
			ctx := context.WithValue(r.Context(), userCtxKey, user)

			// and call the next with our new context
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func ForContext(ctx context.Context) *db.UserModel {
	raw, _ := ctx.Value(userCtxKey).(*db.UserModel)
	return raw
}