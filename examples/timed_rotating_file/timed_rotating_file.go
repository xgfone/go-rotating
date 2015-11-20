package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/xgfone/go-rotating"
	"time"
)

var log = logrus.New()

func init() {
	log.Out = rotating.NewNullWriter()
	hook, err := rotating.NewTimedRotatingFileHook("test.log")
	if err != nil {
		fmt.Println(err)
	} else {
		hook.SetDebug(true).SetBackupCount(7)
		log.Hooks.Add(hook)
	}
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
