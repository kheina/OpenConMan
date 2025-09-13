package test

import (
	"fmt"
	"slices"
	"time"

	"github.com/coreos/go-systemd/v22/sdjournal"
	"github.com/kheina/openconman/src/cli"
)

type Command struct {
	cli.UnimplementedCommand
}

func (c *Command) Run() error {
	j, err := sdjournal.NewJournal()
	if err != nil {
		return err
	}
	defer j.Close()
	j.AddMatch((&sdjournal.Match{Field: sdjournal.SD_JOURNAL_FIELD_SYSTEMD_UNIT, Value: "ocm-copyparty.service"}).String())
	if err = j.SeekTail(); err != nil {
		return err
	}

	entries := []*sdjournal.JournalEntry{}

	for range 100 {
		if n, err := j.Previous(); err != nil || n == 0 {
			break
		}
		entry, err := j.GetEntry()
		if err != nil {
			break
		}
		entries = append(entries, entry)
	}
	for _, e := range slices.Backward(entries) {
		fmt.Println(time.Unix(0, int64(e.RealtimeTimestamp)*int64(time.Microsecond)), e.Fields["MESSAGE"])
	}

	return nil
}

func RegisterCommand(i *cli.CLI) {
	i.NewCommand("test", func() (cli.Command, error) {
		return &Command{}, nil
	})
}
