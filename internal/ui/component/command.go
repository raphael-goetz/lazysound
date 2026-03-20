package component

import "github.com/raphael-goetz/lazysound/internal/ui/style"

func Command(s style.Styles, searchNav bool) string {
	base := "h/l pane   j/k move   enter open   p play   space pause   ,/. seek   x shuffle   R repeat   s stop   r restart   +/- volume   a actions   tab cycle   q quit"
	if searchNav {
		base = "h/l pane   j/k move   enter open   p play   space pause   ,/. seek   x shuffle   R repeat   s stop   r restart   +/- volume   a actions   tab cycle   / search   t/p mode   q quit"
	}
	return s.CmdBar.Render(base)
}
