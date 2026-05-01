package main

import "example.com/tomo/internal/config"

func main() {
	viperConfig := config.NewViper
	log := config.NewLogger(viperConfig())
	db := config.ConnectDB(viperConfig(), log)
	validate := config.NewValidator(viperConfig())
	redis := config.SetUpRedis(viperConfig(), log)
	app := config.NewEcho(viperConfig())
	jwt := config.SetUpJWT(viperConfig(), log)
	config.BootStrap(&config.BootStrapConfig{
		DB:       db,
		App:      app,
		Log:      log,
		Validate: validate,
		Config:   viperConfig(),
		Redis:    redis,
		JWT:      jwt,
	})
	app.Start(":9000")
}
