package main

import "time"

func main() {

	//run(exec.Command("PowerShell", "-Command", "Set-ADAccountPassword", "-Identity 'CN=test,CN=Users,DC=governor,DC=local'", "-Reset", "-NewPassword (ConvertTo-SecureString -AsPlainText "+password+" -Force)"))

	// for i := 0; i < 2; i++ {

	// 	var x task
	// 	x.Action = "reset"
	// 	x.Action = "active"
	// 	x.Number = int64(i)
	// 	//x.Timestamp = time.Now().Format(time.RFC850)
	// 	x.User = "user" + strconv.Itoa(i)

	// 	putdataMongo(x)

	// }

	for {

		tickets := getdataMongo("active", "reset")

		for _, Ticket := range tickets {
			resetPassword(Ticket)
		}

		time.Sleep(5 * time.Second)
	}

}
