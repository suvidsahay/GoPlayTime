package timeDB

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type TimeZone struct {
	Location string
	Zone     string
	Offset   int
	ID       string
}

type WriteCounter struct {
	Total uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	fmt.Printf("\r%s", strings.Repeat(" ", 50))

	fmt.Printf("\rDownloading... %d KB complete", wc.Total/1024)
}

func DownloadFile(timeZoneZip string, timeZoneUrl string) error {
	out, err := os.Create(timeZoneZip + ".tmp")
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(timeZoneUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	counter := &WriteCounter{}
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}
	fmt.Println()

	err = os.Rename(timeZoneZip+".tmp", timeZoneZip)
	if err != nil {
		return err
	}
	return nil
}

func ExtractZip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		fpath := filepath.Join(dest, f.Name)

		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

func GetZoneID(zone string, offset int, timeZoneDir string) (string, error) {
	timezoneCSV, err := os.Open(timeZoneDir + "/timezone.csv")
	defer timezoneCSV.Close()
	if err != nil {
		log.Fatal(err)
	}
	r, err := csv.NewReader(timezoneCSV).ReadAll()
	if err != nil {
		log.Println(err)
	}
	for _, record := range r {
		if record[1] == zone && record[3] == strconv.Itoa(offset) {
			return record[0], nil
		}
	}
	return "", fmt.Errorf("%s %d field not found in %s/timezone.csv", zone, offset, timeZoneDir)
}

func GetZoneLocationFromZoneID(zoneID string, timeZoneDir string) (string, error) {
	zoneCSV, err := os.Open(timeZoneDir + "/zone.csv")
	defer zoneCSV.Close()
	if err != nil {
		log.Fatal(err)
	}
	r, err := csv.NewReader(zoneCSV).ReadAll()
	if err != nil {
		log.Println(err)
	}
	for _, record := range r {
		if record[0] == zoneID {
			return record[2], nil
		}
	}
	return "", fmt.Errorf("%s field not found in %s/zone.csv", zoneID, timeZoneDir)
}

func GetUTC(offset int) string {
	hour := offset / 3600
	minutes := (offset % 3600) / 60
	var timeZone string
	if hour < 10 {
		timeZone = "0" + strconv.Itoa(hour)
	} else {
		timeZone = strconv.Itoa(hour)
	}
	timeZone += ":"
	if minutes < 10 {
		timeZone += "0" + strconv.Itoa(minutes)
	} else {
		timeZone += strconv.Itoa(minutes)
	}
	return timeZone
}

func DownloadAndExtractTimeDB() (string, error) {
	timeZoneUrl := "https://timezonedb.com/files/timezonedb.csv.zip"
	timeZoneZip := "timezone.csv.zip"
	timeZoneDir := "timezones"
	if _, err := os.Stat("timezone.csv.zip"); os.IsNotExist(err) {
		er := DownloadFile(timeZoneZip, timeZoneUrl)
		if er != nil {
			return "", er
		}
	}
	if _, err := os.Stat(timeZoneDir + "/zone.csv"); os.IsNotExist(err) {
		_, er := ExtractZip(timeZoneZip, timeZoneDir)
		if er != nil {
			return "", er
		}
	}
	return timeZoneDir, nil
}
