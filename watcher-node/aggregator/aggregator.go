package aggregator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"thirdlight.com/watcher-node/lib"
)

type Aggregator struct {
	baseUrl *url.URL
	client  *http.Client
}

func New(client *http.Client) (*Aggregator, error) {
	aggregatorAddr := os.Getenv("FILE_AGGREGATOR_ADDRESS")
	if aggregatorAddr == "" {
		return nil, errors.New("no aggregator address - set FILE_AGGREGATOR_ADDRESS as an environment variable")
	}
	urlObj, err := url.Parse(aggregatorAddr)
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
		return errors.New("No FILE_AGGREGATOR_ADDRESS set")
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
	req.Close = true

	if err != nil {
		return err
	}

	resp, err := ag.client.Do(req)
	if err != nil {
		return fmt.Errorf("Aggregator client error: %w", err)
	}
	resp.Body.Close()

	return nil
}
