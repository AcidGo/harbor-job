package harbor

import (
    "encoding/json"
)

type Label struct {
    Id          int     `json:"id"`
    Name        string  `json:"name"`
    ProjectId   int     `jsno:"project_id"`
}

type Labels []Label

func (c *Client) Labels(scope string, projectId int) (Labels, error) {
    var labels Labels

    url := c.apiPath.Labels(scope, projectId)
    resp, err := c.do(MethodGet, url, nil, nil)
    if err != nil {
        return labels, err
    }
    defer resp.Body.Close()

    err = json.NewDecoder(resp.Body).Decode(&labels)
    if err != nil {
        return labels, err
    }

    return labels, nil
}