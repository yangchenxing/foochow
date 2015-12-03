package ipipnet

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	ids = struct {
		places        map[string]placeTree
		isps          map[string]*ISP
		unknownPlaces map[string]bool
		unknownISPs   map[string]bool
	}{
		places:        make(map[string]placeTree),
		isps:          make(map[string]*ISP),
		unknownPlaces: make(map[string]bool),
		unknownISPs:   make(map[string]bool),
	}
)

type placeTree struct {
	*Place
	subplaces map[string]placeTree
}

func loadIDs() error {
	file, err := os.Open(config.idsPath)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		text := scanner.Text()
		record := strings.Split(text, ",")
		switch record[0] {
		case "region":
			id, err := strconv.ParseUint(record[len(record)-1], 10, 64)
			if err != nil {
				return fmt.Errorf("错误的地域ID记录: line=%d, text=%q", lineNum, text)
			}
			if len(record) >= 2 && len(record) <= 5 {
				if err = setPlaceID(id, record[1:len(record)-1]); err != nil {
					return fmt.Errorf("错误的地域ID记录: line=%d, text=%q, error=%q",
						lineNum, text, err.Error())
				}
			} else {
				return fmt.Errorf("错误的地域ID记录: line=%d, text=%q, error=%q",
					lineNum, text, err.Error())
			}
		case "isp":
			id, err := strconv.ParseUint(record[len(record)-1], 10, 32)
			if err != nil {
				return fmt.Errorf("错误的ISP ID记录: line=%d, text=%q, error=%q",
					lineNum, text, err.Error())
			}
			switch len(record) - 1 {
			case 2:
				if _, found := ids.isps[record[1]]; found {
					return fmt.Errorf("重复的ISP ID: line=%d, text=%q", lineNum, text)
				} else {
					ids.isps[record[1]] = &ISP{
						ID:   id,
						Name: record[1],
					}
				}
			default:
				return fmt.Errorf("错误的ISP ID记录: line=%d, text=%q", lineNum, text)
			}
		default:
			return fmt.Errorf("无效的ID记录: line=%d, text=%q", lineNum, text)
		}
	}
	return nil
}

func setPlaceID(id uint64, names []string) error {
	if len(names) > 3 {
		return fmt.Errorf("地域ID深度大于3")
	}
	places := ids.places
	depth := len(names) - 1
	for i := 0; i < depth; i++ {
		if _, found := places[names[i]]; !found {
			return fmt.Errorf("未知地点: %s", strings.Join(names[:i+1], ","))
		} else {
			places = places[names[i]].subplaces
		}
	}
	if _, found := places[names[depth]]; found {
		return fmt.Errorf("重复地域: id=%d, name=%s", id, strings.Join(names, ","))
	}
	place := placeTree{
		Place: &Place{
			ID:   id,
			Name: names[depth],
		},
	}
	if depth < 2 {
		place.subplaces = make(map[string]placeTree)
	}
	places[names[depth]] = place
	return nil
}
