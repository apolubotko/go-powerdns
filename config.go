package powerdns

import (
	"fmt"

	"github.com/joeig/go-powerdns/v2/lib"
)

// ConfigService handles communication with the zones related methods of the Client API
type ConfigService service

// List retrieves a list of ConfigSettings
func (c *ConfigService) List() ([]lib.ConfigSetting, error) {
	req, err := c.client.newRequest("GET", fmt.Sprintf("servers/%s/config", c.client.VHost), nil, nil)
	if err != nil {
		return nil, err
	}

	config := make([]lib.ConfigSetting, 0)
	_, err = c.client.do(req, &config)

	return config, err
}
