package harbor

import (
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    neturl "net/url"
    "strings"
)

const (
    MethodPost      = http.MethodPost
    MethodGet       = http.MethodGet
    MethodDelete    = http.MethodDelete

    ScopePrject     = "p"
)

type Client struct {
    url         string
    username    string
    passwd      string
    hasLogin    bool
    client      *http.Client
    cookies     []*http.Cookie
    apiPath     *ApiPath
}

func NewClient(url, version string) (*Client, error) {
    u, err := neturl.Parse(url)
    if err != nil {
        return nil, err
    }

    if u.Scheme != "https" && u.Scheme != "http" {
        return nil, fmt.Errorf("not support the url schema %s", u.Scheme)
    }

    apiPath, err := NewApiPath(u.String(), version)
    if err != nil {
        return nil, err
    }

    return &Client{
        url:        u.String(),
        hasLogin:   false,
        client:     &http.Client{},
        apiPath:    apiPath,
    }, nil
}

func (c *Client) do(method, url string, param map[string]string, data map[string]string) (*http.Response, error) {
    var body io.Reader
    var err error
    var req *http.Request

    v := neturl.Values{}
    if data != nil {
        for key, val := range data {
            v.Set(key, val)
        }
    }
    body = strings.NewReader(v.Encode())

    req, err = http.NewRequest(method, url, body)
    if err != nil {
        return nil, err
    }

    q := req.URL.Query()
    if param != nil {
        for key, val := range param {
            q.Add(key, val)
        }
    }
    req.URL.RawQuery = q.Encode()
    log.Printf("url of request: %s", req.URL.String())

    req.Header.Set("Content-Type", "application/json")
    for i := range c.cookies {
        req.AddCookie(c.cookies[i])
    }

    // log.Printf("%#v", req)
    // log.Printf("%#v", req.Cookies())
    resp, err := c.client.Do(req)
    if err != nil {
        return nil, err
    }

    if resp.StatusCode/100 > 2 {
        log.Printf("statusCode of response is %d", resp.StatusCode)

        b, err := ioutil.ReadAll(resp.Body)
        defer resp.Body.Close()
        if err != nil {
            return nil, err
        }

        return nil, fmt.Errorf(string(b))
    }

    return resp, nil
}