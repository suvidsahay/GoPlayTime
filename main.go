package main

import (
	"fmt"
	"github.com/suvidsahay/InvideTest/timeDB"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

func main() {
	var zone timeDB.TimeZone
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		fileData, err := os.Readlink("/etc/localtime")
		if err != nil {
			log.Println(err)
		}
		lines := strings.Split(fileData, "/")
		zone.Location = lines[len(lines) - 2] + "/" + lines[len(lines) - 1]
		location, err := time.LoadLocation(zone.Location)
		zone.Zone, zone.Offset = time.Now().In(location).Zone()
	} else {
		var err error
		timeZoneDir, err := timeDB.DownloadAndExtractTimeDB()
		if err != nil {
			log.Panic(err)
		}
		zone.Zone, zone.Offset = time.Now().Local().Zone()
		zone.ID, err = timeDB.GetZoneID(zone.Zone, zone.Offset, timeZoneDir)
		if err != nil {
			log.Panic(err)
		}
		zone.Location, err = timeDB.GetZoneLocationFromZoneID(zone.ID, timeZoneDir)
	}
	fmt.Printf("%s UTC ", zone.Location)
	if zone.Offset >= 0 {
		fmt.Printf("+%s \n", timeDB.GetUTC(zone.Offset))
	} else {
		zone.Offset *= -1
		fmt.Printf("-%s \n", timeDB.GetUTC(zone.Offset))
	}
}