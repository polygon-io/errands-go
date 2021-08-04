package errands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/polygon-io/errands-server/schemas"
)

//easyjson:json
type CreatePipelineResponse struct {
	Results schemas.Pipeline `json:"results"`
	Status  string           `json:"status"`
}

func (e *ErrandsAPI) CreatePipeline(ctx context.Context, pipeline *schemas.Pipeline) (*CreatePipelineResponse, error) {
	pipelineBytes, err := pipeline.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal pipeline json: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.EndpointURL+"/v1/pipeline/", bytes.NewReader(pipelineBytes))
	if err != nil {
		return nil, fmt.Errorf("new http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	response := &CreatePipelineResponse{}
	if err := requestAndUnmarshalResponse(req, response); err != nil {
		return nil, err
	}

	return response, nil
}

//easyjson:json
type DeletePipelineResponse struct {
	Status string `json:"status"`
}

func (e *ErrandsAPI) DeletePipeline(ctx context.Context, pipelineID string) (*DeletePipelineResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, e.EndpointURL+"/v1/pipeline/"+pipelineID, nil)
	if err != nil {
		return nil, fmt.Errorf("new http request: %w", err)
	}

	response := &DeletePipelineResponse{}
	if err := requestAndUnmarshalResponse(req, response); err != nil {
		return nil, err
	}

	return response, nil
}

//easyjson:json
type GetPipelineResponse struct {
	Results schemas.Pipeline `json:"results"`
	Status  string           `json:"status"`
}

func (e *ErrandsAPI) GetPipeline(ctx context.Context, pipelineID string) (*GetPipelineResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.EndpointURL+"/v1/pipeline/"+pipelineID, nil)
	if err != nil {
		return nil, fmt.Errorf("new http request: %w", err)
	}

	response := &GetPipelineResponse{}
	if err := requestAndUnmarshalResponse(req, response); err != nil {
		return nil, err
	}

	return response, nil
}

//easyjson:json
type ListPipelineResponse struct {
	Results []*schemas.Pipeline `json:"results"`
	Status  string              `json:"status"`
}

// ListPipelines lists all pipelines filtered by status. If statusFilter is empty string, it will list all pipelines of all statuses.
// Note that the pipelines in the response will _not_ have their list of `errands` or `dependencies` populated.
// To see that information you'll have to do a Get for the individual pipeline by ID.
func (e *ErrandsAPI) ListPipelines(ctx context.Context, statusFilter string) (*ListPipelineResponse, error) {
	var query string

	if statusFilter != "" {
		v := make(url.Values, 1)

		v.Set("status", statusFilter)
		query = "?" + v.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.EndpointURL+"/v1/pipelines/"+query, nil)
	if err != nil {
		return nil, fmt.Errorf("new http request: %w", err)
	}

	response := &ListPipelineResponse{}
	if err := requestAndUnmarshalResponse(req, response); err != nil {
		return nil, err
	}

	return response, nil
}

func requestAndUnmarshalResponse(req *http.Request, unmarshaller json.Unmarshaler) error {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}

	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if err := unmarshaller.UnmarshalJSON(respBytes); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	return nil
}
