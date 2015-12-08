package ipipnet

import (
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/yangchenxing/foochow/logging"
)

type section struct {
	*Location
	lower uint32
	upper uint32
}

type ipIndex struct {
	sections []section
	index    [256]struct {
		lower int
		upper int
	}
	checksum string
}

type IPIPNetConfig struct {
	IDsPath       string
	DataPath      string
	DataURL       string
	CheckInterval time.Duration
}

var (
	config struct {
		idsPath       string
		dataPath      string
		dataURL       string
		checkInterval time.Duration
	}
	index *ipIndex

	ErrDuplicatedDownload = errors.New("重复下载")
	ErrNotIPv4            = errors.New("非IPv4地址")
	ErrNotData            = errors.New("未加载数据")

	UnknownPlaceCallback func(string)
	UnknownISPCallback   func(string)
)

func Initialize(conf IPIPNetConfig) error {
	config.idsPath = conf.IDsPath
	config.dataPath = conf.DataPath
	config.dataURL = conf.DataURL
	config.checkInterval = conf.CheckInterval
	if err := loadIDs(); err != nil {
		return err
	}
	if err := loadData(); err != nil {
		return err
	}
	go autoReload()
	return nil
}

func Locate(ip net.IP) (*Location, error) {
	if index == nil {
		return nil, ErrNotData
	}
	if ip = ip.To4(); ip == nil {
		return nil, ErrNotIPv4
	}
	v := binary.BigEndian.Uint32([]byte(ip))
	idx := index
	pos := index.index[v>>24]
	for lower, upper := pos.lower, pos.upper; lower <= upper; {
		mid := (lower + upper) / 2
		section := idx.sections[mid]
		if v < section.lower {
			upper = mid - 1
		} else if v > section.upper {
			lower = mid + 1
		} else {
			return section.Location, nil
		}
	}
	return nil, nil
}

func loadData() error {
	if _, err := os.Lstat(config.dataPath); err != nil {
		if err := download(); err != nil {
			return fmt.Errorf("下载数据出错: %s", err.Error())
		}
	}
	data, err := ioutil.ReadFile(config.dataPath)
	if err != nil {
		return fmt.Errorf("读取数据文件出错: path=%q, error=%q", config.dataPath, err.Error())
	}
	textOffset := binary.BigEndian.Uint32(data[:4]) - 1024
	newIndex := &ipIndex{
		sections: make([]section, (textOffset-4-1024)/8),
	}
	startIP := uint32(1)
	for i, offset := 0, uint32(1028); offset < textOffset; i, offset = i+1, offset+8 {
		endIP := binary.BigEndian.Uint32(data[offset : offset+4])
		newIndex.sections[i].lower = startIP
		newIndex.sections[i].upper = endIP
		dataOffset := textOffset + (uint32(data[offset+4]) | uint32(data[offset+5])<<8 | uint32(data[offset+6])<<16)
		dataLength := uint32(data[offset+7])
		newIndex.sections[i].Location = getLocation(data[dataOffset : dataOffset+dataLength])
		startIP = endIP + 1
		newIndex.index[endIP>>24].upper = i
	}
	for i := 1; i < 256; i++ {
		newIndex.index[i].lower = newIndex.index[i-1].upper + 1
	}
	newIndex.checksum = fmt.Sprintf("%x", sha1.Sum(data))
	index = newIndex
	logging.Debug("加载IPIP.net数据完成: checksum=%s", index.checksum)
	return nil
}

func autoReload() {
	for {
		logging.Debug("等待检查IPIP.net数据更新: %s", config.checkInterval)
		time.Sleep(config.checkInterval)
		if err := download(); err == ErrDuplicatedDownload {
			continue
		}
		if err := loadData(); err != nil {
			logging.Error("加载IPIP.net新数据出错: %s", err.Error())
		}
	}
}

func download() error {
	response, err := http.Get(config.dataURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	etag := response.Header.Get("ETag")
	if etag == "" {
		return errors.New("HTTP应答头缺少ETag字段")
	} else if !strings.HasPrefix(etag, "sha1-") {
		return fmt.Errorf("不支持的ETag: %q", etag)
	}
	if index != nil && index.checksum == strings.ToLower(etag[5:]) {
		return ErrDuplicatedDownload
	}
	if file, err := os.OpenFile(config.dataPath+".tmp", os.O_CREATE|os.O_WRONLY, 0755); err != nil {
		return fmt.Errorf("打开临时文件出错: path=%q, error=%q", config.dataPath+".tmp", err.Error())
	} else if _, err := io.Copy(file, response.Body); err != nil {
		file.Close()
		return fmt.Errorf("下载内容出错: path=%q, error=%q", config.dataPath+".tmp", err.Error())
	} else {
		file.Close()
	}
	if err := os.Rename(config.dataPath+".tmp", config.dataPath); err != nil {
		return fmt.Errorf("重命名临时文件出错: oldpath=%q, newpath=%q, error=%q",
			config.dataPath+".tmp", config.dataPath, err.Error())
	}
	logging.Debug("下载IPIP.net数据完成: checksum=%s", etag[5:])
	return nil
}

func getLocation(data []byte) *Location {
	fields := strings.Split(string(data), "\t")
	placeNames := fields[:3]
	ispNames := strings.Split(fields[len(fields)-1], "/")
	location := new(Location)
	// 获取地点
	places := ids.places
	for i, name := range placeNames {
		if i > 2 || name == "" || (i == 1 && name == placeNames[0]) {
			break
		}
		if len(places) == 0 {
			break
		}
		place, found := places[name]
		if !found {
			if key := strings.Join(placeNames[:+1], "/"); !ids.unknownPlaces[key] {
				ids.unknownPlaces[key] = true
				if UnknownPlaceCallback != nil {
					UnknownPlaceCallback(key)
				}
			}
			break
		}
		switch i {
		case 0:
			location.Country = place.Place
		case 1:
			location.Province = place.Place
		case 2:
			location.City = place.Place
		}
		places = place.subplaces
	}
	location.ISPs = make([]*ISP, 0, len(ispNames))
	for _, name := range ispNames {
		if name == "" {
			continue
		}
		isp := ids.isps[name]
		if isp != nil {
			location.ISPs = append(location.ISPs, isp)
		} else {
			ids.unknownISPs[name] = true
			if UnknownISPCallback != nil {
				UnknownISPCallback(name)
			}
		}
	}
	return location
}
