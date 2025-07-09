package gphoto2

/** \file
 *
 * \author Copyright 2025 Carlos Urreta
 *
 * \note
 * This library is free software; you can redistribute it and/or
 * modify it under the terms of the GNU Lesser General Public
 * License as published by the Free Software Foundation; either
 * version 2 of the License, or (at your option) any later version.
 *
 * \note
 * This library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * Lesser General Public License for more details.
 *
 * \note
 * You should have received a copy of the GNU Lesser General Public
 * License along with this library; if not, write to the
 * Free Software Foundation, Inc., 51 Franklin Street, Fifth Floor,
 * Boston, MA  02110-1301  USA
 */

// #include <gphoto2/gphoto2.h>
// #include "callbacks.h"
// #include <stdlib.h>
// #include <string.h>
//
// // Wrapper function to handle the CameraEventType casting
// int gp_camera_wait_for_event_wrapper(Camera *camera, int timeout, int *eventtype, void **eventdata, GPContext *context) {
//     return gp_camera_wait_for_event(camera, timeout, (CameraEventType*)eventtype, eventdata, context);
// }
import "C"
import (
	"time"
	"unsafe"
)

// EventType represents the type of camera event
type EventType int

// CameraEvent represents an event from the camera
type CameraEvent struct {
	Type EventType
	Data EventData
}

// EventData is the interface for event-specific data
type EventData interface{}

// EventFileData contains file information for file-related events
type EventFileData struct {
	Folder string
	Name   string
}

// EventUnknownData contains string data for unknown events
type EventUnknownData struct {
	Message string
}

// WaitForEvent waits for camera events with timeout
// Returns nil event on timeout (not an error)
// timeoutMs is timeout in milliseconds
func (c *Camera) WaitForEvent(timeoutMs int) (*CameraEvent, error) {
	if c.gpCamera == nil {
		return nil, newError("Camera disconnected", Error)
	}

	var eventType C.int
	var eventData unsafe.Pointer

	// Call libgphoto2's wait_for_event function using our wrapper
	res := C.gp_camera_wait_for_event_wrapper(c.gpCamera, C.int(timeoutMs), &eventType, &eventData, c.Ctx.gpContext)

	if res != GPOK {
		// Check if it's a timeout (not an error condition)
		if int(res) == ErrorTimeout {
			return nil, nil // Return nil event on timeout
		}
		return nil, newError("Error waiting for camera event", int(res))
	}

	// Convert C event to Go event
	event := &CameraEvent{
		Type: EventType(eventType),
	}

	// Parse event data based on event type
	switch EventType(eventType) {
	case EventFileAdded, EventFolderAdded, EventFileChanged:
		if eventData != nil {
			// eventData is a CameraFilePath for file events
			cPath := (*cameraFilePathInternal)(eventData)
			// Convert to Go CameraFilePath using existing function
			goPath := newCameraFilePath(cPath, c)
			event.Data = &EventFileData{
				Folder: goPath.Folder,
				Name:   goPath.Name,
			}
		}
	case EventUnknown:
		if eventData != nil {
			// eventData is a string for unknown events
			event.Data = &EventUnknownData{
				Message: C.GoString((*C.char)(eventData)),
			}
			// Free the allocated string
			C.free(eventData)
		}
	case EventCaptureComplete:
		// No additional data for capture complete events
		event.Data = nil
	case EventTimeout:
		// This shouldn't happen since we handle timeout above
		return nil, nil
	}

	return event, nil
}

// WaitForEventTimeout waits for camera events with Go duration timeout
// This is a convenience wrapper around WaitForEvent
func (c *Camera) WaitForEventTimeout(timeout time.Duration) (*CameraEvent, error) {
	timeoutMs := int(timeout.Nanoseconds() / 1000000) // Convert to milliseconds
	return c.WaitForEvent(timeoutMs)
}
