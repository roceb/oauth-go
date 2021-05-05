package auth

import (
	"fmt"

	"golang.org/x/oauth2"
)

type Config map[string]*App

type App struct {
	Name      string
	AuthState string
	AuthCode  string
	oauth2.Config
	oauth2.Token
}

func main() {
	fmt.Println("vim-go")
}
