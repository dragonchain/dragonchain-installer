package downloader

import (
	"errors"
	"io"
	"net/http"
	"os"
)

// DownloadFile downloads a file from url to filepath
func DownloadFile(filepath string, url string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return errors.New("Error creating file " + filepath + ":\n" + err.Error())
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return errors.New("Error retrieving data from " + url + ":\n" + err.Error())
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return errors.New("Error copying data from " + url + " to " + filepath + ":\n" + err.Error())
	}
	return nil
}
