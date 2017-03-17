package main

import (
	"../ADDS"
	"fmt"
	"github.com/kellydunn/golang-geo"
)

func main() {
	rectBottomLeft := geo.NewPoint(41.913392, -84.501343)
	rectTopRight := geo.NewPoint(42.846646, -82.611694)
	a, b := ADDS.GetLatestADDSMETARsInRect(rectBottomLeft, rectTopRight)
	fmt.Printf("%v %v\n", a, b)

	a, b = ADDS.GetLatestADDSMETARsAlongRoute(25.0, "KDTW;KRMY")
	fmt.Printf("%v %v\n", a, b)

	fmt.Printf("***PIREPS***\n")
	t := geo.NewPoint(40.369305, -123.200684)
	c, d := ADDS.GetLatestADDSPIREPsInRadiusOf(500, t)
	fmt.Printf("%v %v\n", c, d)

	fmt.Printf("***TAFS***\n")
	e, f := ADDS.GetADDSTAFsByIdent("KDTW")
	fmt.Printf("%v %v\n", e, f)

}
