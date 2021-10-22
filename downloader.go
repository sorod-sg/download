package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
)

type Downloader struct {
	concurrency int
}

func NewDownloader(concurrency int) *Downloader {
	return &Downloader{concurrency: concurrency}
}

func (d *Downloader) Download(URL, filename string) error {
	if filename == "" {
		filename = path.Base(URL)
	}
	resp, err := http.Head(URL)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK && resp.Header.Get("Accept-Ranges") == "bytes" {
		return d.multiDownload(URL, "12", int(resp.ContentLength))
	}
	return d.singDownload(URL, "12")
}
func (d *Downloader) multiDownload(URL, filename string, contentlen int) error {
	partSize := contentlen / d.concurrency
	partDir := d.getPartDir(filename)
	os.Mkdir(partDir, 0777)
	defer os.RemoveAll(partDir)
	var wg sync.WaitGroup
	wg.Add(d.concurrency)
	fileHasDone, err := ioutil.ReadDir("./")
	rangstart := 0
	if err != nil {
		log.Fatal(err)
	}
	for _, doneFileName1 := range fileHasDone {
		for _, doneFileName2 := range fileHasDone {
			if doneFileName1 == doneFileName2 {
				rangstart += partSize + 1
			}
		}
	}
	for i := 0; i < d.concurrency; i++ {
		go func(i, start int) {
			defer wg.Done()
			rangeEnd := start + partSize
			if i == d.concurrency-1 {
				rangeEnd = contentlen
			}
			d.downloadPartial(URL, filename, start, rangeEnd, i)
		}(0, rangstart)
		rangstart += partSize + 1
	}
	wg.Wait()
	d.merge(filename)

	return nil

}
func (d *Downloader) singDownload(URL, filename string) error {
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	File, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer File.Close()
	_, err = io.Copy(File, resp.Body)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		log.Fatal(err)
	}
	return nil
}
func (d *Downloader) downloadPartial(URL, filename string, rangeStart, rangEnd, i int) {
	if rangeStart >= rangEnd {
		return
	}
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Range", fmt.Sprintf("byte=%d-%d", rangeStart, rangEnd))
	resp, err := http.DefaultClient.Do(req)
	fmt.Println(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	flags := os.O_CREATE | os.O_WRONLY
	partFile, err := os.OpenFile(d.getPartFilename(filename, i), flags, 0777)
	if err != nil {
		log.Fatal(err)
	}
	defer partFile.Close()
	buf := make([]byte, 32*1024) //buf做下载缓冲区
	_, err = io.CopyBuffer(partFile, resp.Body, buf)
	if err != nil {
		if err == io.EOF {
			return
		}
		log.Fatal(err)
	}

}
func (d *Downloader) getPartDir(filename string) string {
	return strings.SplitN(filename, ".", 1)[0] //删除后缀
}
func (d *Downloader) getPartFilename(filename string, partNum int) string {
	partDir := d.getPartDir(filename)
	return fmt.Sprintf("%s/%s-%d", partDir, filename, partNum)
}

func (d *Downloader) merge(filename string) error {
	destFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer destFile.Close()
	for i := 0; i < d.concurrency; i++ {
		partFileName := d.getPartFilename(filename, i)
		partFile, err := os.Open(partFileName)
		if err != nil {
			return err
		}
		io.Copy(destFile, partFile)
		partFile.Close()
		os.Remove(partFileName)
	}

	return nil
} //合并分开的文件
