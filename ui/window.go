package ui

import (
	"image"
	"image/color"
	"log"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/imageutil"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/draw"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

type Visualizer struct {
	Title         string
	Debug         bool
	OnScreenReady func(s screen.Screen)

	w    screen.Window
	tx   chan screen.Texture
	done chan struct{}

	sz  size.Event
	pos image.Rectangle
}

func (pw *Visualizer) Main() {
	pw.tx = make(chan screen.Texture)
	pw.done = make(chan struct{})
	pw.pos.Max.X = 200
	pw.pos.Max.Y = 200
	driver.Main(pw.run)
}

func (pw *Visualizer) Update(t screen.Texture) {
	pw.tx <- t
}

func (pw *Visualizer) run(s screen.Screen) {
	w, err := s.NewWindow(&screen.NewWindowOptions{
		Title:  pw.Title,
		Width:  800,
		Height: 800,
	})
	if err != nil {
		log.Fatal("Failed to initialize the app window:", err)
	}
	defer func() {
		w.Release()
		close(pw.done)
	}()

	if pw.OnScreenReady != nil {
		pw.OnScreenReady(s)
	}

	pw.w = w

	events := make(chan any)
	go func() {
		for {
			e := w.NextEvent()
			if pw.Debug {
				log.Printf("new event: %v", e)
			}
			if detectTerminate(e) {
				close(events)
				break
			}
			events <- e
		}
	}()

	var t screen.Texture

	for {
		select {
		case e, ok := <-events:
			if !ok {
				return
			}
			pw.handleEvent(e, t)
			log.Printf("Window size: %v", pw.sz)

		case t = <-pw.tx:
			w.Send(paint.Event{})
		}
	}
}

func detectTerminate(e any) bool {
	switch e := e.(type) {
	case lifecycle.Event:
		if e.To == lifecycle.StageDead {
			return true // Window destroy initiated.
		}
	case key.Event:
		if e.Code == key.CodeEscape {
			return true // Esc pressed.
		}
	}
	return false
}

func (pw *Visualizer) handleEvent(e any, t screen.Texture) {
	switch e := e.(type) {
	case size.Event:
		pw.sz = e
	case error:
		log.Printf("ERROR: %s", e)
	case mouse.Event:
		if e.Button == mouse.ButtonLeft && e.Direction == mouse.DirPress {
			// Переміщення центру фігури в нове місце по натисканню лівої кнопки миші.
			// В якості прикладу переміщимо його в координати натискання миші.
			pw.pos.Min.X = int(e.X) - 200     // Від центру віднімаємо половину ширини фігури
			pw.pos.Min.Y = int(e.Y) - 125     // Від центру віднімаємо половину висоти фігури
			pw.pos.Max.X = pw.pos.Min.X + 400 // Ширина фігури
			pw.pos.Max.Y = pw.pos.Min.Y + 250 // Висота фігури
			pw.w.Send(paint.Event{})
		}

	case paint.Event:
		// Малювання контенту вікна.
		if t == nil {
			pw.drawDefaultUI()
		} else {
			// Використання текстури отриманої через виклик Update.
			pw.w.Scale(pw.sz.Bounds(), t, t.Bounds(), draw.Src, nil)
		}
		pw.w.Publish()
	}
}

func (pw *Visualizer) drawDefaultUI() {
	pw.w.Fill(pw.sz.Bounds(), color.Black, draw.Src) // Фон.

	// Розміри прямокутників.
	horizontalRectWidth := 400
	horizontalRectHeight := 20
	verticalRectWidth := 40
	verticalRectHeight := 200

	// Координати верхнього лівого кута горизонтального прямокутника.
	horizontalRectX := pw.pos.Min.X
	horizontalRectY := pw.pos.Min.Y

	// Малюємо горизонтальний прямокутник.
	horizontalRect := image.Rect(horizontalRectX, horizontalRectY, horizontalRectX+horizontalRectWidth, horizontalRectY+horizontalRectHeight)
	pw.w.Fill(horizontalRect, color.White, draw.Src)

	// Координати верхнього лівого кута вертикального прямокутника.
	verticalRectX := pw.pos.Min.X + 180
	verticalRectY := pw.pos.Min.Y + 20

	// Малюємо вертикальний прямокутник.
	verticalRect := image.Rect(verticalRectX, verticalRectY, verticalRectX+verticalRectWidth, verticalRectY+verticalRectHeight)
	pw.w.Fill(verticalRect, color.White, draw.Src)

	// Опубліковуємо зміни.
	pw.w.Publish()

	// Малювання білої рамки.
	for _, br := range imageutil.Border(pw.sz.Bounds(), 10) {
		pw.w.Fill(br, color.White, draw.Src)
	}

}
