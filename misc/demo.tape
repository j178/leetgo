# Usage: brew install vhs && vhs misc/demo.vhs

Set Theme "Monokai Vivid"
Set FontSize 20
Set FontFamily "FiraCode Nerd Font Mono"
Set LineHeight 1.2
Set Width 1600
Set Height 800
Set Padding 12
Set Shell fish
Set TypingSpeed 250ms
Set PlaybackSpeed 0.6

Output misc/demo.gif

Require leetgo
Require vim
Require tree

Hide
Type "cd /tmp && rm -rf my-leetcode-solutions && clear" Enter
Show

Type "# First, let's create a new leetgo workspace" Enter
Sleep 2
Type "mkdir my-leetcode-solutions" Enter
Sleep 2
Type "cd my-leetcode-solutions" Enter
Sleep 2
Type "leetgo init -t cn # Init with Chinese template" Enter
Sleep 2
Type "# A new leetgo.yaml file is created" Enter
Sleep 2
Type "tree" Sleep 1 Enter
Sleep 5

Type "# Now, let's pick a question to solve" Enter
Sleep 2
Type "leetgo pick 1" Enter
Sleep 5
Type "# Let's see what we got" Enter
Sleep 2
Type "tree" Enter
Sleep 5

Type "# Test it without changing any code" Enter
Sleep 2
Type "leetgo test last -L"
Sleep 2
Enter 2
Sleep 5

Type "# Of course, it failed :(" Enter
Sleep 3

Type "# Let's fix it" Enter
Sleep 2
Type "vim go/0001.two-sum/solution.go"
Sleep 2
Enter

Type "17Gi"
Sleep 1
Tab Type@0.1 "m := make(map[int]int)" Enter
Tab Type@0.1 "for i, num := range nums {" Enter
Tab 2 Type@0.1 "if j, ok := m[target-num]; ok {" Enter
Tab 3 Type@0.1 "return []int{j, i}" Enter
Tab 2 Type@0.1 "}" Enter
Tab 2 Type@0.1 "m[num] = i" Enter
Tab Type@0.1 "}"
Escape
Sleep 2
Type "G"
Sleep 3
Type ":wq" Enter
Sleep 3

Type "# Test and submit again" Enter
Sleep 2
Type "leetgo test last -L -s" Enter
Sleep 3
Type "# It's accepted!" Enter
Sleep 3

Enter 2
Type "# Some other powerful commands you may want to know:" Enter
Sleep 2
Type "# `leetgo contest` to watch and generate contest questions" Enter
Sleep 2
Type "# `leetgo fix` use ChatGPT to fix your code" Enter
Sleep 5
