// These tests assume there is a user sqlboiler_driver_user and a database
// by the name of sqlboiler_driver_test that it has full R/W rights to.
// In order to create this you can use the following steps from a root
// mssql account:
//
//   create database sqlboiler_driver_test;
//   go
//   use sqlboiler_driver_test;
//   go
//   create user sqlboiler_driver_user with password = 'sqlboiler';
//   go
//   exec sp_configure 'contained database authentication', 1;
//   go
//   reconfigure
//   go
//   alter database sqlboiler_driver_test set containment = partial;
//   go
//   create user sqlboiler_driver_user with password = 'Sqlboiler@1234';
//   go
//   grant alter, control to sqlboiler_driver_user;
//   go

package driver

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"os/exec"
	"regexp"
	"testing"

	"github.com/volatiletech/sqlboiler/drivers"
)

var (
	flagGolden = flag.Bool("test.golden", false, "Overwrite the golden file with the current execution results")

	rgxKeyIDs = regexp.MustCompile(`__[A-F0-9]+$`)
)

func TestDriver(t *testing.T) {
	out := &bytes.Buffer{}
	createDB := exec.Command("sqlcmd", "-S", "localhost", "-U", "sqlboiler_driver_user", "-P", "Sqlboiler@1234", "-d", "sqlboiler_driver_test", "-b", "-i", "testdatabase.sql")
	createDB.Stdout = out
	createDB.Stderr = out

	if err := createDB.Run(); err != nil {
		t.Logf("mssql output:\n%s\n", out.Bytes())
		t.Fatal(err)
	}
	t.Logf("mssql output:\n%s\n", out.Bytes())

	config := drivers.Config{
		"user":    "sqlboiler_driver_user",
		"pass":    "Sqlboiler@1234",
		"dbname":  "sqlboiler_driver_test",
		"host":    "localhost",
		"port":    1433,
		"sslmode": "disable",
		"schema":  "dbo",
	}

	p := &MSSQLDriver{}
	info, err := p.Assemble(config)
	if err != nil {
		t.Fatal(err)
	}

	for _, t := range info.Tables {
		t.PKey.Name = rgxKeyIDs.ReplaceAllString(t.PKey.Name, "")
		for i := range t.FKeys {
			t.FKeys[i].Name = rgxKeyIDs.ReplaceAllString(t.FKeys[i].Name, "")
		}
	}

	got, err := json.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}

	if *flagGolden {
		if err = ioutil.WriteFile("mssql.golden.json", got, 0664); err != nil {
			t.Fatal(err)
		}
		t.Log("wrote:", string(got))
		return
	}

	want, err := ioutil.ReadFile("mssql.golden.json")
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(want, got) != 0 {
		t.Errorf("want:\n%s\ngot:\n%s\n", want, got)
	}
}
