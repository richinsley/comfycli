package pkg

import (
	"fmt"

	"github.com/richinsley/comfy2go/client"
	"github.com/richinsley/comfy2go/graphapi"
)

// Property setter must return true if they read the value for the property from a pipe.  This is to imply that that
// there may be a cycle where the operation that led to the value being read from a pipe may be repeated until no
// more values can be ready from the pipe.

func SetGenericPropertValue(c *client.ComfyClient, options *ComfyOptions, prop graphapi.Property, value interface{}) (bool, error) {
	err := prop.SetValue(value)
	if err != nil {
		return false, err
	}
	return false, nil
}

func SetfileUploadPropertyValue(c *client.ComfyClient, options *ComfyOptions, prop graphapi.Property, value interface{}) (bool, error) {
	// the ImageUploadProperty value is not directly settable.  We need to pass the property to the call to client.UploadImage
	uploadprop, _ := prop.ToImageUploadProperty()

	// convert value to a string
	filename, ok := value.(string)
	if !ok {
		return false, fmt.Errorf("expected string value for file upload property")
	}

	if filename == "-" {
		// if the filename is "-" then we read an image from stdin
		img, err := ExpectImage(options.GetStdinReader())
		if err != nil {
			return false, err
		}

		_, err = c.UploadImage(img, "image_"+c.ClientID()+".png", true, client.InputImageType, "", uploadprop)
		if err != nil {
			return false, err
		}
		return true, nil
	} else {
		// because we set it to not overwrite existing, the returned filename may
		// be different than the one we provided
		_, err := c.UploadFileFromPath(filename, true, client.InputImageType, "", uploadprop)
		if err != nil {
			return false, err
		}
		return false, nil
	}
}
