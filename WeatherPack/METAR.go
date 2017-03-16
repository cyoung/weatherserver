package WeatherPack

type METAR struct {
}

func PackMETAR(time_hour, time_minute, temperature, dewpoint, wind_direction, wind_velocity, visibility int, altimeter float64, weather_phenomena []string, ceiling_height int, ceiling_type string, notification bool) []byte {
	return []byte{}
}
