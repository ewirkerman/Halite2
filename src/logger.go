package src

// Fohristiwhirl, November 2017.
// A sort of alternative starter kit.

import (
	"fmt"
	"os"
	"reflect"
)

type Logfile struct {
	outfile     *os.File
	outfilename string
	logged_once map[string]bool
}

func NewLog(outfilename string) *Logfile {
	return &Logfile{
		nil,
		outfilename,
		make(map[string]bool),
	}
}

func (self *Logfile) Log(format_string string, args ...interface{}) {

	if self == nil {
		return
	}

	if self.outfile == nil {

		var err error

		if _, tmp_err := os.Stat(self.outfilename); tmp_err == nil && false {
			// File exists
			self.outfile, err = os.OpenFile(self.outfilename, os.O_CREATE|os.O_WRONLY, 0666)
		} else {
			// File needs creating
			self.outfile, err = os.Create(self.outfilename)
		}

		if err != nil {
			return
		}
	}

	fmt.Fprintf(self.outfile, format_string, args...)
	fmt.Fprintf(self.outfile, "\r\n") // Because I use Windows...
}

func (self *Logfile) LogOnce(format_string string, args ...interface{}) bool {
	if self.logged_once[format_string] == false {
		self.logged_once[format_string] = true // Note that it's format_string that is checked / saved
		self.Log(format_string, args...)
		return true
	}
	return false
}

// ---------------------------------------------------------------

func (g *Game) StartLog(logfilename string) {
	g.logfile = NewLog(logfilename)
}

func (g *Game) Log(format_string string, args ...interface{}) {
	format_string = fmt.Sprintf("%vt %3d: ", g.pid, g.Turn()) + format_string
	g.logfile.Log(format_string, args...)
}

func (g *Game) LogEach(format_string string, slice interface{}) {
	rv := reflect.ValueOf(slice)
	g.Log(format_string)
	for i := 0; i < rv.Len(); i++ {
		g.Log("%v", rv.Index(i))
	}
}

func (g *Game) LogOnce(format_string string, args ...interface{}) bool {
	format_string = "t %3d: " + format_string
	var newargs []interface{}
	newargs = append(newargs, g.Turn())
	newargs = append(newargs, args...)
	return g.logfile.LogOnce(format_string, newargs...)
}

func (g *Game) LogWithoutTurn(format_string string, args ...interface{}) {
	g.logfile.Log(format_string, args...)
}
