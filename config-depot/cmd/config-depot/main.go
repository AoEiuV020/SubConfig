package main

import (
	"log"
	"net/http"

	"github.com/AoEiuV020/SubConfig/config-depot/internal/app"
)

func main() {
	server, err := app.NewServer(app.SettingsFromEnv())
	if err != nil {
		log.Fatalf("初始化服务失败: %v", err)
	}

	log.Printf("config-depot 监听 %s", server.Address())
	if err := http.ListenAndServe(server.Address(), server); err != nil {
		log.Fatalf("服务退出: %v", err)
	}
}
