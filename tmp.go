package main

import (
	"encoding/json"
	"fmt"
)

func GetMysqlMsg() string {
	args := make(map[string]interface{})
	args["database"] = "mimictl_user1"
	args["username"] = "mimictl_user1"
	args["password"] = "123456"
	cmd := &Command{Plugin: "mysql", Args: args}

	msg, err := json.Marshal(cmd)
	if err != nil {
		panic(err)
	}
	return string(msg)
}

func GetPgsqlMsg() string {
	args := make(map[string]interface{})
	args["database"] = "mimictl_user"
	args["username"] = "mimictl_user"
	args["password"] = "123456"
	cmd := &Command{Plugin: "pgsql", Args: args}

	msg, err := json.Marshal(cmd)
	if err != nil {
		panic(err)
	}
	return string(msg)
}

func GetMongoMsg() string {
	args := make(map[string]interface{})
	args["database"] = "mimictl_user"
	args["username"] = "mimictl_user"
	args["password"] = "123456"
	cmd := &Command{Plugin: "mongo", Args: args}

	msg, err := json.Marshal(cmd)
	if err != nil {
		panic(err)
	}
	return string(msg)
}

func GetOpenvzMsg() string {
	args := make(map[string]interface{})

	cpuunits := "600"
	cpulimit := "25"
	memorylimit := 256
	disklimit := 2048

	args["cpuunits"] = cpuunits
	args["cpulimit"] = cpulimit

	args["ram"] = fmt.Sprintf("%sM", memorylimit)
	args["swap"] = fmt.Sprintf("%sM", memorylimit*2)

	args["vmguarpages"] = fmt.Sprintf("%sM", memorylimit)
	args["oomguarpages"] = fmt.Sprintf("%sM", memorylimit)
	args["privvmpages"] = fmt.Sprintf("%sM:%sM", memorylimit, memorylimit*2)

	args["diskspace"] = fmt.Sprintf("%sM:%sM", disklimit, disklimit+200)
	args["diskinodes"] = fmt.Sprintf("%s:%s", 300000*(disklimit/1024), 320000*(disklimit/1024))

	// args["ioprio"] = fmt.Sprintf("%s", 4)

	// максимальное количество процессов контейнере (защитит от форкбомбы)
	args["numproc"] = "1024:1024"
	// максимальное количество TCP-сокетов
	// args["numtcpsock"] = "1024:1024"
	// максимальное количество не TCP-сокетов
	// args["numothersock"] = "1024:1024"

	// ограничения на размер буфера отправки TCP
	// args["tcpsndbuf"] = "1m:2m"
	// ограничения на размер буфера приема TCP
	// args["tcprcvbuf"] = "1m:2m"

	// максимальное количество открытых файлов
	args["numfile"] = "4096"

	cmd := &Command{Plugin: "openvz", Action: "Create", Args: args}

	msg, err := json.Marshal(cmd)
	if err != nil {
		panic(err)
	}
	return string(msg)
}
