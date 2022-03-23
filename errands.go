package errands

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	schemas "github.com/polygon-io/errands-server/schemas"
)

type ErrandsAPI struct {
	EndpointURL string
	Processors  []*Processor
}

func New(url string) *ErrandsAPI {
	obj := &ErrandsAPI{}
	obj.EndpointURL = url
	return obj
}

//easyjson:json
type ErrandsResponse struct {
	Results []schemas.Errand `json:"results"`
	Status  string           `json:"status"`
}

//easyjson:json
type ErrandResponse struct {
	Results schemas.Errand `json:"results"`
	Status  string         `json:"status"`
}

func (e *ErrandsAPI) GetErrands() (*ErrandsResponse, error) {
	res, err := e.get("/v1/errands/")
	if err != nil {
		return &ErrandsResponse{}, err
	}
	return parseErrandsResponse(res)
}

// ListErrands queries the errands API for a list of errands that match the given query.
// Possible options for key are: status and type.
func (e *ErrandsAPI) ListErrands(key, val string) (*ErrandsResponse, error) {
	path := fmt.Sprintf("/v1/errands/list/%s/%s", key, val)
	res, err := e.get(path)
	if err != nil {
		return &ErrandsResponse{}, err
	}
	return parseErrandsResponse(res)
}

func (e *ErrandsAPI) CreateErrand(errand *schemas.Errand) (*ErrandResponse, error) {
	errandBytes, err := errand.MarshalJSON()
	if err != nil {
		return &ErrandResponse{}, err
	}
	resp, err := http.Post(e.EndpointURL+"/v1/errands/", "application/json", bytes.NewBuffer(errandBytes))
	if err != nil {
		return &ErrandResponse{}, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &ErrandResponse{}, err
	}
	return parseErrandResponse(body)
}

func (e *ErrandsAPI) RequestErrandToProcess(topic string) (*ErrandResponse, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", e.EndpointURL+"/v1/errands/process/"+topic, nil)
	if err != nil {
		return &ErrandResponse{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return &ErrandResponse{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &ErrandResponse{}, err
	}
	return parseErrandResponse(body)
}

//easyjson:json
type FailErrandReq struct {
	Reason string `json:"reason"`
}

func (e *ErrandsAPI) FailErrand(errandId, reason string) (*ErrandResponse, error) {
	failReq := &FailErrandReq{reason}
	failReqBytes, err := failReq.MarshalJSON()
	if err != nil {
		return &ErrandResponse{}, err
	}
	client := &http.Client{}
	req, err := http.NewRequest("PUT", e.EndpointURL+"/v1/errand/"+errandId+"/failed", bytes.NewBuffer(failReqBytes))
	if err != nil {
		return &ErrandResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return &ErrandResponse{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &ErrandResponse{}, err
	}
	return parseErrandResponse(body)
}

//easyjson:json
type CompleteErrandReq struct {
	Results map[string]interface{} `json:"results"`
}

func (e *ErrandsAPI) CompleteErrand(errandId string, results map[string]interface{}) (*ErrandResponse, error) {
	compReq := &CompleteErrandReq{results}
	compReqBytes, err := compReq.MarshalJSON()
	if err != nil {
		return &ErrandResponse{}, err
	}
	client := &http.Client{}
	req, err := http.NewRequest("PUT", e.EndpointURL+"/v1/errand/"+errandId+"/completed", bytes.NewBuffer(compReqBytes))
	if err != nil {
		return &ErrandResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return &ErrandResponse{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &ErrandResponse{}, err
	}
	return parseErrandResponse(body)
}

func (e *ErrandsAPI) DeleteErrand(errandId string) (*ErrandResponse, error) {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", e.EndpointURL+"/v1/errand/"+errandId, nil)
	if err != nil {
		return &ErrandResponse{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return &ErrandResponse{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &ErrandResponse{}, err
	}
	return parseErrandResponse(body)
}

func parseErrandResponse(res []byte) (*ErrandResponse, error) {
	errandRes := &ErrandResponse{}
	if err := errandRes.UnmarshalJSON(res); err != nil {
		return &ErrandResponse{}, err
	}
	return errandRes, nil
}

func parseErrandsResponse(res []byte) (*ErrandsResponse, error) {
	errandRes := &ErrandsResponse{}
	if err := errandRes.UnmarshalJSON(res); err != nil {
		return &ErrandsResponse{}, err
	}
	return errandRes, nil
}

func (e *ErrandsAPI) get(url string) ([]byte, error) {
	resp, err := http.Get(e.EndpointURL + url)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
