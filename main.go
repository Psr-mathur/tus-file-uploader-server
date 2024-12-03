package main

import (
	utils "go-tus-server/utils"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/tus/tusd/v2/pkg/filestore"
	"github.com/tus/tusd/v2/pkg/handler"
	tusd "github.com/tus/tusd/v2/pkg/handler"
)

func main() {
	// Step 1: Initialize Gin
	router := gin.Default()

	// Step 1: Set up CORS middleware
	// router.Use(cors.Default()) // Allows all origins by default

	// Alternatively, for custom CORS configuration:
	router.Use(cors.New(cors.Config{
		AllowCredentials: true, // Allow credentials (cookies, etc.)
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
	}))

	// Step 2: Set up the upload and final directories
	tempUploadDir := "C:\\TUS-Server\\uploads"

	if _, err := os.Stat(tempUploadDir); os.IsNotExist(err) {
		os.MkdirAll(tempUploadDir, os.ModePerm)
	}

	// Step 3: Configure TUS server
	store := filestore.FileStore{
		Path: tempUploadDir,
	}

	composer := tusd.NewStoreComposer()
	store.UseIn(composer)

	tusHandler, err := tusd.NewHandler(tusd.Config{
		BasePath:                "/tus/upload/",
		StoreComposer:           composer,
		NotifyCompleteUploads:   true,
		RespectForwardedHeaders: true,
		PreUploadCreateCallback: func(info tusd.HookEvent) (tusd.HTTPResponse, tusd.FileInfoChanges, error) {
			// Generate a custom file name
			newFileNameWithExt := info.HTTPRequest.Header.Get("FileName")
			if newFileNameWithExt == "" {
				newFileNameWithExt = info.Upload.MetaData["filename"]
			}
			info.Upload.MetaData["filename"] = newFileNameWithExt
			return tusd.HTTPResponse{}, tusd.FileInfoChanges{
				MetaData: info.Upload.MetaData,
				ID:       newFileNameWithExt,
			}, nil
		},
		// Cors: &tusd.CorsConfig{
		// 	Disable:          true,
		// 	AllowCredentials: true,
		// },
	})
	if err != nil {
		log.Fatalf("Failed to create TUS handler: %v", err)
	}

	// Step 4: Listen for completed uploads
	go func() {
		for event := range tusHandler.CompleteUploads {
			go func(event handler.HookEvent) {
				fileNameWithExt := event.Upload.ID
				finalDir := event.HTTPRequest.Header.Get("FinalDir")
				if finalDir == "" {
					log.Printf("No final directory provided for upload %s", fileNameWithExt)
					return
				}

				// Create final directory if it doesn't exist
				if _, err := os.Stat(finalDir); os.IsNotExist(err) {
					os.MkdirAll(finalDir, os.ModePerm)
				}

				// Extract file name and move to final directory
				err = utils.MoveFile(fileNameWithExt, tempUploadDir, finalDir)
				if err != nil {
					log.Printf("Failed to move file %s: %v", fileNameWithExt, err)
				} else {
					log.Printf("File %s successfully moved to %s", fileNameWithExt, finalDir)
				}
			}(event)
		}
	}()

	// Step 5: Add routes
	router.GET("/tus/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Tus server is running",
		})
	})

	// router.Any("/tus/upload/*path", func(c *gin.Context) {
	// 	tusHandler.ServeHTTP(c.Writer, c.Request)
	// })

	router.Any("/tus/upload/*path", gin.WrapH(http.StripPrefix("/tus/upload/", tusHandler)))
	router.Any("/tus/upload", gin.WrapH(http.StripPrefix("/tus/upload", tusHandler)))

	// Step 6: Start the server
	port := ":3456"
	log.Printf("Starting Gin TUS server on port %s...", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
