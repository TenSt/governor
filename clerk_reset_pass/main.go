package main

import (
	"time"

	utils "github.com/verfio/governor-clerk-utils"
)

func main() {
	for {
		tickets := utils.GetTickets("active", "reset")
		for _, ticket := range *tickets {
			utils.ResetPassword(&ticket)
		}
		time.Sleep(5 * time.Second)
	}
}
