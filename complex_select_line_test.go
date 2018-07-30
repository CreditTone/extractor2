package extractor2

import (
	"testing"
)

func TestNewComplexSelectLine(t *testing.T) {
	NewComplexSelectLine("string(g_page_config =(\\s*\\{.*);\\s*g_srp_load) 	json(mods.shoplist.data.shopItems.(title='xxx').[0].isTmall) > boolean >   string")
}
