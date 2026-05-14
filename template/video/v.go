// Package video: 这里横竖视频看长宽比，这里全局认为长宽比大于3/4即为横视频
package video

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/wield18/all-in-one/template/constant"
)

var (
	widthDevideHeight   = 0.74 // 横竖视频界限
	V                   = constant.V
	H                   = constant.H
	errInvaildParameter = errors.New("Invaild Parameter: 非法参数")
)

type FFProbeOutput struct {
	Streams []struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"streams"`
}

func getVideoSize(filePath string) (width, height int, err error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		filePath)

	output, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	var data FFProbeOutput
	if err := json.Unmarshal(output, &data); err != nil {
		return 0, 0, err
	}

	// 找到视频流（通常第一个）
	for _, stream := range data.Streams {
		if stream.Width > 0 && stream.Height > 0 {
			return stream.Width, stream.Height, nil
		}
	}

	return 0, 0, fmt.Errorf("no video stream found")
}

func getAllFile(root string) ([]string, error) {
	var allFiles []string

	// 遍历目录及其所有子目录
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 如果是文件（不是目录）
		if !info.IsDir() {
			allFiles = append(allFiles, path)
		}

		return nil
	})
	if err != nil {
		return make([]string, 0), err
	}
	return allFiles, nil
}

type FileInfo struct {
	Width    int
	Height   int
	FilePath string
}

func GetAllFileInfo(root string) (map[string][]FileInfo, error) {
	infos := make(map[string][]FileInfo, 0)
	allFilePath, err := getAllFile(root)
	if err != nil {
		var t map[string][]FileInfo
		return t, err
	}
	for _, onePath := range allFilePath {
		width, height, err := getVideoSize(onePath)
		if err != nil {
			fmt.Println("err: ", err)
			continue
		}
		if (float64(width) / float64(height)) >= widthDevideHeight {
			infos[H] = append(infos[H], FileInfo{
				Width:    width,
				Height:   height,
				FilePath: onePath,
			})
		} else {
			infos[V] = append(infos[V], FileInfo{
				Width:    width,
				Height:   height,
				FilePath: onePath,
			})
		}
	}
	return infos, err
}

func GetRandomVideoFromSlice(infos []FileInfo, count int) ([]FileInfo, error) {
	l := len(infos)
	if l == 0 || count == 0 {
		return nil, errInvaildParameter
	}

	if count >= l {
		for i := l; i < count; i++ {
			infos = append(infos, infos[0])
		}
		return infos, nil
	}

	// 复制一份，避免修改原切片
	shuffled := make([]FileInfo, l)
	copy(shuffled, infos)

	// 随机打乱
	rand.Shuffle(l, func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// 取前 count 个
	return shuffled[:count], nil
}
