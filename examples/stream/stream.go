package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/xgfone/go-rotating"
	"os"
	"time"
)

var log = logrus.New()

func init() {
	log.Out = rotating.NewNullWriter()
	hook := rotating.NewStreamHook(os.Stderr)
	log.Hooks.Add(hook)
}

func Loop() {
	log.WithFields(logrus.Fields{
		"test1": "test1",
		"test2": "test2",
	}).Info("This is a test.")
}

func main() {
	for true {
		time.Sleep(time.Second)
		Loop()
	}
}
