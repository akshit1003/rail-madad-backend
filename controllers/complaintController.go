package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"rail-madad/config"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"

)

type PNRRequest struct {
	PNR     string                `json:"pnr"`
	Subject string                `json:"subject"`
	Image   *multipart.FileHeader `form:"image"`
}

const (
	API_URL = "https://api-inference.huggingface.co/models/Salesforce/blip-image-captioning-large"
	API_KEY = "Bearer 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"
)

func queryImageCaption(imageData []byte) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", API_URL, bytes.NewBuffer(imageData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", API_KEY)
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Hugging Face API error: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Log the full response body for debugging
	log.Printf("Hugging Face API Response: %s", body)

	var response []map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	if len(response) == 0 {
		return "", fmt.Errorf("no response received")
	}

	// Access generated_text instead of caption
	caption, ok := response[0]["generated_text"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response format: %v", response[0])
	}
	return caption, nil
}



func UploadImageToBucket(file *multipart.FileHeader) (string, error) {
	ctx := context.Background()

	bucketName := os.Getenv("GCS_BUCKET_NAME")
	if bucketName == "" {
		return "", fmt.Errorf("bucket name is not configured")
	}

	bucket := config.StorageClient.Bucket(bucketName)

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	data, err := ioutil.ReadAll(src)
	if err != nil {
		return "", err
	}

	obj := bucket.Object("images/" + file.Filename)
	writer := obj.NewWriter(ctx)
	defer writer.Close()

	if _, err := writer.Write(data); err != nil {
		return "", err
	}

	return fmt.Sprintf("https://storage.googleapis.com/%s/images/%s", bucketName, file.Filename), nil
}

func SubmitPNR(c *fiber.Ctx) error {
	var pnrRequest PNRRequest

	pnrRequest.PNR = c.FormValue("pnr")
	pnrRequest.Subject = c.FormValue("subject")
	image, err := c.FormFile("image")
	if err != nil {
		log.Printf("Error getting image: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to get image from request",
		})
	}
	pnrRequest.Image = image

	log.Printf("Received PNR: %s", pnrRequest.PNR)
	log.Println("Image provided in the request")

	imageURL, err := UploadImageToBucket(pnrRequest.Image)
	if err != nil {
		log.Printf("Error uploading image: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to upload image",
		})
	}

	src, err := image.Open()
	if err != nil {
		log.Printf("Error opening image file: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to open image for captioning",
		})
	}
	defer src.Close()

	imageData, err := ioutil.ReadAll(src)
	if err != nil {
		log.Printf("Error reading image file: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read image for captioning",
		})
	}

	queryGenerated, err := queryImageCaption(imageData)
	if err != nil {
		log.Printf("Error querying image caption: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate image caption",
		})
	}

	_, err = config.FirestoreClient.Collection("pnrs").Doc(pnrRequest.PNR).Set(context.Background(), map[string]interface{}{
		"pnr":           pnrRequest.PNR,
		"subject":       pnrRequest.Subject,
		"image":         imageURL,
		"queryGenerated": queryGenerated, 
		"status":        "Pending",
	})
	if err != nil {
		log.Printf("Error storing PNR: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to store PNR",
		})
	}

	return c.JSON(fiber.Map{
		"message":       "PNR submitted successfully",
		"pnr":          pnrRequest.PNR,
		"image":        imageURL,
		"queryGenerated": queryGenerated, 
		"subject":      pnrRequest.Subject,
		"status":       "Pending",
	})
}

func GetComplaints(c *fiber.Ctx) error {
	pnr := c.Params("pnr")
	if pnr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "PNR is required",
		})
	}

	doc, err := config.FirestoreClient.Collection("pnrs").Doc(pnr).Get(context.Background())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "PNR not found",
			})
		}
		log.Printf("Error fetching PNR: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch PNR",
		})
	}

	var pnrData map[string]interface{}
	if err := doc.DataTo(&pnrData); err != nil {
		log.Printf("Error converting PNR data: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process PNR data",
		})
	}

	return c.JSON(fiber.Map{
		"pnr":           pnr,
		"subject":       pnrData["subject"],
		"queryGenerated": pnrData["queryGenerated"],
		"image":         pnrData["image"],
		"status":        pnrData["status"],
	})
}


