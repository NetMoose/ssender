package main

import (
	"fmt"
	"strings"

	"github.com/jessevdk/go-flags"
)

type Options struct {
	Telegram      bool   `short:"t" long:"telegram-send" description:"Send to telegram"`
	Telegramtoken string `short:"k" long:"telegram-token" description:"Token for send to telegram"`
	VK            bool   `short:"v" long:"vk-send" description:"Send to vk"`
	VKtoken       string `short:"b" long:"vk-token" description:"Token for send to VK"`
	Facebook      bool   `short:"f" long:"fb-send" description:"Send to Facebook"`
	FBtoken       string `short:"m" long:"fb-token" description:"Token for send to Facebook"`
}

func main() {
	var options Options
	var parser = flags.NewParser(&options, flags.Default)
	// parser.CommandHandler = func(command flags.Commander, args []string) error {
	// 	print(options.Telegram)
	// }
	args, err := parser.Parse()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Telegram send: %v\n", options.Telegram)
	fmt.Printf("Telegram token: %s\n", options.Telegramtoken)
	fmt.Printf("VK send: %v\n", options.VK)
	fmt.Printf("VK token: %s\n", options.VKtoken)
	fmt.Printf("Facebook send: %v\n", options.Facebook)
	fmt.Printf("Facebook token: %s\n", options.FBtoken)
	fmt.Printf("Remaining args: %s\n", strings.Join(args, " "))
}
