package main

import (
	"fmt"
	"io/ioutil"
)

func main() {
	str1, str2 := make([]byte, 128), make([]byte, 128)
	str1[4], str1[5], str1[6], str1[7] = 0xff, 0xfe, 0x01, 0x20
	str2[4], str2[5], str2[6], str2[7] = 0xff, 0xfe, 0x01, 0x29

	jpeg1, _ := ioutil.ReadFile("./1.jpeg")
	jpeg2, _ := ioutil.ReadFile("./2.jpeg")

	out1, out2 := genPic(0, jpeg1, jpeg2, str1, str2)
	ioutil.WriteFile("out1.jpeg", out1, 0644)
	ioutil.WriteFile("out2.jpeg", out2, 0644)
}

func genPic(prefixLen int, jpeg1, jpeg2 []byte, str1, str2 []byte) (a, b []byte) {
	fmt.Println(len(jpeg1), len(jpeg2))

	if len(str1) != len(str2) || len(str1) != 128 {
		panic("len(str1) != len(str2)")
	}

	commentIdx := -1
	for i := range str1 {
		isComment := func(str []byte, i int) bool { return str[i] == 0xff && str[i+1] == 0xfe && str[i+2] == 0x01 }
		if i+3 < len(str1) && isComment(str1, i) && isComment(str2, i) {
			commentIdx = i
		}
	}
	if commentIdx == -1 {
		panic("check str1 str2")
	}

	l1 := 128 - (prefixLen % 128)
	if l1 < 6 {
		l1 += 128
	}

	ans := make([]byte, l1)
	ans[0], ans[1] = 0xff, 0xd8
	ans[2], ans[3], ans[4], ans[5] = 0xff, 0xfe, byte((l1-4+commentIdx)>>8), byte((l1-4+commentIdx)&0xff)
	ans = append(ans, make([]byte, 128)...)

	len1, len2 := int(str1[commentIdx+3])+128, int(str2[commentIdx+3])+128
	commentInStr := 128 - commentIdx - 2
	len1, len2 = len1-commentInStr, len2-commentInStr
	diff := max(len1, len2) - min(len1, len2)
	if diff < 4 {
		panic("len diff too small")
	}
	ans = append(ans, make([]byte, min(len1, len2))...)
	ans = append(ans, 0xff, 0xfe, byte((diff+2)>>8), byte((diff+2)&0xff))
	ans = append(ans, make([]byte, diff-4)...)
	ans = append(ans, 0xff, 0xfe, byte((len(jpeg1[2:])+2)>>8), byte((len(jpeg1[2:])+2)&0xff))
	ans = append(ans, jpeg1[2:]...)
	ans = append(ans, jpeg2[2:]...)

	ans1, ans2 := make([]byte, len(ans)), make([]byte, len(ans))
	copy(ans1, ans)
	copy(ans2, ans)
	//copy(ans1[l1:], str1)
	//copy(ans2[l1:], str2)
	return ans1, ans2
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func min(a, b int) int {
	if a > b {
		return b
	}

	return a
}
