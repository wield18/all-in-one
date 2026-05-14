package video

import (
	"fmt"
	"testing"
)

func TestMain(t *testing.T) {
	w, h, err := getVideoSize("C:/Games/Videos/The Other/Diives_Videos/ani/V/01_BaoziVSXingyun_Buried_Bone.mp4")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Video size: %dx%d\n", w, h)
}

// 不修改
func TestSlice(t *testing.T) {
	aSlice := []int{1, 2, 3}
	appendSlice(&aSlice, 1)
	fmt.Println(aSlice)

}
func appendSlice(aSlice *[]int, minLen int) {
	// if len(aSlice) < minLen {
	// 	for i := len(aSlice); i < minLen; i++ {
	// 		aSlice = append(aSlice, aSlice[0])
	// 	}
	// }
	// return aSlice
	*aSlice = (*aSlice)[1:]
}

func TestGetAllFileInfo(t *testing.T) {
	root := `D:\Videos\Yuumeilyn`
	infos, _ := GetAllFileInfo(root)
	for _, v := range infos {
		for _, oneInfo := range v {
			fmt.Println(oneInfo)
		}
	}
}
