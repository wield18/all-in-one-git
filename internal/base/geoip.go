package base

import "github.com/oschwald/geoip2-golang"

func NewGeoipLoc(path string) *geoip2.Reader {
	db, err := geoip2.Open(path)
	if err != nil {
		panic(err)
	}
	// 	// defer db.Close()

	// 	// ip := net.ParseIP("8.8.8.8")
	// 	// record, _ := db.City(ip)

	// 	// fmt.Println(record.Country.IsoCode)   // "US"
	// 	// fmt.Println(record.City.Names["en"])  // 城市名
	// 	// fmt.Println(record.Location.Latitude) // 纬度

	return db
}
