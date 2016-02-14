package gmr

import (
    "bytes"
    "encoding/json"
    "fmt"
    "github.com/cheggaaa/pb"
    "github.com/PuerkitoBio/goquery"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"
)

var cache struct {
    my_id string
    games ApiGamesResponse
}

const gmr_url = "http://multiplayerrobot.com/"
const timeout = 10 // Max time in seconds to wait for gmr
const concurrency = 5 // Max concurrent requests to gmr

type Player struct {
    Name       string
    Civ        string
    Turn_order int
    Id         string
}

type Game struct {
    Id      string
    Name    string
    Players []Player // First player is current player
}

func (r ApiGamesResponse) IsMyTurn(game_id string) string {
    game_found := false
    turn_id := ""
    for _, g := range r.Games {
        if strconv.Itoa(g.GameId) == game_id {
            log.Printf("debug: Found game %s, current turn is for user %d", game_id, g.CurrentTurn.UserId)
            if strconv.Itoa(g.CurrentTurn.UserId) != cache.my_id {
                return ""
            }
            turn_id = strconv.Itoa(g.CurrentTurn.TurnId)
        }
    }
    if ! game_found {
        log.Printf("info: You are not playing in game %s", game_id)
        return ""
    }
    return turn_id
}

func Authenticate(api_key string) string {
    url := gmr_url + "Api/Diplomacy/AuthenticateUser?authKey=" + api_key
    log.Printf("debug: -> GET %s", url)
    resp, _ := http.Get(url)

    if resp.StatusCode != 200 {
        log.Printf("error: GMR returned status %d", resp.StatusCode)
        return ""
    }
    content, _ := ioutil.ReadAll(resp.Body)
    s_content := string(content)
    cache.my_id = s_content

    log.Printf("debug: [%d] %s", resp.StatusCode, s_content)

    if s_content == "null" {
        return ""
    }
    return s_content
}

func GetMyGames(api_key string) ApiGamesResponse {
    my_id := Authenticate(api_key)
    empty := ApiGamesResponse{}

    if my_id == "" {
        return empty
    }
    url := gmr_url + fmt.Sprintf("Api/Diplomacy/GetGamesAndPlayers?playerIDTEXT=%s&authKey=%s", my_id, api_key)
    log.Printf("debug: -> GET %s", url)
    resp, _ := http.Get(url)
    if resp.StatusCode != 200 {
        log.Printf("error: GMR returned status %d", resp.StatusCode)
        return empty
    }
    content, _ := ioutil.ReadAll(resp.Body)

    var data ApiGamesResponse
    err := json.Unmarshal(content, &data)
    if err != nil {
        log.Printf("error: could not parse GMR response")
        return empty
    }

    var player_ids []string
    for _, g := range data.Games {
        for _, p := range g.Players {
            if s_id := strconv.Itoa(p.UserId); s_id != "0" {
                player_ids = append(player_ids, s_id)
            }
        }
    }
    s_player_ids := strings.Join(player_ids, "_")

    // Now we have all PlayerIDs we want to fetch, do it again.
    url = gmr_url + fmt.Sprintf("Api/Diplomacy/GetGamesAndPlayers?playerIDTEXT=%s&authKey=%s", s_player_ids, api_key)

    log.Printf("debug: -> GET %s", url)
    resp, _ = http.Get(url)
    if resp.StatusCode != 200 {
        log.Printf("error: GMR returned status %d", resp.StatusCode)
        return empty
    }
    content, _ = ioutil.ReadAll(resp.Body)

    var return_data ApiGamesResponse
    err = json.Unmarshal(content, &return_data)
    if err != nil {
        log.Printf("error: could not parse GMR response")
        return empty
    }

    return return_data
}

func GetGame(game_id string, c chan Game) {
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
        user_id, _ := gp.Find(".user-id").Attr("value")
        player := Player{name, civ, turn_order_i, user_id}
        players = append(players, player)
    })

    // Turn order is not included in the current player properties. Derive it from the non-active players.
    var current_player_turn_order int
    if players[0].Turn_order == 0 {
        current_player_turn_order = len(players) - 1
    } else {
        current_player_turn_order = players[0].Turn_order - 1
    }

    // Grab info about current player
    current_player_name, _ := body.Find(".game-host").Find(".avatar").Attr("title")
    current_player_civ, _ := body.Find(".game-host").Find("img").Attr("title")
    current_player := Player{current_player_name, current_player_civ, current_player_turn_order, "0"}

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

func GetSaveFile(api_key string, game_id string, download_to string, progressbar bool) {
    games := GetMyGames(api_key)
    if games.IsMyTurn(game_id) == "" {
        log.Println("error: It's not your turn!")
        return
    }
    url := gmr_url + fmt.Sprintf("api/Diplomacy/GetLatestSaveFileBytes?authKey=%s&gameId=%s", api_key, game_id)
    out_path := fmt.Sprintf("%s/GMR-%s.Civ5Save", download_to, game_id)
    out, _ := os.Create(out_path)
    defer out.Close()

    resp, err := http.Get(url)
    defer resp.Body.Close()

    if err != nil {
        log.Print("error:", err)
        return
    }

    if resp.StatusCode != 200 {
        log.Printf("error: GMR returned status %d. Verify your auth key and that the game exists.", resp.StatusCode)
        return
    }

    length, err := strconv.Atoi(resp.Header.Get("Content-Length"))
    if err != nil {
        log.Print("error:", err)
        return
    }
    in := resp.Body

    if progressbar {
        bar := pb.New(length).SetUnits(pb.U_BYTES)
        bar.Start()
        in = ioutil.NopCloser(bar.NewProxyReader(resp.Body))
    }

    n, _ := io.Copy(out, in)
    fmt.Printf("%d bytes downloaded to %s\n", n, out_path)
}

func SubmitSaveFile(api_key string, game_id string, upload_from string, progressbar bool) {
    games := GetMyGames(api_key)
    log.Printf("debug: My ID: %s", cache.my_id)

    if _, err := os.Stat(upload_from); os.IsNotExist(err) {
        log.Printf("error: %s does not exist", upload_from)
        return
    }

    game_found := false
    var turn_id string

    for _, g := range games.Games {
        if strconv.Itoa(g.GameId) == game_id {
            game_found = true
            log.Printf("debug: Found game %s, current turn is for user %d", game_id, g.CurrentTurn.UserId)
            if strconv.Itoa(g.CurrentTurn.UserId) != cache.my_id {
                log.Print("error: It's not your turn")
                return
            }
            turn_id = strconv.Itoa(g.CurrentTurn.TurnId)
        }
    }
    if ! game_found {
        log.Printf("error: You are not playing in game %s", game_id)
        return
    }
    log.Println(turn_id)

    url := gmr_url + fmt.Sprintf("api/Diplomacy/SubmitTurn?authKey=%s&turnId=%s", api_key, turn_id)
    client := &http.Client{}

    log.Printf("debug: -> POST %s", url)

    fi, _ := os.Stat(upload_from)
    data, _ := ioutil.ReadFile(upload_from)
    length := fi.Size()
    in := bytes.NewReader(data)

    // if progressbar {
    //    bar := pb.New(int(length)).SetUnits(pb.U_BYTES)
    //    bar.Start()
    //    //in = ioutil.NopCloser(bar.NewProxyReader(in))
    //    in = bar.NewProxyReader(in)
    //}

    req, _ := http.NewRequest("POST", url, in)
    req.Header.Set("Content-Length", strconv.Itoa(int(length)))

    resp, err := client.Do(req)
    log.Println(resp, err)
}
