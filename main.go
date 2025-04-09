package main

import (
	"box/code/core"
	"box/code/support"
	_ "gorm.io/driver/mysql"
)

func main() {

	core.APPLICATION = &support.TankApplication{}
	core.APPLICATION.Start()

}
