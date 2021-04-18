package harbor

import (
    "fmt"
)

type ApiPath struct {
    url         string
    version     string
}

func NewApiPath(url, version string) (*ApiPath, error) {
    if version == "" {
        version = "1.0"
    }

    return &ApiPath{
        url:        url,
        version:    version,
    }, nil
}

func (ap *ApiPath) Login(user, pwd string) (string) {
    if ap.version >= "1.7" {
        return fmt.Sprintf("%s/c/login?principal=%s&password=%s", ap.url, user, pwd)
    }
    return fmt.Sprintf("%s/login?principal=%s&password=%s", ap.url, user, pwd)
}

func (ap *ApiPath) Ping() (string) {
    return fmt.Sprintf("%s/api/ping", ap.url)
}

func (ap *ApiPath) Projects() (string) {
    return fmt.Sprintf("%s/api/projects", ap.url)
}

func (ap *ApiPath) Repositories(projectId int) (string) {
    return fmt.Sprintf("%s/api/repositories?project_id=%d", ap.url, projectId)
}

func (ap *ApiPath) Labels(scope string, projectId int) (string) {
    return fmt.Sprintf("%s/api/labels?scope=%s&project_id=%d", ap.url, scope, projectId)
}

func (ap *ApiPath) Tags(repoPath string) (string) {
    return fmt.Sprintf("%s/api/repositories/%s/tags", ap.url, repoPath)
}

func (ap *ApiPath) Tag(repoPath, tagName string) (string) {
    return fmt.Sprintf("%s/api/repositories/%s/tags/%s", ap.url, repoPath, tagName)
}