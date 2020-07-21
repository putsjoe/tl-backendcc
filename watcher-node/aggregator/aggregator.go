package aggregator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"thirdlight.com/watcher-node/lib"
)

type Aggregator struct {
	baseUrl *url.URL
	client  *http.Client
}

func New(client *http.Client, addr string) (*Aggregator, error) {
	if addr == "" {
		return nil, errors.New("no aggregation server address provided")
	}
	urlObj, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	log.Println("[INFO] Communicating with aggregator server at", urlObj.String())

	return &Aggregator{
		baseUrl: urlObj,
		client:  client,
	}, nil
}
func (ag *Aggregator) Hello(instance string, listenPort uint) error {
	body := lib.HelloOperation{
		BaseMessage: lib.BaseMessage{instance},
		Port:        listenPort,
	}
	return ag.send(http.MethodPost, "hello", body)
}

func (ag *Aggregator) Bye(instance string) error {
	body := lib.ByeOperation{lib.BaseMessage{instance}}
	return ag.send(http.MethodPost, "bye", body)
}

func (ag *Aggregator) NotifyUpdate(op string, filename string, seqNo int, instance string) error {
	body := []lib.PatchOperation{
		lib.PatchOperation{
			Op: op,
			Value: lib.FileMetadata{
				Filename: filename,
			},
			Sequence:    seqNo,
			BaseMessage: lib.BaseMessage{instance},
		},
	}
	return ag.send(http.MethodPatch, "files", body)
}

func (ag *Aggregator) send(method, path string, body interface{}) error {
	if ag.baseUrl == nil {
		return errors.New("no aggregation server address configured")
	}
	u, err := url.Parse(path)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		method,
		ag.baseUrl.ResolveReference(u).String(),
		bytes.NewReader(payload),
	)
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return err
	}

	resp, err := ag.client.Do(req)
	if err != nil {
		return fmt.Errorf("Aggregator client error: %s", err.Error())
	}

	defer resp.Body.Close()
	_, bodyErr := ioutil.ReadAll(resp.Body)
	if bodyErr != nil {
		return fmt.Errorf("Aggregator client error: %s", bodyErr.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Aggregator client non-200 response: %s", resp.Status)
	}

	return nil
}
