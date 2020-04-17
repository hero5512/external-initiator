package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

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

type MockResponse struct {
	Path   string      `json:"path"`
	Method string      `json:"method"`
	Code   int         `json:"code"`
	Body   interface{} `json:"body"`
}

func SetHttpRoutesFromJSON(routerGroup *gin.RouterGroup) error {
	wd, _ := os.Getwd()
	responsesPath := path.Join(wd, "mock-responses")
	files, err := ioutil.ReadDir(responsesPath)

	if err != nil {
		return err
	}

	for _, f := range files {
		path := path.Join(responsesPath, f.Name())
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		data, err := ioutil.ReadAll(file)

		var resp MockResponse
		err = json.Unmarshal(data, &resp)
		if err != nil {
			return err
		}
		routerGroup.Handle(strings.ToUpper(resp.Method), resp.Path, func(c *gin.Context) {
			c.JSON(resp.Code, resp.Body)
		})
	}

	return nil
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
