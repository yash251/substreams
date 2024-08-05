package common

// Common is a struct all components should embed.
type Common struct {
	Width  int
	Height int
}

// SetSize sets the width and height of the common struct.
func (c *Common) SetSize(width, height int) {
	c.Width = width
	c.Height = height
}
