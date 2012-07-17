// The Player struct defines everything that is specific to a particular
// human player.  This includes their progress through the campaign, as well
// as settings and preferences.
package game

import (
  "fmt"
  "io"
  "os"
  "bytes"
  "hash/fnv"
  "path/filepath"
  "github.com/runningwild/haunts/base"
  "encoding/gob"
  lua "github.com/xenith-studios/golua"
)

type Player struct {
  // Name of the player, as specified by the player himself, this is what is
  // shown in the menu when they are selecting what player to switch to.
  Name string

  // This is the value of the global table named 'store' in the lua scripts.
  // Serialied/deserialized with LuaEncodeTable/LuaDecodeTable
  // This data persists for the lifetime of the player.
  Lua_store []byte

  // This is the value of the global table named 'level' in the lua scripts.
  // This data persists for the duration of a single level.
  Lua_level []byte

  // Game data - if the player is in the middle of a game then the state is
  // stored here.
  // Game *game
}

// Returns a map from player name to the path of that player's file.
func GetAllPlayers() map[string]string {
  root := filepath.Join(base.GetDataDir(), "players")
  players := make(map[string]string)
  filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
    if err != nil || info.IsDir() {
      return nil
    }
    f, err := os.Open(path)
    if err != nil {
      base.Warn().Printf("Unable to open player file: %s.", path)
      return nil
    }
    defer f.Close()
    dec := gob.NewDecoder(f)
    var name string
    err = dec.Decode(&name)
    if err != nil {
      base.Warn().Printf("Unable to read player file: %s.", path)
      return nil
    }
    if err != nil {
      players[name] = path
    }
    return nil
  })
  return players
}

func UpdatePlayer(p *Player, L *lua.State) {
  buffer := bytes.NewBuffer(nil)
  L.GetGlobal("store")
  LuaEncodeTable(buffer, L, -1)
  L.Pop(1)
  p.Lua_store = buffer.Bytes()
}

// Encode a player's name, then the entire player.  This way we can just read
// the first value to get it's name without having to de-gob the entire file.
func EncodePlayer(w io.Writer, p *Player) error {
  enc := gob.NewEncoder(w)
  err := enc.Encode(p.Name)
  if err != nil {
    return err
  }
  return enc.Encode(p)
}

func DecodePlayer(r io.Reader) (*Player, error) {
  var p Player
  dec := gob.NewDecoder(r)
  err := dec.Decode(&p.Name)
  if err != nil {
    return nil, err
  }
  err = dec.Decode(&p)
  return &p, err
}

func LoadPlayer(path string) (*Player, error) {
  f, err := os.Open(path)
  if err != nil {
    return nil, err
  }
  defer f.Close()
  return DecodePlayer(f)
}

func SavePlayer(p *Player) error {
  hash := fnv.New64()
  hash.Write([]byte(p.Name))
  name := fmt.Sprintf("%x.player", hash.Sum64())
  f, err := os.Create(filepath.Join(base.GetDataDir(), "players", name))
  if err != nil {
    return err
  }
  defer f.Close()
  base.SetStoreVal("last player", name)
  return EncodePlayer(f, p)
}