package gmr

import (
    "github.com/PuerkitoBio/goquery"
    "io"
    "net/http"
    "strconv"
    "time"
)

const gmr_url = "http://multiplayerrobot.com/"
const timeout = 10 // Max time in seconds to wait for gmr
const concurrency = 5 // Max concurrent requests to gmr

type Player struct {
    name string
    civ string
    turn_order int
}

type Game struct {
    id string
    name string
    players []Player // First player is current player
}

func GetGame(game_id string, c chan Game) {
    // fmt.Println("Getting game " + game_id)
    url := gmr_url + "Game/Details?id=" + game_id

    // HTTP stuffs
    client := &http.Client{}
    req, _ := http.NewRequest("POST", url, nil)
    req.Header.Set("Content-Length", "0")
    resp, _ := client.Do(req)
    defer resp.Body.Close()

    // Parsing stuffs
    body, _ := goquery.NewDocumentFromReader(io.Reader(resp.Body))
    var players []Player

    // Grab game name
    game_name := body.Find(".name-block").Find("a").Text()

    // Grab list of players - annoyingly this does not include the active player
    // Add each player to the slice of players
    body.Find(".game-players").Find(".game-player-container").Each(func(i int, s *goquery.Selection) {
        gp := s.Find(".game-player")
        civ, _ := gp.Find(".civ-icon").Attr("title")
        turn_order, _ := gp.Find(".turn-order").Attr("value")
        turn_order_i, _ := strconv.Atoi(turn_order)
        name, _ := gp.Find("a").Find("img").Attr("title")
        player := Player{name, civ, turn_order_i}
        players = append(players, player)
    })

    // Turn order is not included in the current player properties. Derive it from the non-active players.
    var current_player_turn_order int
    if players[0].turn_order == 0 {
        current_player_turn_order = len(players) - 1
    } else {
        current_player_turn_order = players[0].turn_order - 1
    }

    // Grab info about current player
    current_player_name, _ := body.Find(".game-host").Find(".avatar").Attr("title")
    current_player_civ, _ := body.Find(".game-host").Find("img").Attr("title")
    current_player := Player{current_player_name, current_player_civ, current_player_turn_order}

    players = append([]Player{current_player}, players...)

    c <- Game { game_id, game_name, players }
}

func GetGames(game_ids ...string) []Game {
    t := time.After(timeout * time.Second)
    c := make(chan Game, 64)
    n := 0
    var games []Game
    for _, game := range game_ids {
        n++
        go GetGame(game, c)
    }
    for i := 0; i < n; i++ {
        select {
            case result := <- c:
                games = append(games, result)
            case <- t:
                return games
        }
    }
    return games
}
