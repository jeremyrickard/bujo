package key

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/gosuri/uitable"
	"github.com/n3wscott/bujo/pkg/glyph"
)

type Key struct{}

func (k *Key) Do(ctx context.Context) error {
	_, _ = fmt.Fprintln(color.Output, "")
	k.Key(ctx, glyph.DefaultGlyphs(), false)
	_, _ = fmt.Fprintln(color.Output, "")
	k.Key(ctx, glyph.DefaultGlyphs(), true)

	return nil
}

func (k *Key) Key(ctx context.Context, glyfs []glyph.Glyph, sig bool) {
	tbl := uitable.New()
	tbl.Separator = "  "
	if sig {
		tbl.AddRow(glyph.Bold("Signifiers"), glyph.Bold("Meaning"))
	} else {
		tbl.AddRow(glyph.Bold("Bullets"), glyph.Bold("Meaning"))
	}
	for _, v := range glyfs {
		if sig == v.Signifier {
			tbl.AddRow(v.Symbol, v.Meaning)
		}
	}
	tbl.RightAlign(0)

	_, _ = fmt.Fprintln(color.Output, tbl)
}
