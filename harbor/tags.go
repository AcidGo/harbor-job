package harbor

import (
    "encoding/json"
    "fmt"
    "time"
)

type Tag struct {
    Digest          string      `json:"digest"`
    Name            string      `json:"name"`
    Size            int32       `json:"size"`
    Architecture    string      `json:"architecture"`
    OS              string      `json:"os"`
    DockerVersion   string      `json:"docker_version"`
    Auhtor          string      `json:"docker_version"`
    CreatedTime     time.Time   `json:"created"`
}

type Tags []Tag

func (ts Tags) Len() (int) {
    return len(ts)
}

func (ts Tags) Less(i, j int) (bool) {
    return ts[i].CreatedTime.After(ts[j].CreatedTime)
}

func (ts Tags) Swap(i, j int) {
    ts[i], ts[j] = ts[j], ts[i]
}

func (c *Client) Tags(repoPath string) (Tags, error) {
    var tags Tags

    url := c.apiPath.Tags(repoPath)
    resp, err := c.do(MethodGet, url, nil, nil)
    if err != nil {
        return tags, err
    }
    defer resp.Body.Close()

    err = json.NewDecoder(resp.Body).Decode(&tags)
    if err != nil {
        return tags, err
    }

    return tags, nil
}

func (c *Client) DeleteTag(repoPath, tagName string) (error) {
    url := c.apiPath.Tag(repoPath, tagName)
    resp, err := c.do(MethodDelete, url, nil, nil)
    if err != nil {
        return err
    }

    if resp.StatusCode/100 > 2 {
        return fmt.Errorf("get bad http status code %d when delete tag %s in %s", resp.StatusCode, tagName, repoPath)
    }

    return nil
}