#!/bin/bash

find src -type d -name .git -exec sh -c "cd \"{}\"/../ && pwd && git pull" \;