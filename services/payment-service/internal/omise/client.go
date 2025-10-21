package omisecli

import "github.com/omise/omise-go"

func NewOmiseClient(pub, sec, apiVersion string) (*omise.Client, error) {
	c, err := omise.NewClient(pub, sec)
	if err != nil {
		return nil, err
	}
	if apiVersion != "" {
		// c.GoVersion = apiVersion // pin version ให้พฤติกรรม API คงที่
	}
	c.SetDebug(false)
	return c, nil
}
