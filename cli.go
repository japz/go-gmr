package main

import (
    "fmt"
    "github.com/japz/go-gmr/gmr"
    "gopkg.in/alecthomas/kingpin.v2"
    "encoding/json"
    "encoding/xml"
    "os"
)

var (
    app          = kingpin.New("gmr", "A multiplayer robot command line client")
    debug        = app.Flag("debug", "Enable debug mode").Bool()
    format       = app.Flag("format", "Output format (text / json / xml)").Default("text").String()
    key          = app.Flag("key", "API key").Default("").String()

    games_cmd    = app.Command("games", "Get information on games")
    games        = games_cmd.Arg("games", "Games to fetch").Strings()

    download_cmd = app.Command("download", "Download save file")
    game_id      = download_cmd.Arg("game_id", "Game id to fetch").String()
    download_to  = download_cmd.Arg("path", "Path to download savefile to").Default(".").String()
)


func main() {
    switch kingpin.MustParse(app.Parse(os.Args[1:])) {
        case games_cmd.FullCommand():
            games := gmr.GetGames(*games...)
            switch *format {
                case "text":
                    fmt.Println(games)
                case "json":
                    data, _ := json.Marshal(games)
                    fmt.Println(string(data))
                case "xml":
                    data, _ := xml.MarshalIndent(games, "", "  ")
                    fmt.Println(string(data))
            }
        case download_cmd.FullCommand():
            if *key == "" {
                fmt.Println("API key required")
                return
            }
            gmr.GetSaveFile(*key, *game_id, *download_to, true)
    }
}
