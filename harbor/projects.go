package harbor

import (
    "encoding/json"
)

type Project struct {
    ProjectId       int     `json:"project_id"`
    OwnerId         int     `json:"owner_id"`
    Name            string  `json:"name"`
    Deleted         bool    `json:"deleted"`
    RepoCount       int     `json:"repo_count"`
}

type Projects []Project

func (c *Client) Projects() ([]Project, error) {
    var projects Projects

    url := c.apiPath.Projects()
    resp, err := c.do(MethodGet, url, nil, nil)
    if err != nil {
        return projects, err
    }
    defer resp.Body.Close()

    err = json.NewDecoder(resp.Body).Decode(&projects)
    if err != nil {
        return projects, err
    }

    return projects, nil
}