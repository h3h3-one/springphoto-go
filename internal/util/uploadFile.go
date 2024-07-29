package util

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log/slog"
	"mime/multipart"
	"springoff/internal/config"
)

type Response struct {
	Data    Data `json:"data"`
	Success bool `json:"success"`
	Status  int  `json:"status"`
}

type Data struct {
	Id   string `json:"id"`
	Link string `json:"link"`
}

func UploadFile(input *multipart.FileHeader, config *config.Config) (*Response, error) {
	f, err := input.Open()
	if err != nil {
		slog.Error("error open file", "error", err, "file name", input.Filename)
		return nil, fmt.Errorf("error open file: %w", err)
	}
	defer f.Close()

	fileContent := make([]byte, input.Size)
	_, err = f.Read(fileContent)
	if err != nil {
		slog.Error("error read file", "error", err, "file name", input.Filename)
		return nil, fmt.Errorf("error read file: %w", err)
	}

	agent := fiber.Post("https://api.imageban.ru/v1")
	agent.Add("Authorization", config.SecretKey)

	base64Image := base64.StdEncoding.EncodeToString(fileContent)
	args := fiber.AcquireArgs()
	args.Set("image", base64Image)
	agent.MultipartForm(args)

	code, body, errors := agent.String()

	if code != 200 {
		slog.Error("Something went wrong", "code", code, "body", body, "errors", errors)
		return nil, errors[0]
	}

	resp := Response{}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		slog.Error("Error unmarshal response json")
		return nil, err
	}

	return &resp, nil
}
