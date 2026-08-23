package cmd

import (
	"github.com/fatih/color"
)

/*
tColumn is a generic description of a single table column, parameterised by the
row type T it renders. It captures how the column is shown, titled and coloured,
where its content comes from, and how it is rendered in markdown.
*/
type tColumn[T any] struct {
	isShown         func(tConfig) bool
	title           func() string
	titleColor      color.Attribute
	contentSource   func(T, tConfig) string
	contentColor    func(T) color.Attribute
	contentAlignMD  int
	contentEscapeMD bool
}
