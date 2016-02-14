package gmr

type ApiGamesResponse struct {
    CurrentTotalPoints  int
    Games               []ApiGameResponse
    Players             []ApiPlayerResponse
}

type ApiGameResponse struct {
    CurrentTurn ApiCurrentTurn
    GameId      int
    Name        string
    Players     []ApiGamePlayer
}

type ApiCurrentTurn struct {
    Expires      string
    IsFirstTurn  bool
    Number       int
    PlayerNumber int
    Skipped      bool
    Started      string
    TurnId       int
    UserId       int
}

type ApiGamePlayer struct {
    TurnOrder int
    UserId    int
}
type ApiPlayerResponse struct {
    AvatarUrl    string
    GameId       int
    PersonaName  string
    PersonaState int
    SteamID      int
}

