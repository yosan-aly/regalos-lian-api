package domain

// Category agrupa los regalos tal como se muestran en la página.
type Category struct {
	ID          string
	Title       string
	Description string
	Gifts       []Gift
}

// Gift es un artículo de la lista. Selected nunca supera a Required
// (lo garantiza un CHECK en la base de datos).
type Gift struct {
	ID          string
	CategoryID  string
	Name        string
	Variant     string
	Description string
	Required    int
	Selected    int
}

// Available es la cantidad que todavía se puede elegir.
func (g Gift) Available() int {
	return max(g.Required-g.Selected, 0)
}
