package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danielperaltamadriz/tinyurl/api"
	"github.com/danielperaltamadriz/tinyurl/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type bodyModel struct {
	URL string `json:"url"`
}

func TestAcceptanceTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AcceptanceTest Suite")
}

var _ = Describe("Request tiny URL", Ordered, func() {
	// var server *ghttp.Server
	var apiServer *api.API
	var postgresContainer *postgres.PostgresContainer
	ctx := context.Background()
	dbUser := "username"
	dbName := "db_name"
	dbPassword := "password"
	serverURL := "http://localhost:8080"

	BeforeAll(func() {
		var err error
		postgresContainer, err = postgres.RunContainer(ctx,
			postgres.WithDatabase(dbName),
			postgres.WithUsername(dbUser),
			postgres.WithPassword(dbPassword),
			testcontainers.WithWaitStrategy(
				wait.ForExposedPort(),
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(1).
					WithStartupTimeout(10*time.Second)),
		)
		if err != nil {
			Fail("couldn't start postgres: " + err.Error())
		}
	})

	BeforeEach(func() {
		connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			Fail("failed to get postgres connection string: " + err.Error())
		}
		apiServer, err = api.NewAPI(config.Config{
			DB: config.DB{
				ConnectionString: connStr,
			},
		})
		if err != nil {
			Fail("failed to create api server: " + err.Error())
		}
		go func() {
			apiServer.Start() // nolint:errcheck
		}()
		// server.AppendHandlers(apiServer.Handler())
	})

	AfterEach(func() {
		if apiServer != nil {
			apiServer.Shutdown() // nolint:errcheck
		}
	})

	AfterAll(func() {
		if postgresContainer != nil {
			postgresContainer.Terminate(context.Background()) // nolint:errcheck
		}
	})

	Context("Given an user", func() {
		When("a valid tiny url is requested", func() {
			var postStatusCode int
			var getStatusCode int
			var postResponse api.Response
			BeforeEach(func() {
				ts := httptestSetup(setupHTTPTest{
					statusCode: http.StatusOK,
				})
				defer ts.Close()
				reqModel := bodyModel{
					URL: ts.URL,
				}
				body, _ := json.Marshal(reqModel)
				postResp, _ := http.Post(serverURL+"/tiny", "application/json", bytes.NewReader(body))
				if postResp != nil && postResp.Body != nil {
					postStatusCode = postResp.StatusCode
					defer postResp.Body.Close()
					b, _ := io.ReadAll(postResp.Body)
					json.Unmarshal(b, &postResponse) //nolint:errcheck
					getResp, err := http.Get(serverURL + postResponse.TinyURL)
					Expect(err).To(BeNil())
					getStatusCode = getResp.StatusCode
				}

			})

			It("should return a valid status code", func() {
				Expect(postStatusCode).To(Equal(http.StatusCreated))
			})
			It("should return a valid tiny url", func() {
				Expect(postResponse.TinyURL).To(Not(BeEmpty()))
			})
			It("should return a valid redirection code", func() {
				Expect(getStatusCode).To(Equal(http.StatusOK))

			})
		})
		When("an invalid tiny url is requested", func() {
			var statusCode int
			BeforeEach(func() {
				ts := httptestSetup(setupHTTPTest{
					statusCode: http.StatusNotFound,
				})
				defer ts.Close()
				model := struct {
					InvalidURL int `json:"url"`
				}{}
				body, _ := json.Marshal(model)
				resp, _ := http.Post(serverURL+"/tiny", "application/json", bytes.NewReader(body))
				statusCode = resp.StatusCode
			})

			It("should return a 400 status code", func() {
				Expect(statusCode).To(Equal(http.StatusBadRequest))
			})
		})
	})
})

type setupHTTPTest struct {
	statusCode int
}

func httptestSetup(setup setupHTTPTest) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(setup.statusCode)
	}))
	return ts
}
