package mrgnet

import (
  "crypto/rand"
  "math/big"
)

type NetId int64

const Host_url = "http://localhost:8080"

// Creates a random id that will be unique among all other engines with high
// probability.
func RandomId() NetId {
  b := big.NewInt(1 << 62)
  v, err := rand.Int(rand.Reader, b)
  if err != nil {
    // uh-oh
    panic(err)
  }
  return NetId(v.Int64())
}

type User struct {
  Id   NetId
  Name string
}

type UpdateUserRequest User
type UpdateUserResponse struct {
  User
  Err string
}

type NewGameRequest struct {
  Id     NetId
  Script string
}

type NewGameResponse struct {
  Err  string
  Name string
  Id   string
}

type ListGamesRequest struct {
  Id        NetId
  Unstarted bool
}

type ListGamesResponse struct {
  Err   string
  Games []Game
  Ids   []string
}

type JoinGameRequest struct {
  Id       NetId
  Game_key string
}

type JoinGameResponse struct {
  Err        string
  Successful bool
}

type Game struct {
  Name string

  Denizens_name  string
  Denizens_id    NetId
  Intruders_name string
  Intruders_id   NetId

  Playbacks []Playback

  // If this is non-zero then the game is over and the winner is the player
  // whose NetId matches this value
  Winner NetId
}

// State is the state of the game before the playback.
// The state of the game after the playback is the State field of the next
// Playback.
type Playback struct {
  State []byte
  Execs []byte
}
