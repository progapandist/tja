# Regenerate the README screenshots: go build -o /tmp/tjabin . && vhs docs/demo.tape
Output docs/.tmp.gif
Set Shell "bash"
Set FontSize 18
Set Width 1700
Set Height 780
Set Padding 24
Set Theme "Catppuccin Mocha"

Hide
Type "/tmp/tjabin" Enter
Sleep 2s
Type "/"
Sleep 300ms
Type "annehm"
Sleep 500ms
Enter
Sleep 500ms
Show
Sleep 1s
Screenshot docs/browse.png

Type "t"
Sleep 800ms
Space
Sleep 1s
Screenshot docs/test.png
Sleep 500ms
Type "q"
Sleep 500ms
