package vault

import (
	"fmt"
	"log"
	"os"

	"camera_api/app"

	"github.com/joho/godotenv"
)

func ReadEnv(env string) (*Setting, *app.Setting, error) {
	// Load .env for local development.
	if err := godotenv.Load(".env"); err == nil {
		log.Println("run with loaded .env")
	}

	hc := HashiCorp{
		Product: os.Getenv("PRODUCT"),
		ENV:     env,
		File:    os.Getenv("FILE"),
		Host:    os.Getenv("HOST"),
		Token:   os.Getenv("TOKEN"),
	}

	log.Printf("product : %s", hc.Product)
	log.Printf("env     : %s", hc.ENV)

	conf, appSetting, err := ReadVault(hc)
	if err != nil {
		return nil, nil, fmt.Errorf("ReadVault:%w", err)
	}

	log.Printf("logger  : %s", conf.General.Log)

	return conf, appSetting, nil
}
