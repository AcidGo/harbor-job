package harbor

import (
    "fmt"
)

func (c *Client) Login(user, pwd string) (error) {
    url := c.apiPath.Ping()
    resp, err := c.do(MethodGet, url, nil, nil)
    if err != nil {
        return err
    }
    if resp.StatusCode/100 > 2 {
        return fmt.Errorf("get http status code %d when getting cookies from api ping", resp.StatusCode)
    }

    c.cookies = resp.Cookies()

    url = c.apiPath.Login(user, pwd)
    _, err = c.do(MethodPost, url, nil, nil)
    if err != nil {
        return err
    }

    return nil
}