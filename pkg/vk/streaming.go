package vk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"
)

type StreamingAuth struct {
	Endpoint string
	Key      string
}

type (
	vkApiResponse struct {
		Response StreamingAuth
	}
)

func (c *Client) GetServerUrl() (StreamingAuth, error) {

	method := "streaming.getServerUrl"

	apiUrl := fmt.Sprintf("%s/%s/?access_token=%s&v=%s", c.baseUrl, method, c.accessToken, c.Version)

	resp, err := http.Get(apiUrl)

	if err != nil {
		log.Fatal("streaming api authorization failed:", err)
	}
	defer resp.Body.Close()
	bodyBuf, err := ioutil.ReadAll(resp.Body)

	var v vkApiResponse
	if err := json.Unmarshal(bodyBuf, &v); err != nil {
		log.Fatal("unmarshal response json failed:", err)
	}

	return v.Response, nil
}

type Rule struct {
	Tag   string
	Value string
}

type ErrorMessage struct {
	Message   string
	ErrorCode int
}

type RulesResponse struct {
	Code  int
	Rules []Rule
	Error ErrorMessage
}

func (c *Client) GetRules(credentials StreamingAuth) ([]Rule, error) {
	url := fmt.Sprintf("https://%s/rules/?key=%s", credentials.Endpoint, credentials.Key)

	req, err := http.NewRequest("GET", url, nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("http request error:", err)
	}
	defer resp.Body.Close()

	bodyBuf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("response body read error:", err)
	}

	var v RulesResponse
	if err := json.Unmarshal(bodyBuf, &v); err != nil {
		log.Fatal("unmarshal response json failed:", err)
	}

	return v.Rules, nil
}

type RuleResponse struct {
	Code  int
	Error ErrorMessage
}

func (c *Client) RemoveRule(credentials StreamingAuth, tag string) (RuleResponse, error) {
	var v RuleResponse

	url := fmt.Sprintf("https://%s/rules/?key=%s", credentials.Endpoint, credentials.Key)

	body, err := json.Marshal(struct{ Tag string }{tag})

	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("http request error:", err)
	}
	defer resp.Body.Close()

	bodyBuf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("response body read error:", err)
	}

	if err := json.Unmarshal(bodyBuf, &v); err != nil {
		log.Fatal("unmarshal response json failed:", err)
		return v, err
	}

	return v, nil
}

func (c *Client) AddRule(credentials StreamingAuth, rule Rule) (RuleResponse, error) {
	var v RuleResponse

	url := fmt.Sprintf("https://%s/rules/?key=%s", credentials.Endpoint, credentials.Key)

	body, err := json.Marshal(struct{ Rule Rule }{rule})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("http request error:", err)
	}
	defer resp.Body.Close()

	bodyBuf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("response body read error:", err)
	}

	if err := json.Unmarshal(bodyBuf, &v); err != nil {
		log.Fatal("unmarshal response json failed:", err)
		return v, err
	}

	return v, nil
}

type ServiceMessage struct {
	Message     string
	ServiceCode int
}

//post; comment; share; topic_post
type EventType string
type EventId struct {
	PostOwnerId  int
	PostId       int
	CommentId    int
	SharedPostId int
	TopicOwnerId int
	TopicId      int
	TopicPostId  int
}

//new; update; delete; restore
type ActionType string

type AttachmentType struct {
	Type  string
	Photo AttachmentPhotoType
	Video AttachmentVideoType
}

type StreamEvent struct {
	EventType EventType
	EventId   EventId
	EventUrl  string
	Text      string

	Action       ActionType
	ActionTime   int
	CreationTime int
	Attachments  []interface{}

	Geo                    GeoType
	SharedPostText         string
	SharedPostCreationTime int
	SignerId               int
	Tags                   []string

	Author Author
}

type StreamResponse struct {
	Code           int
	Event          StreamEvent
	ServiceMessage ServiceMessage
}

func (c *Client) Stream(credentials StreamingAuth, streamsCount int) {

	var wg sync.WaitGroup
	interruptChannels := make([]chan struct{}, streamsCount)

	for streamId := 0; streamId < streamsCount; streamId++ {
		u := url.URL{
			Scheme:   "wss",
			Host:     credentials.Endpoint,
			Path:     "/stream/",
			RawQuery: "key=" + credentials.Key + "&stream_id=" + strconv.Itoa(streamId),
		}

		log.Printf("connecting to %s\n", u.String())

		c, wsResp, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			if err == websocket.ErrBadHandshake {
				log.Printf("handshake failed with status %d\n", wsResp.StatusCode)
				bodyBuf, _ := ioutil.ReadAll(wsResp.Body)
				log.Println("respBody:", string(bodyBuf))
			}
			log.Printf("stream id: %d, dial error: %s\n", streamId, err.Error())
			continue
		}
		defer c.Close()

		log.Printf("stream id: %d, connection established\n", streamId)

		wg.Add(1)

		go func(id int) {
			done := make(chan struct{})
			defer close(done)

			go func() {
				var streamMessage StreamResponse
				for {
					_, message, err := c.ReadMessage()
					if err != nil {
						log.Printf("stream_id: %d, read: %s\n", id, err.Error())
						done <- struct{}{}
						return
					}
					if err := json.Unmarshal(message, &streamMessage); err != nil {
						log.Println("unmarshal response json failed:", err)
						continue
					}
					log.Printf("recv: %s", message)
					log.Println(streamMessage)
				}
			}()

			interruptChan := make(chan struct{})
			interruptChannels[id] = interruptChan

			select {
			case <-interruptChan:
				err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Printf("stream id: %d, write close error: %s\n", id, err.Error())
					wg.Done()
					return
				}
				select {
				case <-done:
				case <-time.After(time.Second):
				}
				wg.Done()
			case <-done:
				wg.Done()
			}
		}(streamId)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	select {
	case <-interrupt:
		log.Println("interrupt")
		for _, v := range interruptChannels {
			if v != nil {
				v <- struct{}{}
			}
		}
		wg.Wait()
	case <-done:
	}
}
