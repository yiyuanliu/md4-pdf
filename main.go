package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
)

var (
	pdfHeader = []byte("%PDF-1.3\n%\xe2\xe3\xcf\xd3\n\n\n1 0 obj\n<</Width 2 0 R/Height 3 0 R/Type 4 0 R/Subtype 5 0 R/Filter 6 0 R/ColorSpace 7 0 R/Length 8 0 R/BitsPerComponent 8>>\nstream\n")
)

func main() {
	str1, _ := hex.DecodeString("17250c98fffe01519e19e0c11121deb73c7ab759c3ab19ae1256f3f69a8aaf141ccf8ef1e5698ffbd939c869a1ebf4993c3c4dab29e64f1bd7c27b5c9de9b8bc")
	str2, _ := hex.DecodeString("17250c98fffe01d19e19e0311121deb73c7ab759c3ab19ae1256f3f69a8aaf141ccf8ef1e5698ffbd939c869a1ebf4993c3c4cab29e64f1bd7c27b5c9de9b8bc")

	jpeg1, _ := ioutil.ReadFile("./1.jpeg")
	jpeg2, _ := ioutil.ReadFile("./2.jpeg")

	out1, out2 := genPic(len(pdfHeader), jpeg1, jpeg2, str1, str2)
	ioutil.WriteFile("out1.jpeg", out1, 0644)
	ioutil.WriteFile("out2.jpeg", out2, 0644)

	pdf1, pdf2 := genPdf(out1, out2)
	ioutil.WriteFile("pdf1.pdf", pdf1, 0644)
	ioutil.WriteFile("pdf2.pdf", pdf2, 0644)
}

func genPic(prefixLen int, jpeg1, jpeg2 []byte, str1, str2 []byte) (a, b []byte) {
	if len(jpeg1) > 65533 || len(jpeg2) > 65533 {
		fmt.Println("jpeg is too big")
	}

	if len(str1) != len(str2) || len(str1) != 64 {
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

	l1 := 64 - (prefixLen % 64)
	if l1 < 6 {
		l1 += 64
	}

	ans := make([]byte, l1)
	ans[0], ans[1] = 0xff, 0xd8
	ans[2], ans[3], ans[4], ans[5] = 0xff, 0xfe, byte((l1-4+commentIdx)>>8), byte((l1-4+commentIdx)&0xff)
	ans = append(ans, make([]byte, 64)...)

	len1, len2 := int(str1[commentIdx+3])+256, int(str2[commentIdx+3])+256
	commentInStr := 64 - commentIdx - 2
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
	copy(ans1[l1:], str1)
	copy(ans2[l1:], str2)

	if len(ans1) != len(ans2) {
		panic("ans with different len")
	}

	return ans1, ans2
}

func getPicSize(pic []byte) (height, width int) {
	return 1920, 1080
}

func genPdf(pic1, pic2 []byte) ([]byte, []byte) {
	height, width := getPicSize(pic1)
	if h2, w2 := getPicSize(pic2); h2 != height || w2 != width {
		fmt.Println("different pic size, set to larger one")
		height = max(h2, height)
		width = max(h2, width)
	}

	var data, xref []byte
	data = []byte("endstream\nendobj\n\n")
	xref = []byte("xref\n0 13 \n0000000000 65535 f \n0000000017 00000 n \n")
	xref = append(xref, []byte(fmt.Sprintf("%010d 00000 n \n", len(pdfHeader)+len(pic1)+len(data)))...)

	// width
	data = append(data, []byte(fmt.Sprintf("2 0 obj\n%010d\nendobj\n\n", width))...)
	xref = append(xref, []byte(fmt.Sprintf("%010d 00000 n \n", len(pdfHeader)+len(pic1)+len(data)))...)
	//height
	data = append(data, []byte(fmt.Sprintf("3 0 obj\n%010d\nendobj\n\n", width))...)
	xref = append(xref, []byte(fmt.Sprintf("%010d 00000 n \n", len(pdfHeader)+len(pic1)+len(data)))...)

	data = append(data, []byte("4 0 obj\n/XObject\nendobj\n\n")...)
	xref = append(xref, []byte(fmt.Sprintf("%010d 00000 n \n", len(pdfHeader)+len(pic1)+len(data)))...)
	data = append(data, []byte("5 0 obj\n/Image\nendobj\n\n")...)
	xref = append(xref, []byte(fmt.Sprintf("%010d 00000 n \n", len(pdfHeader)+len(pic1)+len(data)))...)
	data = append(data, []byte("6 0 obj\n/DCTDecode\nendobj\n\n")...)
	xref = append(xref, []byte(fmt.Sprintf("%010d 00000 n \n", len(pdfHeader)+len(pic1)+len(data)))...)
	data = append(data, []byte("7 0 obj\n/DeviceRGB\nendobj\n\n")...)
	xref = append(xref, []byte(fmt.Sprintf("%010d 00000 n \n", len(pdfHeader)+len(pic1)+len(data)))...)
	data = append(data, []byte(fmt.Sprintf("8 0 obj\n%010d\nendobj\n\n", len(pdfHeader)+len(pic1)))...)
	xref = append(xref, []byte(fmt.Sprintf("%010d 00000 n \n", len(pdfHeader)+len(pic1)+len(data)))...)
	data = append(data, []byte("9 0 obj\n<<\n  /Type /Catalog\n  /Pages 10 0 R\n>>\nendobj\n\n")...)
	xref = append(xref, []byte(fmt.Sprintf("%010d 00000 n \n", len(pdfHeader)+len(pic1)+len(data)))...)
	data = append(data, []byte("10 0 obj\n<<\n  /Type /Pages\n  /Count 1\n  /Kids [11 0 R]\n>>\nendobj\n\n")...)
	xref = append(xref, []byte(fmt.Sprintf("%010d 00000 n \n", len(pdfHeader)+len(pic1)+len(data)))...)
	data = append(data, []byte(fmt.Sprintf("11 0 obj\n<<\n  /Type /Page\n  /Parent 10 0 R\n  /MediaBox [0 0 %010d %010d]\n  /CropBox [0 0 %010d %010d]\n  /Contents 12 0 R\n  /Resources\n  <<\n    /XObject <</Im0 1 0 R>>\n  >>\n>>\nendobj\n\n", width, height, width, height))...)
	xref = append(xref, []byte(fmt.Sprintf("%010d 00000 n \n", len(pdfHeader)+len(pic1)+len(data)))...)
	data = append(data, []byte(fmt.Sprintf("12 0 obj\n<</Length 49>>\nstream\nq\n  %010d 0 0 %010d 0 0 cm\n  /Im0 Do\nQ\nendstream\nendobj\n\n", width, height))...)

	xrefPos := len(pdfHeader) + len(pic1) + len(data)
	tailer := []byte(fmt.Sprintf("\ntrailer << /Root 9 0 R /Size 13>>\n\nstartxref\n%010d\n%%%%EOF\n", xrefPos))

	pdf1 := make([]byte, 0, len(pdfHeader)+len(pic1)+len(data)+len(xref)+len(tailer))
	pdf2 := make([]byte, 0, len(pdfHeader)+len(pic1)+len(data)+len(xref)+len(tailer))

	pdf1 = append(pdf1, pdfHeader...)
	pdf1 = append(pdf1, pic1...)
	pdf1 = append(pdf1, data...)
	pdf1 = append(pdf1, xref...)
	pdf1 = append(pdf1, tailer...)

	pdf2 = append(pdf2, pdfHeader...)
	pdf2 = append(pdf2, pic2...)
	pdf2 = append(pdf2, data...)
	pdf2 = append(pdf2, xref...)
	pdf2 = append(pdf2, tailer...)

	return pdf1, pdf2
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
