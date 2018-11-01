package main

import (
	"time"

	utils "github.com/governor-clerk-utils"
)

func main() {
	for {
		tickets := utils.GetTickets("active", "disable")
		for _, ticket := range *tickets {
			utils.DisableUser(&ticket)
		}
		time.Sleep(5 * time.Second)
	}
}
