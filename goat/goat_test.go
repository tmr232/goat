package goat

import (
	"testing"
)

func TestApp(t *testing.T) {
	t.Run("Single action app", func(t *testing.T) {
		var target string
		action := func(args struct{ S string }) error {
			target = args.S
			return nil
		}
		app := App("test", Action(action))

		err := app.Run([]string{"test", "--S", "magic"})
		if err != nil {
			t.Errorf("App failed with error %v", err)
		}

		if target != "magic" {
			t.Errorf("Want=%v, got=%v", "magic", target)
		}
	})
	t.Run("Single command app", func(t *testing.T) {
		var target string
		command := func(args struct{ S string }) error {
			target = args.S
			return nil
		}
		app := App("test", Command("command", command))

		err := app.Run([]string{"test", "command", "--S", "magic"})
		if err != nil {
			t.Errorf("App failed with error %v", err)
		}

		if target != "magic" {
			t.Errorf("Want=%v, got=%v", "magic", target)
		}
	})
}
