package files

import (
	"fmt"
	"net/http"
	"os"
)

func DownloadFromHTTP(URL, username, password, key string) (*os.File, error) {
	client := &http.Client{}
	downloadURL := fmt.Sprintf("%s/%s", URL, key)
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(username, password)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file: %s", resp.Status)
	}

	localFile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("%s-*", key))
	if err != nil {
		return nil, err
	}

	_, err = localFile.ReadFrom(resp.Body)
	if err != nil {
		_ = localFile.Close()
		return nil, err
	}

	err = resp.Body.Close()
	if err != nil {
		_ = localFile.Close()
		return nil, err
	}

	return localFile, localFile.Close()
}
