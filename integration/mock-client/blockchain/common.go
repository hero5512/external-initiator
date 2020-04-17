package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type JsonrpcMessage struct {
	Version string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   *interface{}    `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

func HandleRequest(conn, platform string, msg JsonrpcMessage) ([]JsonrpcMessage, error) {
	switch platform {
	case "eth":
		return HandleEthRequest(conn, msg)
	default:
		return nil, errors.New(fmt.Sprint("unexpected platform: ", platform))
	}
}

func SetHttpRoutes(routerGroup *gin.RouterGroup) {
	routerGroup.GET("/xtz/monitor/heads/:chain_id", HandleXtzMonitorRequest)
	routerGroup.GET("/xtz/chains/main/blocks/:block_id/operations", HandleXtzOperationsRequest)
}

func HandleXtzMonitorRequest(c *gin.Context) {
	resp, err := GetXtzMonitorResponse(c.Param("chain_id"))

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func HandleXtzOperationsRequest(c *gin.Context) {
	resp, err := GetXtzOperationsResponse(c.Param("block_id"))

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	c.JSON(http.StatusOK, resp)
}
