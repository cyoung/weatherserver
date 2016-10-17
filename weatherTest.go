package main

import (
 "fmt"
 "./WeatherPack"
 "time"
)

func main() {
  p := WeatherPack.WeatherCollectionTime(time.Now())
  fmt.Printf("%v\n", p.Encode())

 var t WeatherPack.WeatherCollectionTime
 t.Decode(p.Encode())

 fmt.Printf("%v\n", t)
 fmt.Printf("%s\n", t)

}
