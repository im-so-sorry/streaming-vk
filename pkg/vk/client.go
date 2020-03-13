package vk

type Client struct {
	accessToken string
	Version string

	baseUrl string
}

func (c *Client) makeRequest() {

}

func (c *Client) Initialize(accessToken string) error {
	c.accessToken = accessToken

	c.Version = "5.103"
	c.baseUrl = "https://api.vk.com/method"

	return nil
}

func (c *Client) Authorize() (string, error) {
	return "", nil
}