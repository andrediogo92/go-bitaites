package vue

import (
	"github.com/Seriyin/go-bitaites/timeline"
	"github.com/zserge/webview"
)

func Start(timeline timeline.Timeline) {
	w := webview.New(webview.Settings{
		Title: "Bitaites",
		Resizable: true,
	})
	b, err := static.ReadFile("statics/deps/bootstrap.min.css")
	if (err == nil) {
		w.Dispatch(func() {
			w.InjectCSS(string(b))
		})
		b, err := static.ReadFile("statics/deps/bootstrap.min.js")
		if (err == nil) {
			w.Dispatch(func() {
				w.Eval(string(b))
			})
			b, err := static.ReadFile("statics/deps/jquery-3.3.1.slim.min.js")
			if (err == nil) {
				w.Dispatch(func() {
					w.Eval(string(b))
				})
				b, err := static.ReadFile("statics/deps/popper.min.js")
				if (err == nil) {
					w.Dispatch(func() {
						w.Eval(string(b))
					})
					b, err := static.ReadFile("statics/bitaites.js")
					if (err == nil) {
						w.Bind("timeline", timeline)
						w.Dispatch(func() {
							w.Eval(string(b))
						})
					}
				}
			}
		}
	}
}
