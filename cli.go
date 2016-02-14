package main

import (
    "encoding/json"
    "encoding/xml"
    "fmt"
    "github.com/comail/colog"
    "github.com/japz/go-gmr/gmr"
    "gopkg.in/alecthomas/kingpin.v2"
    "log"
    "os"
)

var (
    app          = kingpin.New("gmr", "A multiplayer robot command line client")
    debug        = app.Flag("debug", "Enable debug mode").Bool()
    format       = app.Flag("format", "Output format (text / json / xml)").Default("text").String()
    key          = app.Flag("key", "API key").Default("").String()

    auth_cmd     = app.Command("auth", "Authenticate to GMR (use this to test your key)")

    games_cmd    = app.Command("games", "Get information on games")
    games        = games_cmd.Arg("games", "Games to fetch").Strings()

    download_cmd = app.Command("download", "Download save file")
    d_game_id    = download_cmd.Arg("game_id", "Game id to fetch").Required().String()
    download_to  = download_cmd.Arg("path", "Path to download savefile to").Default(".").String()

    submit_cmd   = app.Command("submit", "Submit save file")
    s_game_id    = submit_cmd.Arg("game_id", "Game id to submit").Required().String()
    submit_from  = submit_cmd.Arg("path", "Path to savefile").Required().String()
)

func main() {
    parsed := kingpin.MustParse(app.Parse(os.Args[1:]))

    colog.Register()
    colog.SetOutput(os.Stderr)
    colog.SetMinLevel(colog.LInfo)
    colog.SetDefaultLevel(colog.LInfo)
    if *debug {
        colog.SetMinLevel(colog.LDebug)
    }

    switch parsed {
        case auth_cmd.FullCommand():
            if *key == "" {
                fmt.Println("API key required")
                return
            }
            resp := gmr.Authenticate(*key)
            if resp == "" {
                log.Printf("error: Authentication unsuccessful")
            } else {
                log.Printf("info: Authentication successful (%s)", resp)
            }
            gmr.GetMyGames(*key)
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
            gmr.GetSaveFile(*key, *d_game_id, *download_to, true)
        case submit_cmd.FullCommand():
            if *key == "" {
                fmt.Println("API key required")
                return
            }
            gmr.SubmitSaveFile(*key, *s_game_id, *submit_from, true)
    }
}
