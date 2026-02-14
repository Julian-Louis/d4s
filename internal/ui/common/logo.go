package common

import (
	"fmt"

	"github.com/jr-k/d4s/internal/ui/styles"
)

func GetLogo() []string {
	p := styles.TagAccentLight     // primary (light orange)
	s := styles.TagAccent // shadow (darker orange)
	return []string{
		fmt.Sprintf(" [%s]██████[%s]╗[%s]   ██[%s]╗[%s]  ██[%s]╗[%s]   █████[%s]╗[%s] ", p, s, p, s, p, s, p, s, p),
		fmt.Sprintf(" [%s]██[%s]╔══[%s]██[%s]╗  [%s]██[%s]║  [%s]██[%s]║  [%s]██[%s]╔═══╝ ", p, s, p, s, p, s, p, s, p, s),
		fmt.Sprintf(" [%s]██[%s]║  [%s]██[%s]║  [%s]███████[%s]║  [%s]█████[%s]╗ ", p, s, p, s, p, s, p, s),
		fmt.Sprintf(" [%s]██[%s]║  [%s]██[%s]║       [%s]██[%s]║       [%s]██[%s]╗", p, s, p, s, p, s, p, s),
		fmt.Sprintf(" [%s]██████[%s]╔╝       [%s]██[%s]║  [%s]██████[%s]╔╝", p, s, p, s, p, s),
		fmt.Sprintf(" [%s]╚═════╝        ╚═╝  ╚═════╝ ", s),
	}
}
