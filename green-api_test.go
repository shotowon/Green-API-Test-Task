package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
)

type Config struct {
	APIUrl           string `json:"apiUrl"`
	MediaUrl         string `json:"mediaUrl"`
	IDInstance       string `json:"idInstance"`
	APITokenInstance string `json:"apiTokenInstance"`
}

var config Config

func loadConfig() error {
	file, err := os.Open("config.json")
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, &config)
}

type SendMessageInput struct {
	ChatId  string `json:"chatId"`
	Message string `json:"message"`
}

type GetChatHistoryInput struct {
	ChatId string `json:"chatId"`
	Count  int    `json:"count"`
}

type SendMessageTestCase struct {
	Input    SendMessageInput
	Expected ExpectedResult
}

type GetChatHistoryTestCase struct {
	Input    GetChatHistoryInput
	Expected ExpectedResult
}

type ExpectedResult struct {
	Status   int
	Contains []string
}

var sendMessageTests = []SendMessageTestCase{
	{
		Input: SendMessageInput{
			ChatId: "invalid id",
		},
		Expected: ExpectedResult{
			Status:   400,
			Contains: []string{"Validation failed", "'chatId'"},
		},
	},
	{
		Input: SendMessageInput{
			ChatId:  "77083674713@c.us",
			Message: "",
		},
		Expected: ExpectedResult{
			Status:   400,
			Contains: []string{"Validation failed", "'message'"},
		},
	},
	{
		Input: SendMessageInput{
			ChatId:  "77083674713@c.us",
			Message: "workin??",
		},
		Expected: ExpectedResult{
			Status: 200,
		},
	},
}

var getChatHistoryTests = []GetChatHistoryTestCase{
	{
		Input: GetChatHistoryInput{
			ChatId: "invalid id",
		},
		Expected: ExpectedResult{
			Status:   400,
			Contains: []string{"Validation failed", "'chatId'"},
		},
	},
	{
		Input: GetChatHistoryInput{
			ChatId: "77083674713@c.us",
			Count:  1,
		},
		Expected: ExpectedResult{
			Status: 200,
		},
	},
}

func ReqURLFromMethod(method string) string {
	return config.APIUrl + fmt.Sprintf("/waInstance%s", config.IDInstance) + fmt.Sprintf("/%s", method) + fmt.Sprintf("/%s", config.APITokenInstance)
}

func sendMessage(chatId string, message string) (int, string) {
	reqUrl := ReqURLFromMethod("sendMessage")
	payload := SendMessageInput{ChatId: chatId, Message: message}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(reqUrl, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return 500, err.Error()
	}
	defer resp.Body.Close()

	return resp.StatusCode, ""
}

func getChatHistory(chatId string, count int) (int, string) {
	reqUrl := ReqURLFromMethod("getChatHistory")
	payload := GetChatHistoryInput{ChatId: chatId, Count: count}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(reqUrl, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return 500, err.Error()
	}
	defer resp.Body.Close()

	return resp.StatusCode, ""
}

func TestAPIFunctions(t *testing.T) {
	if err := loadConfig(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	for _, testCase := range sendMessageTests {
		status, _ := sendMessage(testCase.Input.ChatId, testCase.Input.Message)
		if status == http.StatusTooManyRequests {
			t.Errorf("sendMessage: Too many requests, try again")
			continue
		}

		if status == http.StatusInternalServerError {
			t.Errorf("sendMessage: Internal Server error, try again")
			continue
		}
		if status != testCase.Expected.Status && status != http.StatusTooManyRequests {
			t.Errorf("sendMessage: expected status %d, got %d", testCase.Expected.Status, status)
		}
	}

	for _, testCase := range getChatHistoryTests {
		status, _ := getChatHistory(testCase.Input.ChatId, testCase.Input.Count)
		if status == http.StatusTooManyRequests {
			t.Errorf("getChatHistory: Too many requests, try again")
			continue
		}

		if status == http.StatusInternalServerError {
			t.Errorf("getChatHistory: Internal Server error, try again")
			continue
		}

		if status != testCase.Expected.Status && status != http.StatusTooManyRequests {
			t.Errorf("getChatHistory: expected status %d, got %d", testCase.Expected.Status, status)
			continue
		}
	}
}
