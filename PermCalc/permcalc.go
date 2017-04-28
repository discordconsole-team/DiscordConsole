package permcalc

import "errors"
import "github.com/nsf/termbox-go"

// ErrAlreadyRunning is thrown when you attempt to show
// a prompt when there already is one displayed.
var ErrAlreadyRunning = errors.New("permcalc prompt already running")

// PromptPerm shows the permission calculator
// and returns whatever the user checked.
func PromptPerm() (int, error) {
	pm := PermCalc{}

	err := pm.Show()
	if err != nil {
		return 0, err
	}

	return pm.Perm, nil
}

// PermCalc is a permission calculator object.
// It stores data and similar
type PermCalc struct {
	Perm     int
	ReadOnly bool
}

// Show shows the permission calculator
// according to the PermCalc object.
func (pm *PermCalc) Show() error {
	if d.running {
		return ErrAlreadyRunning
	}

	d = data{
		running:  true,
		x:        optionX1 + 1,
		y:        optionY1,
		perm:     pm.Perm,
		readonly: pm.ReadOnly,
	}

	err := termbox.Init()
	if err != nil {
		return err
	}
	defer termbox.Close()

	for d.running {
		err := drawScreen()
		if err != nil {
			return err
		}

		event := termbox.PollEvent()
		handleKey(event.Key, event.Ch)
	}

	pm.Perm = d.perm
	return nil
}
