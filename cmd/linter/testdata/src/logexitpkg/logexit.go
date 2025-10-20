package logexitpkg

import (
	logpkg "log"
	ospkg "os"
)

func useLogFatal() {
	logpkg.Fatal("stop") // want "log.Fatal допустим только в функции main пакета main"
}

func useOsExit() {
	ospkg.Exit(1) // want "os.Exit допустим только в функции main пакета main"
}
