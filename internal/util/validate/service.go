package validate

import (
	"fmt"
	"mime/multipart"
)

func Service(title, description, price []string, image []*multipart.FileHeader) error {
	if len(title) == 0 || len(description) == 0 || len(price) == 0 || len(image) == 0 {
		return fmt.Errorf("not all fields are filled in")
	}
	if len(title) > 30 {
		return fmt.Errorf("long title, need no more than 30 characters")
	}
	if len(description) > 255 {
		return fmt.Errorf("long description, need no more than 255 characters")
	}
	return nil
}
