on run launchArguments
  set previousTabIds to {}
  set snapshotError to missing value
  if application id "com.google.Chrome" is running then
    try
      with timeout of 5 seconds
        tell application id "com.google.Chrome"
          repeat with chromeWindow in windows
            if (count tabs of chromeWindow) > 0 then
              set end of previousTabIds to id of active tab of chromeWindow
            end if
          end repeat
        end tell
      end timeout
    on error errorMessage
      set snapshotError to errorMessage
    end try
  end if

  set launchCommand to "/usr/bin/open"
  repeat with launchArgument in launchArguments
    set launchCommand to launchCommand & " " & quoted form of (launchArgument as text)
  end repeat
  do shell script launchCommand

  if snapshotError is not missing value then
    error "Link opened, but focusWindow could not inspect Chrome tabs. Allow Chrome Automation access: " & snapshotError
  end if

  set targetWindowId to missing value
  set pollIntervalSeconds to 0.05
  set focusTimeoutSeconds to 5
  set focusDeadline to (current date) + focusTimeoutSeconds
  repeat while (current date) < focusDeadline
    with timeout of 5 seconds
      tell application id "com.google.Chrome"
        repeat with chromeWindow in windows
          if (count tabs of chromeWindow) > 0 then
            if id of active tab of chromeWindow is not in previousTabIds then
              set targetWindowId to id of chromeWindow
              exit repeat
            end if
          end if
        end repeat
      end tell
    end timeout
    if targetWindowId is not missing value then exit repeat
    delay pollIntervalSeconds
  end repeat

  if targetWindowId is missing value then
    error "Link opened, but focusWindow could not identify its Chrome window"
  end if

  try
    with timeout of 5 seconds
      tell application id "com.google.Chrome"
        set index of window id targetWindowId to 1
        activate
      end tell
      tell application "System Events"
        tell process "Google Chrome"
          perform action "AXRaise" of window 1
          set frontmost to true
        end tell
      end tell
    end timeout
  on error errorMessage
    error "Link opened, but focusWindow could not raise its Chrome window. Allow System Events Automation and Accessibility access: " & errorMessage
  end try
end run
