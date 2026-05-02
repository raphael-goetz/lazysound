package layout

import "github.com/raphael-goetz/lazysound/lib/uikit/math"

func BodyHeight(total, header, player, cmd int) int {
	body := total - header - player - cmd
	if body < 6 {
		body = 6
	}
	return body
}

func PaneWidths(total int) (navW, listW, inspectW int) {
	navW = math.Clamp(int(float64(total)*0.22), 18, 28)
	inspectW = math.Clamp(int(float64(total)*0.30), 24, 40)
	listW = total - navW - inspectW - 4

	if listW < 30 {
		inspectW = 0
		listW = total - navW - 2
		if listW < 20 {
			navW = 0
			listW = total
		}
	}
	return navW, listW, inspectW
}
