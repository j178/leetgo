Set Theme { "black": "#1d202f", "red": "#f7768e", "green": "#9ece6a", "yellow": "#e0af68", "blue": "#7aa2f7", "purple": "#bb9af7", "cyan": "#7dcfff", "white": "#a9b1d6", "brightBlack": "#414868", "brightRed": "#f7768e", "brightGreen": "#a6e3a1", "brightYellow": "#e0af68", "brightBlue": "#7aa2f7", "brightPurple": "#bb9af7", "brightCyan": "#7dcfff", "brightWhite": "#c0caf5", "background": "#24283b", "foreground": "#c0caf5", "selectionBackground": "#364a82", "cursorColor": "#c0caf5" }

Set FontSize 20
Set FontFamily "FiraCode Nerd Font Mono"
Set Width 1600
Set Height 800
Set Framerate 30
Set Padding 12
Set Shell fish
Set TypingSpeed 200ms

Output misc/demo.gif
Output misc/demo.ascii

Require leetgo
Require vim
Require tree

Type "# initialize a new leetgo project" Enter
Type "mkdir my-leetcode-solutions" Enter
Type "cd my-leetcode-solutions" Enter
Type "leetgo init -t cn" Enter
Hide
Sleep 15s
Show
Sleep 500ms
Type "tree"
Enter

Type "# Change config to suit your needs" Enter
Type "vim leetgo.yaml" Enter
Sleep 500ms
Type "jjjjjjjjjxxjxxllllllldwavim"
Escape
Type ":wq" Enter

Type "# Generate a question" Enter
Type "leetgo pick 1" Enter
Hide
Sleep 5s
Show
Sleep 500ms
Type "tree" Enter
