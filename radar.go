package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"sort"
)

const (
	REDUCE_MATRIX_SZ = 8
)

// n0q raster values: https://mesonet.agron.iastate.edu/GIS/rasters.php?rid=2

var n0qRaster = map[int]float64{
	0:   -9999, // idx == 0 "empty pixel". Fake dBz value
	1:   -32,
	2:   -31.5,
	3:   -31,
	4:   -30.5,
	5:   -30,
	6:   -29.5,
	7:   -29,
	8:   -28.5,
	9:   -28,
	10:  -27.5,
	11:  -27,
	12:  -26.5,
	13:  -26,
	14:  -25.5,
	15:  -25,
	16:  -24.5,
	17:  -24,
	18:  -23.5,
	19:  -23,
	20:  -22.5,
	21:  -22,
	22:  -21.5,
	23:  -21,
	24:  -20.5,
	25:  -20,
	26:  -19.5,
	27:  -19,
	28:  -18.5,
	29:  -18,
	30:  -17.5,
	31:  -17,
	32:  -16.5,
	33:  -16,
	34:  -15.5,
	35:  -15,
	36:  -14.5,
	37:  -14,
	38:  -13.5,
	39:  -13,
	40:  -12.5,
	41:  -12,
	42:  -11.5,
	43:  -11,
	44:  -10.5,
	45:  -10,
	46:  -9.5,
	47:  -9,
	48:  -8.5,
	49:  -8,
	50:  -7.5,
	51:  -7,
	52:  -6.5,
	53:  -6,
	54:  -5.5,
	55:  -5,
	56:  -4.5,
	57:  -4,
	58:  -3.5,
	59:  -3,
	60:  -2.5,
	61:  -2,
	62:  -1.5,
	63:  -1,
	64:  -0.5,
	65:  0,
	66:  0.5,
	67:  1,
	68:  1.5,
	69:  2,
	70:  2.5,
	71:  3,
	72:  3.5,
	73:  4,
	74:  4.5,
	75:  5,
	76:  5.5,
	77:  6,
	78:  6.5,
	79:  7,
	80:  7.5,
	81:  8,
	82:  8.5,
	83:  9,
	84:  9.5,
	85:  10,
	86:  10.5,
	87:  11,
	88:  11.5,
	89:  12,
	90:  12.5,
	91:  13,
	92:  13.5,
	93:  14,
	94:  14.5,
	95:  15,
	96:  15.5,
	97:  16,
	98:  16.5,
	99:  17,
	100: 17.5,
	101: 18,
	102: 18.5,
	103: 19,
	104: 19.5,
	105: 20,
	106: 20.5,
	107: 21,
	108: 21.5,
	109: 22,
	110: 22.5,
	111: 23,
	112: 23.5,
	113: 24,
	114: 24.5,
	115: 25,
	116: 25.5,
	117: 26,
	118: 26.5,
	119: 27,
	120: 27.5,
	121: 28,
	122: 28.5,
	123: 29,
	124: 29.5,
	125: 30,
	126: 30.5,
	127: 31,
	128: 31.5,
	129: 32,
	130: 32.5,
	131: 33,
	132: 33.5,
	133: 34,
	134: 34.5,
	135: 35,
	136: 35.5,
	137: 36,
	138: 36.5,
	139: 37,
	140: 37.5,
	141: 38,
	142: 38.5,
	143: 39,
	144: 39.5,
	145: 40,
	146: 40.5,
	147: 41,
	148: 41.5,
	149: 42,
	150: 42.5,
	151: 43,
	152: 43.5,
	153: 44,
	154: 44.5,
	155: 45,
	156: 45.5,
	157: 46,
	158: 46.5,
	159: 47,
	160: 47.5,
	161: 48,
	162: 48.5,
	163: 49,
	164: 49.5,
	165: 50,
	166: 50.5,
	167: 51,
	168: 51.5,
	169: 52,
	170: 52.5,
	171: 53,
	172: 53.5,
	173: 54,
	174: 54.5,
	175: 55,
	176: 55.5,
	177: 56,
	178: 56.5,
	179: 57,
	180: 57.5,
	181: 58,
	182: 58.5,
	183: 59,
	184: 59.5,
	185: 60,
	186: 60.5,
	187: 61,
	188: 61.5,
	189: 62,
	190: 62.5,
	191: 63,
	192: 63.5,
	193: 64,
	194: 64.5,
	195: 65,
	196: 65.5,
	197: 66,
	198: 66.5,
	199: 67,
	200: 67.5,
	201: 68,
	202: 68.5,
	203: 69,
	204: 69.5,
	205: 70,
	206: 70.5,
	207: 71,
	208: 71.5,
	209: 72,
	210: 72.5,
	211: 73,
	212: 73.5,
	213: 74,
	214: 74.5,
	215: 75,
	216: 75.5,
	217: 76,
	218: 76.5,
	219: 77,
	220: 77.5,
	221: 78,
	222: 78.5,
	223: 79,
	224: 79.5,
	225: 80,
	226: 80.5,
	227: 81,
	228: 81.5,
	229: 82,
	230: 82.5,
	231: 83,
	232: 83.5,
	233: 84,
	234: 84.5,
	235: 85,
	236: 85.5,
	237: 86,
	238: 86.5,
	239: 87,
	240: 87.5,
	241: 88,
	242: 88.5,
	243: 89,
	244: 89.5,
	245: 90,
	246: 90.5,
	247: 91,
	248: 91.5,
	249: 92,
	250: 92.5,
	251: 93,
	252: 93.5,
	253: 94,
	254: 94.5,
}

var n0qPalette = color.Palette{
	color.RGBA{0, 0, 0, 0},
	color.RGBA{133, 113, 143, 255},
	color.RGBA{133, 114, 143, 255},
	color.RGBA{134, 115, 141, 255},
	color.RGBA{135, 117, 139, 255},
	color.RGBA{135, 118, 139, 255},
	color.RGBA{136, 119, 137, 255},
	color.RGBA{137, 121, 135, 255},
	color.RGBA{137, 122, 135, 255},
	color.RGBA{138, 123, 133, 255},
	color.RGBA{139, 125, 132, 255},
	color.RGBA{139, 126, 132, 255},
	color.RGBA{140, 127, 130, 255},
	color.RGBA{141, 129, 128, 255},
	color.RGBA{141, 130, 128, 255},
	color.RGBA{142, 131, 126, 255},
	color.RGBA{143, 132, 124, 255},
	color.RGBA{143, 133, 124, 255},
	color.RGBA{144, 135, 123, 255},
	color.RGBA{145, 136, 121, 255},
	color.RGBA{145, 137, 121, 255},
	color.RGBA{146, 139, 119, 255},
	color.RGBA{147, 141, 117, 255},
	color.RGBA{150, 145, 83, 255},
	color.RGBA{152, 148, 87, 255},
	color.RGBA{155, 151, 91, 255},
	color.RGBA{157, 154, 96, 255},
	color.RGBA{160, 157, 100, 255},
	color.RGBA{163, 160, 104, 255},
	color.RGBA{165, 163, 109, 255},
	color.RGBA{168, 166, 113, 255},
	color.RGBA{170, 169, 118, 255},
	color.RGBA{173, 172, 122, 255},
	color.RGBA{176, 175, 126, 255},
	color.RGBA{178, 178, 131, 255},
	color.RGBA{183, 184, 140, 255},
	color.RGBA{186, 187, 144, 255},
	color.RGBA{189, 190, 148, 255},
	color.RGBA{191, 193, 153, 255},
	color.RGBA{194, 196, 157, 255},
	color.RGBA{196, 199, 162, 255},
	color.RGBA{199, 202, 166, 255},
	color.RGBA{202, 205, 170, 255},
	color.RGBA{204, 208, 175, 255},
	color.RGBA{210, 212, 180, 255},
	color.RGBA{207, 210, 180, 255},
	color.RGBA{201, 204, 180, 255},
	color.RGBA{198, 201, 180, 255},
	color.RGBA{195, 199, 180, 255},
	color.RGBA{192, 196, 180, 255},
	color.RGBA{189, 193, 180, 255},
	color.RGBA{185, 190, 180, 255},
	color.RGBA{182, 187, 180, 255},
	color.RGBA{179, 185, 180, 255},
	color.RGBA{176, 182, 180, 255},
	color.RGBA{173, 179, 180, 255},
	color.RGBA{170, 176, 180, 255},
	color.RGBA{164, 171, 180, 255},
	color.RGBA{160, 168, 180, 255},
	color.RGBA{157, 165, 180, 255},
	color.RGBA{154, 162, 180, 255},
	color.RGBA{151, 160, 180, 255},
	color.RGBA{148, 157, 180, 255},
	color.RGBA{145, 154, 180, 255},
	color.RGBA{148, 155, 181, 255},
	color.RGBA{144, 152, 180, 255},
	color.RGBA{140, 149, 179, 255},
	color.RGBA{136, 146, 178, 255},
	color.RGBA{128, 140, 176, 255},
	color.RGBA{124, 137, 175, 255},
	color.RGBA{120, 134, 174, 255},
	color.RGBA{116, 131, 172, 255},
	color.RGBA{112, 128, 171, 255},
	color.RGBA{108, 125, 170, 255},
	color.RGBA{103, 121, 169, 255},
	color.RGBA{99, 118, 168, 255},
	color.RGBA{95, 115, 167, 255},
	color.RGBA{91, 112, 166, 255},
	color.RGBA{87, 109, 164, 255},
	color.RGBA{79, 103, 162, 255},
	color.RGBA{75, 100, 161, 255},
	color.RGBA{71, 97, 160, 255},
	color.RGBA{67, 94, 159, 255},
	color.RGBA{65, 91, 158, 255},
	color.RGBA{67, 97, 162, 255},
	color.RGBA{69, 104, 166, 255},
	color.RGBA{72, 111, 170, 255},
	color.RGBA{74, 118, 174, 255},
	color.RGBA{77, 125, 178, 255},
	color.RGBA{79, 132, 182, 255},
	color.RGBA{81, 139, 187, 255},
	color.RGBA{86, 153, 195, 255},
	color.RGBA{89, 159, 199, 255},
	color.RGBA{91, 166, 203, 255},
	color.RGBA{94, 173, 207, 255},
	color.RGBA{96, 180, 212, 255},
	color.RGBA{98, 187, 216, 255},
	color.RGBA{101, 194, 220, 255},
	color.RGBA{103, 201, 224, 255},
	color.RGBA{106, 208, 228, 255},
	color.RGBA{111, 214, 232, 255},
	color.RGBA{104, 214, 215, 255},
	color.RGBA{89, 214, 179, 255},
	color.RGBA{82, 214, 162, 255},
	color.RGBA{75, 214, 144, 255},
	color.RGBA{67, 214, 126, 255},
	color.RGBA{60, 214, 109, 255},
	color.RGBA{53, 214, 91, 255},
	color.RGBA{17, 213, 24, 255},
	color.RGBA{17, 209, 23, 255},
	color.RGBA{16, 205, 23, 255},
	color.RGBA{16, 200, 22, 255},
	color.RGBA{16, 196, 22, 255},
	color.RGBA{15, 188, 21, 255},
	color.RGBA{15, 183, 20, 255},
	color.RGBA{14, 179, 20, 255},
	color.RGBA{14, 175, 19, 255},
	color.RGBA{14, 171, 19, 255},
	color.RGBA{13, 166, 18, 255},
	color.RGBA{13, 162, 18, 255},
	color.RGBA{13, 158, 17, 255},
	color.RGBA{12, 153, 17, 255},
	color.RGBA{12, 149, 16, 255},
	color.RGBA{12, 145, 16, 255},
	color.RGBA{11, 136, 15, 255},
	color.RGBA{11, 132, 14, 255},
	color.RGBA{10, 128, 14, 255},
	color.RGBA{10, 124, 13, 255},
	color.RGBA{10, 119, 13, 255},
	color.RGBA{9, 115, 12, 255},
	color.RGBA{9, 111, 12, 255},
	color.RGBA{9, 107, 11, 255},
	color.RGBA{8, 102, 11, 255},
	color.RGBA{8, 98, 10, 255},
	color.RGBA{9, 94, 9, 255},
	color.RGBA{50, 115, 8, 255},
	color.RGBA{70, 125, 8, 255},
	color.RGBA{91, 136, 7, 255},
	color.RGBA{111, 146, 7, 255},
	color.RGBA{132, 157, 6, 255},
	color.RGBA{152, 168, 6, 255},
	color.RGBA{173, 178, 5, 255},
	color.RGBA{193, 189, 5, 255},
	color.RGBA{214, 199, 4, 255},
	color.RGBA{234, 210, 4, 255},
	color.RGBA{255, 226, 0, 255},
	color.RGBA{255, 216, 0, 255},
	color.RGBA{255, 211, 0, 255},
	color.RGBA{255, 206, 0, 255},
	color.RGBA{255, 201, 0, 255},
	color.RGBA{255, 196, 0, 255},
	color.RGBA{255, 192, 0, 255},
	color.RGBA{255, 187, 0, 255},
	color.RGBA{255, 182, 0, 255},
	color.RGBA{255, 177, 0, 255},
	color.RGBA{255, 172, 0, 255},
	color.RGBA{255, 167, 0, 255},
	color.RGBA{255, 162, 0, 255},
	color.RGBA{255, 153, 0, 255},
	color.RGBA{255, 148, 0, 255},
	color.RGBA{255, 143, 0, 255},
	color.RGBA{255, 138, 0, 255},
	color.RGBA{255, 133, 0, 255},
	color.RGBA{255, 128, 0, 255},
	color.RGBA{255, 0, 0, 255},
	color.RGBA{248, 0, 0, 255},
	color.RGBA{241, 0, 0, 255},
	color.RGBA{234, 0, 0, 255},
	color.RGBA{227, 0, 0, 255},
	color.RGBA{213, 0, 0, 255},
	color.RGBA{205, 0, 0, 255},
	color.RGBA{198, 0, 0, 255},
	color.RGBA{191, 0, 0, 255},
	color.RGBA{184, 0, 0, 255},
	color.RGBA{177, 0, 0, 255},
	color.RGBA{170, 0, 0, 255},
	color.RGBA{163, 0, 0, 255},
	color.RGBA{155, 0, 0, 255},
	color.RGBA{148, 0, 0, 255},
	color.RGBA{141, 0, 0, 255},
	color.RGBA{127, 0, 0, 255},
	color.RGBA{120, 0, 0, 255},
	color.RGBA{113, 0, 0, 255},
	color.RGBA{255, 255, 255, 255},
	color.RGBA{255, 245, 255, 255},
	color.RGBA{255, 234, 255, 255},
	color.RGBA{255, 223, 255, 255},
	color.RGBA{255, 212, 255, 255},
	color.RGBA{255, 201, 255, 255},
	color.RGBA{255, 190, 255, 255},
	color.RGBA{255, 179, 255, 255},
	color.RGBA{255, 157, 255, 255},
	color.RGBA{255, 146, 255, 255},
	color.RGBA{255, 117, 255, 255},
	color.RGBA{252, 107, 253, 255},
	color.RGBA{249, 96, 250, 255},
	color.RGBA{246, 86, 247, 255},
	color.RGBA{243, 75, 244, 255},
	color.RGBA{240, 64, 241, 255},
	color.RGBA{237, 54, 239, 255},
	color.RGBA{234, 43, 236, 255},
	color.RGBA{231, 32, 233, 255},
	color.RGBA{225, 11, 227, 255},
	color.RGBA{178, 0, 255, 255},
	color.RGBA{172, 0, 252, 255},
	color.RGBA{164, 0, 247, 255},
	color.RGBA{155, 0, 244, 255},
	color.RGBA{147, 0, 239, 255},
	color.RGBA{136, 0, 234, 255},
	color.RGBA{131, 0, 232, 255},
	color.RGBA{121, 0, 226, 255},
	color.RGBA{114, 0, 221, 255},
	color.RGBA{105, 0, 219, 255},
	color.RGBA{5, 236, 240, 255},
	color.RGBA{5, 235, 240, 255},
	color.RGBA{5, 234, 240, 255},
	color.RGBA{5, 221, 224, 255},
	color.RGBA{5, 220, 224, 255},
	color.RGBA{5, 219, 224, 255},
	color.RGBA{5, 205, 208, 255},
	color.RGBA{5, 204, 208, 255},
	color.RGBA{4, 189, 192, 255},
	color.RGBA{4, 188, 192, 255},
	color.RGBA{4, 187, 192, 255},
	color.RGBA{4, 174, 176, 255},
	color.RGBA{4, 173, 176, 255},
	color.RGBA{4, 158, 160, 255},
	color.RGBA{4, 157, 160, 255},
	color.RGBA{4, 156, 160, 255},
	color.RGBA{3, 142, 144, 255},
	color.RGBA{3, 141, 144, 255},
	color.RGBA{3, 140, 144, 255},
	color.RGBA{3, 126, 128, 255},
	color.RGBA{3, 125, 128, 255},
	color.RGBA{3, 111, 112, 255},
	color.RGBA{3, 110, 112, 255},
	color.RGBA{3, 109, 112, 255},
	color.RGBA{2, 95, 96, 255},
	color.RGBA{2, 94, 96, 255},
	color.RGBA{2, 79, 80, 255},
	color.RGBA{2, 78, 80, 255},
	color.RGBA{2, 77, 80, 255},
	color.RGBA{2, 63, 64, 255},
	color.RGBA{2, 62, 64, 255},
	color.RGBA{2, 61, 64, 255},
	color.RGBA{1, 48, 48, 255},
	color.RGBA{1, 47, 48, 255},
	color.RGBA{1, 32, 32, 255},
	color.RGBA{1, 31, 32, 255},
	color.RGBA{1, 30, 32, 255},
	color.RGBA{58, 103, 181, 255},
	color.RGBA{58, 102, 181, 255},
	color.RGBA{58, 101, 181, 255},
	color.RGBA{58, 100, 181, 255},
	color.RGBA{58, 99, 181, 255},
}

func findRasterVal(c color.Color) (float64, error) {
	idx := n0qPalette.Index(c)

	// Check if we have information on this index.
	if v, ok := n0qRaster[idx]; !ok {
		return 0, errors.New("No data.")
	} else {
		return v, nil
	}
	return 0, errors.New("Error.")
}

func extractMatrix(bounds image.Rectangle, img image.Image, sz int) [][][][]color.Color {
	ret := make([][][][]color.Color, bounds.Max.X/sz)
	for i := 0; i < bounds.Max.X; i += sz {
		ret[i/sz] = make([][][]color.Color, bounds.Max.Y/sz)
		for j := 0; j < bounds.Max.Y; j += sz {
			ret[i/sz][j/sz] = make([][]color.Color, sz)
			for k := 0; k < sz; k++ {
				ret[i/sz][j/sz][k] = make([]color.Color, sz)
				for l := 0; l < sz; l++ {
					ret[i/sz][j/sz][k][l] = img.At(i+k, j+l)
				}
			}
		}
	}
	return ret
}

func getMaxIntensityFromMatrix(grp [][]color.Color) float64 {
	var dBzs []float64
	for i := 0; i < len(grp); i++ {
		for j := 0; j < len(grp[i]); j++ {
			dBz, _ := findRasterVal(grp[i][j])
			dBzs = append(dBzs, dBz)
		}
	}
	sort.Float64s(dBzs)
	ret := dBzs[len(dBzs)-1] // Highest will be last in the sorted list.
	return ret
}

func fillImage(newImg *image.NRGBA64, newColor color.Color, x, y, sz int) {
	for i := 0; i < sz; i++ {
		for j := 0; j < sz; j++ {
			newImg.Set(x*sz+i, y*sz+j, newColor)
		}
	}
}

func main() {
	inf, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Printf("can't open '%s'\n", os.Args[1])
		return
	}
	defer inf.Close()

	src, _, err := image.Decode(inf)
	if err != nil {
		fmt.Printf("image decode err: %s\n", err.Error())
		return
	}

	bounds := src.Bounds()
	// Check if the image is 256x256.
	if ((bounds.Max.X - bounds.Min.X) != 256) || ((bounds.Max.Y - bounds.Min.Y) != 256) {
		fmt.Printf("not a 256x256 tile.\n")
		return
	}

	// Create new image with reduced color depth.
	newImg := image.NewNRGBA64(bounds)

	grp := extractMatrix(bounds, src, REDUCE_MATRIX_SZ)

	for i := 0; i < len(grp); i++ {
		for j := 0; j < len(grp[i]); j++ {
			dBz := getMaxIntensityFromMatrix(grp[i][j])

			var newColor color.RGBA
			newColor.A = 255
			if dBz >= 40.0 {
				newColor.R = 255
			} else if dBz < 40.0 && dBz >= 20.0 {
				newColor.B = 255
			} else if dBz < 20.0 && dBz > -999 {
				newColor.G = 255
			} else if dBz <= -999 {
				newColor.A = 0 // No data.
			}

			fillImage(newImg, newColor, i, j, REDUCE_MATRIX_SZ)

		}
	}

	// Write new image.
	outf, err := os.Create(os.Args[2])
	if err != nil {
		fmt.Printf("can't write new file: %s\n", err.Error())
		return
	}
	defer outf.Close()

	png.Encode(outf, newImg)

}
