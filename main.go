package main

import (
	"encoding/xml"
	"html/template"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Telegram struct {
		Send  bool   `yaml:"send"`
		Token string `yaml:"token"`
	} `yaml:"telegram"`
	VK struct {
		Send  bool   `yaml:"send"`
		Token string `yaml:"token"`
	} `yaml:"vk"`
	Facebook struct {
		Send  bool   `yaml:"send"`
		Token string `yaml:"token"`
	} `yaml:"facebook"`
}

func NewConfig(configPath string) (*Config, error) {
	// Create config structure
	config := &Config{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

type Options struct {
	FileParse  string `short:"f" long:"fileparse" description:"File for parce (rss xml)"`
	ConfigPath string `short:"c" long:"configpath" description:"Config file path"`
}

var ConfigPath = "/etc/ssender/config.yml"

type Rss2 struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	// Required
	Title       string `xml:"channel>title"`
	Link        string `xml:"channel>link"`
	Description string `xml:"channel>description"`
	// Optional
	PubDate  string `xml:"channel>pubDate"`
	ItemList []Item `xml:"channel>item"`
}

type Item struct {
	// Required
	Title       string        `xml:"title"`
	Link        string        `xml:"link"`
	Description template.HTML `xml:"description"`
	// Optional
	Content  template.HTML `xml:"encoded"`
	PubDate  string        `xml:"pubDate"`
	Comments string        `xml:"comments"`
}

func NewRSS(rssPath string) (*Rss2, error) {
	rss := &Rss2{}

	// Open rss2 file
	file, err := os.Open(rssPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := xml.NewDecoder(file)

	// Start RSS decoding from file
	if err := d.Decode(&rss); err != nil {
		return nil, err
	}

	return rss, nil
}

func (config Config) Run() {
	if config.Telegram.Send {
		log.Println("Send to telegram.")
	}
	if config.VK.Send {
		log.Println("Send to VK.")
	}
	if config.Facebook.Send {
		log.Println("Send to Facebook.")
	}
}

func main() {
	// Parse flags
	var options Options
	var parser = flags.NewParser(&options, flags.Default)
	if _, err := parser.Parse(); err != nil {
		switch flagsErr := err.(type) {
		case flags.ErrorType:
			if flagsErr == flags.ErrHelp {
				os.Exit(0)
			}
			os.Exit(1)
		default:
			os.Exit(1)
		}
	}
	log.Println("Flags processed.")

	if options.ConfigPath != "" {
		log.Printf("Config from: %s\n", options.ConfigPath)
		ConfigPath = options.ConfigPath
	}

	// Get config
	cfg, err := NewConfig(ConfigPath)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Config processed.")

	// Parse file
	log.Printf("Parse file %s \n", options.FileParse)
	rss, err := NewRSS(options.FileParse)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Rss: %v\n", rss)

	// Run send data depended on configuration options
	log.Println("Run send process.")
	cfg.Run()
}
