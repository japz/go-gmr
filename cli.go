package main

import (
    "fmt"
    "github.com/japz/go-gmr/gmr"
    "gopkg.in/alecthomas/kingpin.v2"
    "os"
)

var (
    app         = kingpin.New("gmr", "A multiplayer robot command line client")
    debug       = app.Flag("debug", "Enable debug mode").Bool()

    games_cmd   = app.Command("games", "Get information on games")
    games       = games_cmd.Arg("games", "Games to fetch").Strings()
)


func main() {
    switch kingpin.MustParse(app.Parse(os.Args[1:])) {
        case games_cmd.FullCommand():
            fmt.Println(gmr.GetGames(*games...))
    }
}
