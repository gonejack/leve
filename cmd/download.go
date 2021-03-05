package cmd

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

var client = http.Client{}

func download(file *os.File, imageRef string) (err error) {
	defer file.Close()

	timeout, cancel := timeout(time.Minute * 2)
	defer cancel()

	req, err := http.NewRequestWithContext(timeout, http.MethodGet, imageRef, nil)
	if err != nil {
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if verbose {
		bar := progressbar.NewOptions64(resp.ContentLength,
			progressbar.OptionSetTheme(progressbar.Theme{Saucer: "=", SaucerPadding: ".", BarStart: "|", BarEnd: "|"}),
			progressbar.OptionSetWidth(10),
			progressbar.OptionSpinnerType(11),
			progressbar.OptionShowBytes(true),
			progressbar.OptionShowCount(),
			progressbar.OptionSetPredictTime(false),
			progressbar.OptionSetDescription(imageRef),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionClearOnFinish(),
		)
		_, err = io.Copy(io.MultiWriter(file, bar), resp.Body)
	} else {
		_, err = io.Copy(file, resp.Body)
	}

	return
}
