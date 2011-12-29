package house

import (
  "glop/util/algorithm"
  "glop/gui"
  "glop/gin"
  "fmt"
)

type WallPanel struct {
  *gui.VerticalTable
  room *Room
  viewer *RoomViewer

  wall_texture *WallTexture
  prev_wall_texture *WallTexture
  drag_anchor struct{ X,Y float32 }
  drop_on_release bool
  selected_walls map[int]bool
}

func MakeWallPanel(room *Room, viewer *RoomViewer) *WallPanel {
  var wp WallPanel
  wp.room = room
  wp.viewer = viewer
  wp.VerticalTable = gui.MakeVerticalTable()
  wp.selected_walls = make(map[int]bool)

  fnames := GetAllWallTextureNames()
  for i := range fnames {
    name := fnames[i]
    wp.VerticalTable.AddChild(gui.MakeButton("standard", name, 300, 1, 1, 1, 1, func(t int64) {
      wt := MakeWallTexture(name)
      if wt == nil { return }
      wp.viewer.Temp.WallTexture = wt
      wp.viewer.Temp.WallTexture.X = 5
      wp.viewer.Temp.WallTexture.Y = 5
      wp.drag_anchor.X = 0
      wp.drag_anchor.Y = 0
      wp.drop_on_release = false
    }))
  }

  return &wp
}

func (w *WallPanel) textureNear(wx,wy int) *WallTexture {
  for _,tex := range w.room.WallTextures {
    ax,ay,_ := w.viewer.modelviewToBoard(float32(wx), float32(wy))
    bx,by,_ := w.viewer.modelviewToLeftWall(float32(wx), float32(wy))
    cx,cy,_ := w.viewer.modelviewToRightWall(float32(wx), float32(wy))
    dx := float32(tex.texture_data.Dx) / 100
    dy := float32(tex.texture_data.Dy) / 100
    for _,p := range [][2]float32{ {ax,ay}, {bx,by}, {cx,cy} } {
      // fmt.Printf("Checking %v against %f %f\n", p, tex.X, tex.Y)
      if p[0] > tex.X - dx && p[0] < tex.X + dx && p[1] > tex.Y - dy && p[1] < tex.Y + dy {
        return tex
      }
    }
  }
  return nil
}

func (w *WallPanel) Respond(ui *gui.Gui, group gui.EventGroup) bool {
  if w.VerticalTable.Respond(ui, group) {
    return true
  }
  if found,event := group.FindEvent(gin.Escape); found && event.Type == gin.Press {
    if w.viewer.Temp.WallTexture != nil {
      w.viewer.Temp.WallTexture = nil
    }
    return true
  }
  if found,event := group.FindEvent(gin.MouseLButton); found {
    if w.viewer.Temp.WallTexture != nil && (event.Type == gin.Press || (event.Type == gin.Release && w.drop_on_release)) {
      w.room.WallTextures = append(w.room.WallTextures, w.viewer.Temp.WallTexture)
      w.viewer.Temp.WallTexture = nil
    } else if w.viewer.Temp.WallTexture == nil && event.Type == gin.Press {
      w.viewer.Temp.WallTexture = w.textureNear(event.Key.Cursor().Point())
      fmt.Printf("Tex: %v\n", w.viewer.Temp.WallTexture)
      // w.viewer.Temp.WallTexture = w.viewer.SelectWallTextureAt(event.Key.Cursor().Point())
      if w.viewer.Temp.WallTexture != nil {
        w.prev_wall_texture = new(WallTexture)
        *w.prev_wall_texture = *w.viewer.Temp.WallTexture
      }
      w.room.WallTextures = algorithm.Choose(w.room.WallTextures, func(a interface{}) bool {
        return a.(*WallTexture) != w.viewer.Temp.WallTexture
      }).([]*WallTexture)
      w.drop_on_release = true
      if w.viewer.Temp.WallTexture != nil {
        wx,wy := w.viewer.BoardToWindow(float32(w.viewer.Temp.WallTexture.X), float32(w.viewer.Temp.WallTexture.Y))
        px,py := event.Key.Cursor().Point()
        w.drag_anchor.X = float32(px) - wx
        w.drag_anchor.Y = float32(py) - wy
      }
    }
    return true
  }
  return false
}

func (w *WallPanel) Think(ui *gui.Gui, t int64) {
  if w.viewer.Temp.WallTexture != nil {
    px,py := gin.In().GetCursor("Mouse").Point()
    tx := float32(px) - w.drag_anchor.X
    ty := float32(py) - w.drag_anchor.Y
    bx,by := w.viewer.WindowToBoard(int(tx), int(ty))
    w.viewer.Temp.WallTexture.X = bx
    w.viewer.Temp.WallTexture.Y = by
  }
  w.VerticalTable.Think(ui, t)
}

func (w *WallPanel) Collapse() {
  if w.viewer.Temp.WallTexture != nil && w.prev_wall_texture != nil {
    w.room.WallTextures = append(w.room.WallTextures, w.prev_wall_texture)
  }
  w.prev_wall_texture = nil
  w.viewer.Temp.WallTexture = nil
}

func (w *WallPanel) Expand() {

}
