package client

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/blockchain"
	"github.com/smartcontractkit/external-initiator/store"
	"net/http"
)

const (
	externalInitiatorAccessKeyHeader = "X-Chainlink-EA-AccessKey"
	externalInitiatorSecretHeader    = "X-Chainlink-EA-Secret"
)

type subscriptionStorer interface {
	SaveSubscription(sub *store.Subscription) error
	DeleteJob(jobid string) error
	GetEndpoint(name string) (*store.Endpoint, error)
	SaveEndpoint(endpoint *store.Endpoint) error
}

// RunWebserver starts a new web server using the access key
// and secret as provided on protected routes.
func RunWebserver(
	accessKey, secret string,
	store subscriptionStorer,
) {
	srv := NewHTTPService(accessKey, secret, store)
	err := srv.Router.Run()
	if err != nil {
		logger.Error(err)
	}
}

// HttpService encapsulates router, EI service
// and access credentials.
type HttpService struct {
	Router    *gin.Engine
	AccessKey string
	Secret    string
	Store     subscriptionStorer
}

// NewHTTPService creates a new HttpService instance
// with the default router.
func NewHTTPService(
	accessKey, secret string,
	store subscriptionStorer,
) *HttpService {
	srv := HttpService{
		Router:    gin.Default(),
		AccessKey: accessKey,
		Secret:    secret,
		Store:     store,
	}
	srv.createRouter()
	return &srv
}

// ServeHTTP calls ServeHTTP on the underlying router,
// which conforms to the http.Handler interface.
func (srv *HttpService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	srv.Router.ServeHTTP(w, r)
}

func (srv *HttpService) createRouter() {
	r := gin.Default()
	r.GET("/health", srv.ShowHealth)

	auth := r.Group("/")
	auth.Use(authenticate(srv.AccessKey, srv.Secret))
	{
		auth.POST("/jobs", srv.CreateSubscription)
		auth.DELETE("/jobs/:jobid", srv.DeleteSubscription)
		auth.POST("/config", srv.CreateEndpoint)
	}

	srv.Router = r
}

func authenticate(accessKey, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqAccessKey := c.GetHeader(externalInitiatorAccessKeyHeader)
		reqSecret := c.GetHeader(externalInitiatorSecretHeader)
		if reqAccessKey == accessKey && reqSecret == secret {
			c.Next()
		} else {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}

// CreateSubscriptionReq holds the payload expected for job POSTs
// from the Chainlink node.
type CreateSubscriptionReq struct {
	JobID  string            `json:"jobId"`
	Type   string            `json:"type"`
	Params blockchain.Params `json:"params"`
}

func validateRequest(t *CreateSubscriptionReq, endpointType string) error {
	validations := append([]int{
		len(t.JobID),
	}, blockchain.GetValidations(endpointType, t.Params)...)

	for _, v := range validations {
		if v < 1 {
			return errors.New("missing required field(s)")
		}
	}

	return nil
}

type resp struct {
	ID string `json:"id"`
}

// CreateSubscription expects a CreateSubscriptionReq payload,
// validates the request and subscribes to the job.
func (srv *HttpService) CreateSubscription(c *gin.Context) {
	var req CreateSubscriptionReq

	if err := c.BindJSON(&req); err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	endpoint, err := srv.Store.GetEndpoint(req.Params.Endpoint)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	if endpoint == nil {
		logger.Error("unknown endpoint provided")
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	if err := validateRequest(&req, endpoint.Type); err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	sub := &store.Subscription{
		ReferenceId:  uuid.New().String(),
		Job:          req.JobID,
		EndpointName: req.Params.Endpoint,
		Endpoint:     *endpoint,
	}

	blockchain.CreateSubscription(sub, req.Params)

	if err := srv.Store.SaveSubscription(sub); err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusCreated, resp{ID: sub.ReferenceId})
}

// DeleteSubscription deletes any job with the jobid
// provided as parameter in the request.
func (srv *HttpService) DeleteSubscription(c *gin.Context) {
	jobid := c.Param("jobid")
	if err := srv.Store.DeleteJob(jobid); err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, resp{ID: jobid})
}

// ShowHealth returns the following when online:
//  {"chainlink": true}
func (srv *HttpService) ShowHealth(c *gin.Context) {
	c.JSON(200, gin.H{"chainlink": true})
}

// CreateEndpoint saves the endpoint configuration provided
// as payload.
func (srv *HttpService) CreateEndpoint(c *gin.Context) {
	var config store.Endpoint
	err := c.BindJSON(&config)
	if err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	if err := srv.Store.SaveEndpoint(&config); err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusCreated, resp{ID: config.Name})
}
