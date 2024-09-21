package controllers

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"mime/multipart"

	"github.com/gofiber/fiber/v2"
	"rail-madad/config"
)

type PNRRequest struct {
	PNR     string                `json:"pnr"`
	Subject string                `json:"subject"`
	Image   *multipart.FileHeader `form:"image"`
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

	_, err = config.FirestoreClient.Collection("pnrs").Doc(pnrRequest.PNR).Set(context.Background(), map[string]interface{}{
		"pnr":     pnrRequest.PNR,
		"subject": pnrRequest.Subject,
		"image":   imageURL,
		"status":  "Pending",
	})
	if err != nil {
		log.Printf("Error storing PNR: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to store PNR",
		})
	}

	return c.JSON(fiber.Map{
		"message": "PNR submitted successfully",
		"pnr":     pnrRequest.PNR,
		"image":   imageURL,
		"subject": pnrRequest.Subject,
		"status":  "Pending",
	})
}
