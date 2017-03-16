package WeatherPack

import (
	"errors"
	"math"
	"time"
)

type WeatherPacker interface {
	Pack() []int
}

type WeatherUnpacker interface {
	Unpack([]int) error
}

type WeatherCollectionTime time.Time
type TempDewpointSpread int // ºC.
type WindDirection int      // Degrees.
type WindVelocity int       // Knots.
type WindVelocityGusts int  // Knots (max).
type Visibility int         // Statute miles.
type Altimeter float64      // Inches hG.
type WeatherPhenomena string
type CeilingHeight string
type CeilingType string

// 8 bits, hh:mm. Minutes rounded to tens.
func (t WeatherCollectionTime) Pack() []int {
	ret := make([]int, 8)
	// Hours.
	hr := time.Time(t).Hour()

	// Minutes.
	min := time.Time(t).Minute()
	roundedMinutes := int(math.Floor((float64(min) / 10.0) + 0.5))

	// Rounded minute is 6? Round to next hour.
	if roundedMinutes == 6 {
		roundedMinutes = 0
		hr++
		if hr == 24 {
			roundedHours = 0
		}
	}

	// Encode hours.
	ret[0], ret[1], ret[2], ret[3], ret[4] = (hr>>4)&1, (hr>>3)&1, (hr>>2)&1, (hr>>1)&1, hr&1
	// Encode tens of minutes.
	ret[5], ret[6], ret[7] = (roundedMinutes>>2)&1, (roundedMinutes>>1)&1, roundedMinutes&1
	return ret
}

func (t *WeatherCollectionTime) Unpack(data []int) error {
	if len(data) != 8 {
		return errors.New("Invalid length.")
	}

	// Hours.
	hrs := data[0]<<4 | data[1]<<3 | data[2]<<2 | data[3]<<1 | data[4]
	// Tens of minutes.
	tensMins := data[5]<<2 | data[6]<<1 | data[7]

	// Set time.
	*t = WeatherCollectionTime(time.Date(0, 0, 0, hrs, tensMins*10, 0, 0, time.UTC))

	return nil
}

type IntDecode struct {
	Min       int
	Max       int
	BitVal    []int
	StringVal int
}

var tempDewpointDecode = []IntDecode{
	{Min: 0, Max: 1, BitVal: []int{0, 0, 0}, StringVal: "0ºC"},
	{Min: 1, Max: 2, BitVal: []int{0, 0, 1}, StringVal: "1ºC"},
	{Min: 2, Max: 3, BitVal: []int{0, 1, 0}, StringVal: "2ºC"},
	{Min: 3, Max: 4, BitVal: []int{0, 1, 1}, StringVal: "3ºC"},
	{Min: 4, Max: 5, BitVal: []int{1, 0, 0}, StringVal: "4ºC"},
	{Min: 5, Max: 6, BitVal: []int{1, 0, 1}, StringVal: "5ºC"},
	{Min: 6, Max: 7, BitVal: []int{1, 1, 0}, StringVal: "6ºC"},
	{Min: 7, Max: 8, BitVal: []int{1, 1, 1}, StringVal: ">=7ºC"},
}

var windDirectionDecode = []IntDecode{
	{Min: 0, Max: 45, BitVal: []int{0, 0, 0}, StringVal: "0-45"},
	{Min: 45, Max: 90, BitVal: []int{0, 0, 1}, StringVal: "45-90"},
	{Min: 90, Max: 135, BitVal: []int{0, 1, 0}, StringVal: "90-135"},
	{Min: 135, Max: 180, BitVal: []int{0, 1, 1}, StringVal: "135-180"},
	{Min: 180, Max: 225, BitVal: []int{1, 0, 0}, StringVal: "180-225"},
	{Min: 225, Max: 270, BitVal: []int{1, 0, 1}, StringVal: "225-270"},
	{Min: 270, Max: 315, BitVal: []int{1, 1, 0}, StringVal: "270-315"},
	{Min: 315, Max: 360, BitVal: []int{1, 1, 1}, StringVal: "315-360"},
}

var windVelocityDecode = []IntDecode{
	{Min: 0, Max: 10, BitVal: []int{0, 0}, StringVal: "=<10kts"},
	{Min: 10, Max: 15, BitVal: []int{0, 1}, StringVal: "10-15kts"},
	{Min: 15, Max: 20, BitVal: []int{1, 0}, StringVal: "15-20kts"},
	{Min: 20, Max: 99999, BitVal: []int{1, 1}, StringVal: ">20kts"},
}

var windVelocityGustsDecode = []IntDecode{
	{Min: 0, Max: 0, BitVal: []int{0, 0}, StringVal: "0kts"},
	{Min: 0, Max: 5, BitVal: []int{0, 1}, StringVal: "0-5kts"},
	{Min: 5, Max: 15, BitVal: []int{1, 0}, StringVal: "5-15kts"},
	{Min: 15, Max: 99999, BitVal: []int{1, 1}, StringVal: ">15kts"},
}

var visibilityDecode = []IntDecode{
	{Min: 0, Max: 1, BitVal: []int{0, 0}, StringVal: "0-1sm"},
	{Min: 1, Max: 3, BitVal: []int{0, 1}, StringVal: "1-3sm"},
	{Min: 3, Max: 5, BitVal: []int{1, 0}, StringVal: "3-5sm"},
	{Min: 5, Max: 99999, BitVal: []int{1, 1}, StringVal: ">5sm"},
}

var altimeterDecode = []IntDecode{
	{Min: -1.00, Max: 26.0, BitVal: []int{0, 0, 0, 0, 0, 0}, StringVal: "<26.00"},
	{Min: 26.00, Max: 26.10, BitVal: []int{0, 0, 0, 0, 0, 1}, StringVal: "26.00-26.10"},
	{Min: 26.10, Max: 26.20, BitVal: []int{0, 0, 0, 0, 1, 0}, StringVal: "26.10-26.20"},
	{Min: 26.20, Max: 26.30, BitVal: []int{0, 0, 0, 0, 1, 1}, StringVal: "26.20-26.30"},
	{Min: 26.30, Max: 26.40, BitVal: []int{0, 0, 0, 1, 0, 0}, StringVal: "26.30-26.40"},
	{Min: 26.40, Max: 26.50, BitVal: []int{0, 0, 0, 1, 0, 1}, StringVal: "26.40-26.50"},
	{Min: 26.50, Max: 26.60, BitVal: []int{0, 0, 0, 1, 1, 0}, StringVal: "26.50-26.60"},
	{Min: 26.60, Max: 26.70, BitVal: []int{0, 0, 0, 1, 1, 1}, StringVal: "26.60-26.70"},
	{Min: 26.70, Max: 26.80, BitVal: []int{0, 0, 1, 0, 0, 0}, StringVal: "26.70-26.80"},
	{Min: 26.80, Max: 26.90, BitVal: []int{0, 0, 1, 0, 0, 1}, StringVal: "26.80-26.90"},
	{Min: 26.90, Max: 27.00, BitVal: []int{0, 0, 1, 0, 1, 0}, StringVal: "26.90-27.00"},
	{Min: 27.00, Max: 27.10, BitVal: []int{0, 0, 1, 0, 1, 1}, StringVal: "27.00-27.10"},
	{Min: 27.10, Max: 27.20, BitVal: []int{0, 0, 1, 1, 0, 0}, StringVal: "27.10-27.20"},
	{Min: 27.20, Max: 27.30, BitVal: []int{0, 0, 1, 1, 0, 1}, StringVal: "27.20-27.30"},
	{Min: 27.30, Max: 27.40, BitVal: []int{0, 0, 1, 1, 1, 0}, StringVal: "27.30-27.40"},
	{Min: 27.40, Max: 27.50, BitVal: []int{0, 0, 1, 1, 1, 1}, StringVal: "27.40-27.50"},
	{Min: 27.50, Max: 27.60, BitVal: []int{0, 1, 0, 0, 0, 0}, StringVal: "27.50-27.60"},
	{Min: 27.60, Max: 27.70, BitVal: []int{0, 1, 0, 0, 0, 1}, StringVal: "27.60-27.70"},
	{Min: 27.70, Max: 27.80, BitVal: []int{0, 1, 0, 0, 1, 0}, StringVal: "27.70-27.80"},
	{Min: 27.80, Max: 27.90, BitVal: []int{0, 1, 0, 0, 1, 1}, StringVal: "27.80-27.90"},
	{Min: 27.90, Max: 28.00, BitVal: []int{0, 1, 0, 1, 0, 0}, StringVal: "27.90-28.00"},
	{Min: 28.00, Max: 28.10, BitVal: []int{0, 1, 0, 1, 0, 1}, StringVal: "28.00-28.10"},
	{Min: 28.10, Max: 28.20, BitVal: []int{0, 1, 0, 1, 1, 0}, StringVal: "28.10-28.20"},
	{Min: 28.20, Max: 28.30, BitVal: []int{0, 1, 0, 1, 1, 1}, StringVal: "28.20-28.30"},
	{Min: 28.30, Max: 28.40, BitVal: []int{0, 1, 1, 0, 0, 0}, StringVal: "28.30-28.40"},
	{Min: 28.40, Max: 28.50, BitVal: []int{0, 1, 1, 0, 0, 1}, StringVal: "28.40-28.50"},
	{Min: 28.50, Max: 28.60, BitVal: []int{0, 1, 1, 0, 1, 0}, StringVal: "28.50-28.60"},
	{Min: 28.60, Max: 28.70, BitVal: []int{0, 1, 1, 0, 1, 1}, StringVal: "28.60-28.70"},
	{Min: 28.70, Max: 28.80, BitVal: []int{0, 1, 1, 1, 0, 0}, StringVal: "28.70-28.80"},
	{Min: 28.80, Max: 28.90, BitVal: []int{0, 1, 1, 1, 0, 1}, StringVal: "28.80-28.90"},
	{Min: 28.90, Max: 29.00, BitVal: []int{0, 1, 1, 1, 1, 0}, StringVal: "28.90-29.00"},
	{Min: 29.00, Max: 29.10, BitVal: []int{0, 1, 1, 1, 1, 1}, StringVal: "29.00-29.10"},
	{Min: 29.10, Max: 29.20, BitVal: []int{1, 0, 0, 0, 0, 0}, StringVal: "29.10-29.20"},
	{Min: 29.20, Max: 29.30, BitVal: []int{1, 0, 0, 0, 0, 1}, StringVal: "29.20-29.30"},
	{Min: 29.30, Max: 29.40, BitVal: []int{1, 0, 0, 0, 1, 0}, StringVal: "29.30-29.40"},
	{Min: 29.40, Max: 29.50, BitVal: []int{1, 0, 0, 0, 1, 1}, StringVal: "29.40-29.50"},
	{Min: 29.50, Max: 29.60, BitVal: []int{1, 0, 0, 1, 0, 0}, StringVal: "29.50-29.60"},
	{Min: 29.60, Max: 29.70, BitVal: []int{1, 0, 0, 1, 0, 1}, StringVal: "29.60-29.70"},
	{Min: 29.70, Max: 29.80, BitVal: []int{1, 0, 0, 1, 1, 0}, StringVal: "29.70-29.80"},
	{Min: 29.80, Max: 29.90, BitVal: []int{1, 0, 0, 1, 1, 1}, StringVal: "29.80-29.90"},
	{Min: 29.90, Max: 30.00, BitVal: []int{1, 0, 1, 0, 0, 0}, StringVal: "29.90-30.00"},
	{Min: 30.00, Max: 30.10, BitVal: []int{1, 0, 1, 0, 0, 1}, StringVal: "30.00-30.10"},
	{Min: 30.10, Max: 30.20, BitVal: []int{1, 0, 1, 0, 1, 0}, StringVal: "30.10-30.20"},
	{Min: 30.20, Max: 30.30, BitVal: []int{1, 0, 1, 0, 1, 1}, StringVal: "30.20-30.30"},
	{Min: 30.30, Max: 30.40, BitVal: []int{1, 0, 1, 1, 0, 0}, StringVal: "30.30-30.40"},
	{Min: 30.40, Max: 30.50, BitVal: []int{1, 0, 1, 1, 0, 1}, StringVal: "30.40-30.50"},
	{Min: 30.50, Max: 30.60, BitVal: []int{1, 0, 1, 1, 1, 0}, StringVal: "30.50-30.60"},
	{Min: 30.60, Max: 30.70, BitVal: []int{1, 0, 1, 1, 1, 1}, StringVal: "30.60-30.70"},
	{Min: 30.70, Max: 30.80, BitVal: []int{1, 1, 0, 0, 0, 0}, StringVal: "30.70-30.80"},
	{Min: 30.80, Max: 30.90, BitVal: []int{1, 1, 0, 0, 0, 1}, StringVal: "30.80-30.90"},
	{Min: 30.90, Max: 31.00, BitVal: []int{1, 1, 0, 0, 1, 0}, StringVal: "30.90-31.00"},
	{Min: 31.00, Max: 31.10, BitVal: []int{1, 1, 0, 0, 1, 1}, StringVal: "31.00-31.10"},
	{Min: 31.10, Max: 31.20, BitVal: []int{1, 1, 0, 1, 0, 0}, StringVal: "31.10-31.20"},
	{Min: 31.20, Max: 31.30, BitVal: []int{1, 1, 0, 1, 0, 1}, StringVal: "31.20-31.30"},
	{Min: 31.30, Max: 31.40, BitVal: []int{1, 1, 0, 1, 1, 0}, StringVal: "31.30-31.40"},
	{Min: 31.40, Max: 31.50, BitVal: []int{1, 1, 0, 1, 1, 1}, StringVal: "31.40-31.50"},
	{Min: 31.50, Max: 31.60, BitVal: []int{1, 1, 1, 0, 0, 0}, StringVal: "31.50-31.60"},
	{Min: 31.60, Max: 31.70, BitVal: []int{1, 1, 1, 0, 0, 1}, StringVal: "31.60-31.70"},
	{Min: 31.70, Max: 31.80, BitVal: []int{1, 1, 1, 0, 1, 0}, StringVal: "31.70-31.80"},
	{Min: 31.80, Max: 31.90, BitVal: []int{1, 1, 1, 0, 1, 1}, StringVal: "31.80-31.90"},
	{Min: 31.90, Max: 32.00, BitVal: []int{1, 1, 1, 1, 0, 0}, StringVal: "31.90-32.00"},
	{Min: 32.00, Max: 32.10, BitVal: []int{1, 1, 1, 1, 0, 1}, StringVal: "32.00-32.10"},
	{Min: 32.10, Max: 32.20, BitVal: []int{1, 1, 1, 1, 1, 0}, StringVal: "32.10-32.20"},
	{Min: 32.20, Max: 32.30, BitVal: []int{1, 1, 1, 1, 1, 1}, StringVal: "32.20-32.30"},
}

var ceilingHeightDecode = []IntDecode{
	{Min: -1, Max: 200, BitVal: []int{0, 0, 0}, StringVal: "<200ft"},
	{Min: 200, Max: 300, BitVal: []int{0, 0, 1}, StringVal: "200-300ft"},
	{Min: 300, Max: 400, BitVal: []int{0, 0, 1}, StringVal: "300-400ft"},
	{Min: 400, Max: 600, BitVal: []int{0, 0, 1}, StringVal: "400-600ft"},
	{Min: 600, Max: 800, BitVal: []int{0, 0, 1}, StringVal: "600-800ft"},
	{Min: 800, Max: 1000, BitVal: []int{0, 0, 1}, StringVal: "800-1000ft"},
	{Min: 1000, Max: 1500, BitVal: []int{0, 0, 1}, StringVal: "1000-1500ft"},
	{Min: 1500, Max: 99999, BitVal: []int{0, 0, 1}, StringVal: ">1500ft"},
}

type StringDecode struct {
	Keywords  []string
	BitVal    []int
	StringVal string
}

var ceilingTypeDecode = []StringDecode{
	{Keywords: []string{"CLR", "FEW"}, BitVal: []int{0, 0}, StringVal: "CLR-FEW"},
	{Keywords: []string{"SCT"}, BitVal: []int{0, 1}, StringVal: "SCT"},
	{Keywords: []string{"BKN"}, BitVal: []int{1, 0}, StringVal: "BKN"},
	{Keywords: []string{"OVC"}, BitVal: []int{1, 1}, StringVal: "OVC"},
}

func Unmarshal(data []byte, v interface{}) error {
	//TODO.
	return nil
}

func Marshal(v interface{}) ([]byte, error) {
	//TODO.
	return []byte{}, nil
}
