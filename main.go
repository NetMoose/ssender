package main

import (
	"encoding/json"
	"encoding/xml"
	"html"
	"html/template"
	"log"
	"os"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/boltdb/bolt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jessevdk/go-flags"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Dbpath   string `yaml:"dbpath"`
	Telegram struct {
		Send      bool   `yaml:"send"`
		SendDebug bool   `yaml:"senddebug"`
		ChatId    int64  `yaml:"chatid"`
		Token     string `yaml:"token"`
	} `yaml:"telegram"`
	VK struct {
		Send    bool   `yaml:"send"`
		Token   string `yaml:"token"`
		OwnerId int64  `yaml:"ownerid"`
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
	FileParse  string `short:"f" long:"fileparse" description:"File for parce (rss xml)" required:"true"`
	ConfigPath string `short:"c" long:"configpath" description:"Config file path"`
	InitDB     bool   `short:"i" long:"initdb" description:"Run initialize from current file"`
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

type SendItems struct {
	ItemList []Item
}

var senditems SendItems

func FindItems(rss Rss2, dbpath string) {
	db, err := bolt.Open(dbpath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for _, v := range rss.ItemList {
		db.View(func(tx *bolt.Tx) error {
			// Assume bucket exists and has keys
			b := tx.Bucket([]byte("rss"))
			c := b.Cursor()
			flag := false
			for key, _ := c.First(); key != nil; key, _ = c.Next() {
				if v.Link == string(key) {
					flag = true
					break
				}
			}
			if flag != true {
				senditems.ItemList = append(senditems.ItemList, v)
			}
			return nil
		})
	}
}

func InitDb(rss Rss2, dbpath string) {
	log.Println("Initialize DB")
	db, err := bolt.Open(dbpath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("rss"))
		if err != nil {
			return err
		}
		for _, v := range rss.ItemList {
			encoded, err := json.Marshal(v)
			if err != nil {
				return err
			}
			err = b.Put([]byte(v.Link), encoded)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (config Config) RunSend() {
	for _, v := range senditems.ItemList {
		if config.Telegram.Send {
			log.Println("Send to telegram")
			bot, err := tgbotapi.NewBotAPI(config.Telegram.Token)
			if err != nil {
				log.Panic(err)
			}
			bot.Debug = config.Telegram.SendDebug

			s := "<b>" + string(v.Title) + "</b>\n" + html.UnescapeString(string(v.Description)) +
				"\n" + v.Link
			msg := tgbotapi.NewMessage(config.Telegram.ChatId, s)
			msg.ParseMode = "Html"
			_, err = bot.Send(msg)
			if err != nil {
				log.Panic(err)

			}
			log.Println("Sended to telegram")
		}
		if config.VK.Send {
			log.Println("Send to VK")
			vk := api.NewVK(config.VK.Token)
			_, err := vk.WallPost(api.Params{
				"owner_id":    config.VK.OwnerId,
				"attachments": v.Link,
			})
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Sended to VK")
		}
		if config.Facebook.Send {
			log.Println("Send to Facebook")
			log.Println("Sending to facebook is not implemented yet")
		}
	}
}

func UpdateDb(dbpath string) {
	log.Println("Update DB")
	db, err := bolt.Open(dbpath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("rss"))
		for _, v := range senditems.ItemList {
			encoded, err := json.Marshal(v)
			if err != nil {
				return err
			}
			err = b.Put([]byte(v.Link), encoded)
			if err != nil {
				return err
			}
		}
		return nil
	})

}

func main() {
	log.Println("Run processing")
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
	log.Println("Flags processed")

	if options.ConfigPath != "" {
		log.Printf("Config from: %s\n", options.ConfigPath)
		ConfigPath = options.ConfigPath
	}

	// Get config
	cfg, err := NewConfig(ConfigPath)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Config processed")

	// Parse rss file
	log.Printf("Parse file %s \n", options.FileParse)
	rss, err := NewRSS(options.FileParse)
	if err != nil {
		log.Fatal(err)
	}
	if options.InitDB {
		InitDb(*rss, cfg.Dbpath)
	} else {
		//Find new items
		FindItems(*rss, cfg.Dbpath)

		if len(senditems.ItemList) > 0 {
			// Run send data depended on configuration options
			log.Println("Run send process")
			cfg.RunSend()
			UpdateDb(cfg.Dbpath)
		}
	}
	log.Println("End processing")
}
