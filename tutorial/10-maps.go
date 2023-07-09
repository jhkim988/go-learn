package main

import "fmt"

func main() {
	StudentAge := make(map[string]int)
	StudentAge["jh"] = 100
	StudentAge["aaa"] = 12
	StudentAge["bbb"] = 23
	StudentAge["ccc"] = 34
	StudentAge["ddd"] = 45

	fmt.Println(StudentAge)
	fmt.Println(StudentAge["jh"])
	fmt.Println(StudentAge["NO"])
	fmt.Println(len(StudentAge))

	superhero := map[string]map[string]string{
		"Superman": map[string]string{
			"RealName": "Clark Kent",
			"City":     "Metropolis",
		},
		"Batman": map[string]string{
			"RealName": "Bruce Wayne",
			"City":     "Gotham City",
		},
	}

	// temp 에 key "Superman" 에 대한 value(map), hero 에 존재하는지 여부(bool) 이 할당된다.
	if temp, hero := superhero["Superman"]; hero {
		fmt.Println(temp["RealName"], temp["City"])
		fmt.Println(temp, hero)
	}
}
