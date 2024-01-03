package storage

import (
	"github.com/pkg/errors"
	"sync"
)

type OpenIdDict struct {
	dict map[string]string
	mu   sync.Mutex
}

func NewOpenIdDict() *OpenIdDict {
	return &OpenIdDict{
		dict: make(map[string]string),
	}
}

var OpenIdService *OpenIdDict

func init() {
	OpenIdService = NewOpenIdDict()
}

func (d *OpenIdDict) GetOpenIdByAccessToken(accessToken string) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	openId, ok := d.dict[accessToken]
	if ok {
		return openId, nil
	}
	return "", errors.New("AccessToken not found")
}

func (d *OpenIdDict) Save(accessToken, openId string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.dict[accessToken] = openId
}
