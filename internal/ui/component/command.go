package component

import (
	"github.com/raphael-goetz/lazysound/internal/ui/style"
)

func Command(s style.Styles) string {
	return s.CmdBar.Render(
		"h/l pane   j/k move   1-3 tabs   tab cycle   q quit",
	)
}
