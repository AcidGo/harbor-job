package harbor

import (
    "encoding/json"
)

type Repository struct {
    Id          int     `json:"id"`
    Name        string  `json:"name"`
    ProjectId   int     `json:"project_id"`
    Description string  `json:"description"`
}

type Repositories []Repository

func (c *Client) Repositories(projectId int) (Repositories, error) {
    var repositories Repositories

    url := c.apiPath.Repositories(projectId)
    resp, err := c.do(MethodGet, url, nil, nil)
    if err != nil {
        return repositories, err
    }
    defer resp.Body.Close()

    err = json.NewDecoder(resp.Body).Decode(&repositories)
    if err != nil {
        return repositories, err
    }

    return repositories, nil
}
