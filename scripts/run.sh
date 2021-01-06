#!/bin/bash

PROCESS=Finicky
num_processes=$(ps aux | grep -v grep | grep -ci $PROCESS)

if [ $num_processes -gt 0 ]; then
  echo "Quitting already running Finicky"
  osascript -e "quit app \"$PROCESS\"";
fi

echo "Opening built Finicky"
open ./Finicky/build/Release/Finicky.app
