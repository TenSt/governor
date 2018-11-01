package main

import (
	"time"

	utils "github.com/verfio/governor-clerk-utils"
)

func main() {
	for {
		tickets := utils.GetTickets("active", "create")
		for _, ticket := range *tickets {
			utils.AddUser(&ticket)
		}
		time.Sleep(5 * time.Second)
	}
}
