// folder_size.go
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	reset  = "\033[0m"
	green  = "\033[92m"
	red    = "\033[91m"
	yellow = "\033[93m"
	blue   = "\033[94m"
)

func colorize(text, color string) string {
	return color + text + reset
}

func humanReadable(size int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	s := float64(size)
	for _, u := range units {
		if s < 1024.0 {
			return fmt.Sprintf("%.1f %s", s, u)
		}
		s /= 1024.0
	}
	return fmt.Sprintf("%.1f PB", s)
}

type Item struct {
	Path string
	Size int64
}

func getFolderSize(path string, recursive bool, depth int, maxDepth int, excludeHidden bool, verbose bool) ([]Item, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", path)
	}

	var result []Item
	var totalSize int64

	entries, err := os.ReadDir(path)
	if err != nil {
		if verbose {
			fmt.Println(colorize("Permission denied: "+path, red))
		}
		return []Item{{Path: path, Size: 0}}, nil
	}

	for _, entry := range entries {
		if excludeHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		fullPath := filepath.Join(path, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if entry.IsDir() {
			if recursive && (maxDepth == 0 || depth < maxDepth) {
				subItems, err := getFolderSize(fullPath, recursive, depth+1, maxDepth, excludeHidden, verbose)
				if err != nil {
					if verbose {
						fmt.Println(colorize("Error reading "+fullPath+": "+err.Error(), red))
					}
					continue
				}
				var subTotal int64
				for _, item := range subItems {
					subTotal += item.Size
				}
				totalSize += subTotal
				result = append(result, subItems...)
			}
		} else {
			size := info.Size()
			totalSize += size
			if verbose {
				fmt.Printf("  %s: %s\n", fullPath, humanReadable(size))
			}
		}
	}
	result = append(result, Item{Path: path, Size: totalSize})
	return result, nil
}

func main() {
	var (
		path          string
		recursive     bool
		human         bool
		sortFlag      bool
		top           int
		depth         int
		excludeHidden bool
		verbose       bool
	)
	flag.StringVar(&path, "p", ".", "Path")
	flag.BoolVar(&recursive, "r", true, "Recursive")
	flag.BoolVar(&human, "h", false, "Human-readable")
	flag.BoolVar(&sortFlag, "s", false, "Sort by size")
	flag.IntVar(&top, "t", 0, "Show top N")
	flag.IntVar(&depth, "d", 0, "Max depth (0=unlimited)")
	flag.BoolVar(&excludeHidden, "exclude-hidden", false, "Exclude hidden")
	flag.BoolVar(&verbose, "v", false, "Verbose")
	flag.Parse()
	if flag.NArg() > 0 {
		path = flag.Arg(0)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Println(colorize("Error: "+err.Error(), red))
		os.Exit(1)
	}

	items, err := getFolderSize(absPath, recursive, 0, depth, excludeHidden, verbose)
	if err != nil {
		fmt.Println(colorize("Error: "+err.Error(), red))
		os.Exit(1)
	}

	// Собираем итоговые размеры папок в map
	sizeMap := make(map[string]int64)
	for _, item := range items {
		sizeMap[item.Path] = item.Size
	}

	// Преобразуем в слайс для сортировки
	var result []Item
	for p, s := range sizeMap {
		result = append(result, Item{Path: p, Size: s})
	}
	if sortFlag {
		sort.Slice(result, func(i, j int) bool {
			return result[i].Size > result[j].Size
		})
	}
	if top > 0 && top < len(result) {
		result = result[:top]
	}

	maxLen := 0
	for _, item := range result {
		if len(item.Path) > maxLen {
			maxLen = len(item.Path)
		}
	}
	for _, item := range result {
		var sizeStr string
		if human {
			sizeStr = humanReadable(item.Size)
		} else {
			sizeStr = fmt.Sprintf("%d B", item.Size)
		}
		color := green
		if item.Size > 1024*1024*1024 {
			color = red
		} else if item.Size > 1024*1024 {
			color = yellow
		}
		fmt.Printf("%s  %s\n", colorize(fmt.Sprintf("%12s", sizeStr), color), item.Path)
	}
}
