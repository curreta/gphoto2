package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/curreta/gphoto2"
)

func main() {
	fmt.Println("=== Continuous Camera Event Monitor ===")
	fmt.Println("Press Ctrl+C to stop monitoring")
	fmt.Println()

	// Create a new camera connection
	camera, err := gphoto2.NewCamera()
	if err != nil {
		log.Fatalf("Failed to connect to camera: %v", err)
	}
	defer camera.Free()

	fmt.Println("✅ Camera connected successfully!")
	fmt.Println("🔄 Starting continuous event monitoring...")
	fmt.Println("📸 Now take some photos manually with the camera button")
	fmt.Println()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Event counter
	eventCount := 0

	// Monitoring loop
	for {
		select {
		case <-sigChan:
			fmt.Println("\n🛑 Shutdown signal received, stopping monitor...")
			return
		default:
			// Wait for events with 2 second timeout
			event, err := camera.WaitForEvent(2000) // 2000ms = 2 seconds
			if err != nil {
				log.Printf("❌ Error waiting for event: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			if event == nil {
				// No event (timeout) - just continue silently
				continue
			}

			// Event received!
			eventCount++
			timestamp := time.Now().Format("15:04:05")

			fmt.Printf("[%s] Event #%d: ", timestamp, eventCount)

			switch event.Type {
			case gphoto2.EventFileAdded:
				fmt.Printf("📁 FILE ADDED")
				if fileData, ok := event.Data.(*gphoto2.EventFileData); ok {
					fmt.Printf(" - %s/%s", fileData.Folder, fileData.Name)
				}
				fmt.Println()

			case gphoto2.EventFolderAdded:
				fmt.Printf("📂 FOLDER ADDED")
				if fileData, ok := event.Data.(*gphoto2.EventFileData); ok {
					fmt.Printf(" - %s/%s", fileData.Folder, fileData.Name)
				}
				fmt.Println()

			case gphoto2.EventCaptureComplete:
				fmt.Println("✅ CAPTURE COMPLETE")

			case gphoto2.EventFileChanged:
				fmt.Printf("📝 FILE CHANGED")
				if fileData, ok := event.Data.(*gphoto2.EventFileData); ok {
					fmt.Printf(" - %s/%s", fileData.Folder, fileData.Name)
				}
				fmt.Println()

			case gphoto2.EventUnknown:
				fmt.Printf("❓ UNKNOWN EVENT")
				if unknownData, ok := event.Data.(*gphoto2.EventUnknownData); ok {
					fmt.Printf(" - %s", unknownData.Message)
				}
				fmt.Println()

			default:
				fmt.Printf("🔍 UNHANDLED EVENT (Type=%d)", event.Type)
				if event.Data != nil {
					fmt.Printf(" - Data: %v", event.Data)
				}
				fmt.Println()
			}
		}
	}
}
