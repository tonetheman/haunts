package status

import (
  "bytes"
  "encoding/json"
  "encoding/gob"
)

type Kind string
const (
  Panic  Kind = "Panic"
  Terror      = "Terror"
  Fire        = "Fire"
  Poison      = "Poison"
)
type Primary int
const (
  Ego    Primary = iota
  Corpus
)
func (k Kind) Primary() Primary {
  switch k {
  case Panic: fallthrough
  case Terror:
    return Ego

  case Fire: fallthrough
  case Poison:
    return Corpus
  }
  panic("Unknown status.Kind")
}

type DoesntModifyBase struct {}
func (DoesntModifyBase) ModifyBase(b Base) Base {
  return b
}

type PermanentEffect struct {}
func (PermanentEffect) Think() (bool) {
  return false
}

// A RoundTimer will last for the specified number of rounds.  A RoundTimer of
// 0 is equivalent to an Immediate effect.
type RoundTimer struct {
  Num_rounds int
}
func (r *RoundTimer) Think() bool {
  r.Num_rounds--
  return r.Num_rounds < 0
}


// Conditions represent instantaneous or ongoing Conditions on an entity.
// Every round the Condition can 
type Condition interface {
  // Called any time a Base stat is queried
  ModifyBase(Base) Base

  // Called at the beginning of each round.  May return a damage object to
  // deal damage, and must return a bool indicating whether this effect has
  // completed or not.
  Think() (complete bool)
}

type Dynamic struct {
  Hp,Ap int
}

type Base struct {
  Ap_max int
  Hp_max int
  Corpus int
  Ego    int
  Sight  int
}

type inst struct {
  Base
  Dynamic
  Conditions []Condition
}

type Inst struct {
  // This prevents external code from modifying any data without goin through
  // the appropriate methods, but also allows us to provide accurate json and
  // gob methods.
  inst inst
}

func (s Inst) modifiedBase() Base {
  b := s.inst.Base
  for _,e := range s.inst.Conditions {
    b = e.ModifyBase(b)
  }
  return b
}

func (s Inst) HpCur() int {
  return s.inst.Dynamic.Hp
}

func (s Inst) ApCur() int {
  return s.inst.Dynamic.Ap
}

func (s Inst) HpMax() int {
  hp_max := s.modifiedBase().Hp_max
  if hp_max < 0 { return 0 }
  return hp_max
}

func (s Inst) ApMax() int {
  ap_max := s.modifiedBase().Ap_max
  if ap_max < 0 { return 0 }
  return ap_max
}

func (s Inst) Corpus() int {
  corpus := s.modifiedBase().Corpus
  if corpus < 0 { return 0 }
  return corpus
}

func (s Inst) Ego() int {
  ego := s.modifiedBase().Ego
  if ego < 0 { return 0 }
  return ego
}

func (s Inst) Sight() int {
  sight := s.modifiedBase().Sight
  if sight < 0 { return 0 }
  return sight
}

func (s *Inst) ApplyEffect(e Condition) {
  s.inst.Conditions = append(s.inst.Conditions, e)
  // s.Dynamic = e.ModifyDynamic(s.Dynamic)
}

func (s *Inst) Think() {
  complete := 0
  for i := 0; i < len(s.inst.Conditions); i++ {
    if s.inst.Conditions[i].Think() {
      complete++
    } else {
      s.inst.Conditions[i - complete] = s.inst.Conditions[i]
    }
  }
  s.inst.Conditions = s.inst.Conditions[0 : len(s.inst.Conditions) - complete]

  // Now that we've removed completed Conditions we can set our dynamic stats
  // accordingly
  s.inst.Ap = s.ApMax()

  // for _,e := range s.Conditions {
  //   s.Dynamic = e.ModifyDynamic(s.Dynamic)
  // }

  // And now that we've modified our dynamic stats we can make sure they lie
  // within the appropriate range.
  if s.inst.Ap < 0 {
    s.inst.Ap = 0
  }
  if s.inst.Ap > s.ApMax() {
    s.inst.Ap = s.ApMax()
  }

  if s.inst.Hp < 0 {
    s.inst.Hp = 0
  }
  if s.inst.Hp > s.HpMax() {
    s.inst.Hp = s.HpMax()
  }
}

// Encoding routines - only support json and gob right now

func (si Inst) MarshalJSON() ([]byte, error) {
  return json.Marshal(si.inst)
}

func (si *Inst) UnmarshalJSON(data []byte) error {
  println("Unmarshaling ", string(data))
  return json.Unmarshal(data, &si.inst)
}

func (si Inst) GobEncode() ([]byte, error) {
  buf := bytes.NewBuffer(nil)
  enc := gob.NewEncoder(buf)
  err := enc.Encode(si.inst)
  return buf.Bytes(), err
}

func (si *Inst) GobDecode(data []byte) error {
  dec := gob.NewDecoder(bytes.NewBuffer(data))
  return dec.Decode(&si.inst)
}
