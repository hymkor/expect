package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
)

func newUsage() {
	w := flag.CommandLine.Output()
	fmt.Fprintf(w, "Expect-lua %s-windows-%s with %s\n",
		version, runtime.GOARCH, runtime.Version())
	fmt.Fprintf(w, "Usage of %s:\n", os.Args[0])
	newPrintDefaults(flag.CommandLine)
}

func newPrintDefaults(fs *flag.FlagSet) {
	fs.VisitAll(func(f *flag.Flag) {
		if f.Usage == "" {
			return // for obsolute flag
		}
		var b strings.Builder

		fmt.Fprintf(&b, "  -%s", f.Name)

		var detail string
		var usage string
		var ok bool

		if detail, usage, ok = strings.Cut(f.Usage, "\v"); ok {
			fmt.Fprintf(&b, " %s", detail)
		} else {
			usage = f.Usage
		}
		if b.Len() <= 4 {
			b.WriteByte('\t')
		} else {
			b.WriteString("\n    \t")
		}
		b.WriteString(strings.ReplaceAll(usage, "\n", "\n    \t"))
		fmt.Fprintln(fs.Output(), b.String())
	})
}
