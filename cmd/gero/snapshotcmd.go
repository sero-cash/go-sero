package main

import (
	"fmt"
	"github.com/sero-cash/go-sero/cmd/utils"
	"github.com/sero-cash/go-sero/zero/snapshot"
	"gopkg.in/urfave/cli.v1"
)

var (
	snapshotCommand = cli.Command{
		Action:    utils.MigrateFlags(makeSnapshot),
		Name:      "snapshot",
		Usage:     "Create snapshot from a target chaindata folder",
		ArgsUsage: "<targetChaindataDir>",
		Flags: []cli.Flag{
			utils.DataDirFlag,
		},
		Category: "BLOCKCHAIN COMMANDS",
		Description: `
The first argument must be the directory containing the blockchain to download from`,
	}
)

func makeSnapshot(ctx *cli.Context) error {

	if len(ctx.Args()) != 1 {
		utils.Fatalf("target chaindata directory path argument missing")
	}

	stack, _:= makeConfigNode(ctx)

	srcDb:=stack.ResolvePath("chaindata")
	tarDb:=ctx.Args().First()

	fmt.Printf("%s ----> %s\n",srcDb,tarDb)

	if sg,err:=snapshot.NewSnapshotGen(srcDb,tarDb);err!=nil {
		println("Make snapshot error:",err.Error())
	} else {
		sg.Run()
		sg.Close()
	}

	return nil
}
