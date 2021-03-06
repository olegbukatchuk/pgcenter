// Stuff related to Postgres config files such as displaying, editing, etc.

package top

import (
	"bytes"
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/lesovsky/pgcenter/lib/stat"
	"github.com/lesovsky/pgcenter/lib/utils"
	"os"
	"os/exec"
	"strings"
)

// Show Postgres config in $PAGER program
func showPgConfig(g *gocui.Gui, _ *gocui.View) error {
	rows, err := conn.Query(stat.PgGetConfigAllQuery)
	if err != nil {
		printCmdline(g, err.Error())
		return nil
	}
	defer rows.Close()

	var buf bytes.Buffer
	var res stat.PGresult

	if err := res.New(rows); err != nil {
		printCmdline(g, err.Error())
		return nil
	}

	res.SetAlignCustom(512)
	fmt.Fprintf(&buf, "PostgreSQL configuration:\n")
	res.Fprint(&buf)

	var pager string
	if pager = os.Getenv("PAGER"); pager == "" {
		pager = utils.DefaultPager
	}

	// Exit from UI and stats loop... will restore it after $PAGER is closed.
	do_exit <- 1
	g.Close()

	cmd := exec.Command(pager)
	cmd.Stdin = strings.NewReader(buf.String())
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		// If external program fails, save error and show it to user in next UI iteration
		errSaved = err
	}

	return err
}

// Open specified configuration file in $EDITOR program
func editPgConfig(g *gocui.Gui, n string) error {
	if !conninfo.ConnLocal {
		printCmdline(g, "Edit config is not supported for remote hosts")
		return nil
	}

	var config_file string

	if n != stat.GucRecoveryFile {
		conn.QueryRow(stat.PgGetSingleSettingQuery, n).Scan(&config_file)
	} else {
		var data_directory string
		conn.QueryRow(stat.PgGetSingleSettingQuery, stat.GucDataDir).Scan(&data_directory)
		config_file = data_directory + "/" + n
	}

	var editor string
	if editor = os.Getenv("EDITOR"); editor == "" {
		editor = utils.DefaultEditor
	}

	// Exit from UI and stats loop... will restore it after $EDITOR is closed.
	do_exit <- 1
	g.Close()

	cmd := exec.Command(editor, config_file)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		// If external program fails, save error and show it to user in next UI iteration
		errSaved = err
		return err
	}

	return nil
}
