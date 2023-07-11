package goja

import (
	"fmt"
	"time"

	"github.com/dop251/goja"
)

type Printer interface {
	Log(string)
	Warn(string)
	Error(string)
}

// Console is a modified copy of github.com/dop251/goja_nodejs/console.go but where
type Console struct {
	runtime *goja.Runtime
	util    *goja.Object
	printer Printer
}

type substreamsPrinter struct {
}

// Error implements console.Printer.
func (p *substreamsPrinter) Error(in string) {
	p.print("ERRO", in)
}

// Log implements console.Printer.
func (p *substreamsPrinter) Log(in string) {
	p.print("INFO", in)
}

// Warn implements console.Printer.
func (p *substreamsPrinter) Warn(in string) {
	p.print("WARN", in)
}

func (*substreamsPrinter) print(level string, in string) {
	fmt.Println(time.Now().Format("2006-01-02T15:04:05.000Z07:00"), level, in)
}
func (c *Console) log(p func(string)) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if format, ok := goja.AssertFunction(c.util.Get("format")); ok {
			ret, err := format(c.util, call.Arguments...)
			if err != nil {
				panic(err)
			}

			p(ret.String())
		} else {
			panic(c.runtime.NewTypeError("util.format is not a function"))
		}

		return nil
	}
}

func requireConsole(runtime *goja.Runtime, util *goja.Object, module *goja.Object) {
	c := &Console{
		runtime: runtime,
		printer: &substreamsPrinter{},
		util:    util,
	}

	o := module.Get("exports").(*goja.Object)
	o.Set("log", c.log(c.printer.Log))
	o.Set("error", c.log(c.printer.Error))
	o.Set("warn", c.log(c.printer.Warn))
}
