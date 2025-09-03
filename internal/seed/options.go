package seed

type Options struct {
	Mode Mode   // Seed mode - UP or DOWN
	Name string // Name of seed file before ".up.sql" or ".down.sql"
	File string // Full path to seed file
}

type Mode string // Seed mode - UP or DOWN

const (
	ModeUp   Mode = "up"
	ModeDown Mode = "down"
)
