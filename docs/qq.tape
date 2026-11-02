# Renders docs/demo.gif. Run from the repo root with `make demo`.

Output docs/demo.gif

Require bash

Set Shell bash
Set FontSize 20
Set Width 1400
Set Height 820
Set Padding 28
Set Margin 16
Set MarginFill "#1e1e2e"
Set BorderRadius 10
Set WindowBar Colorful
Set Theme "Catppuccin Mocha"
Set TypingSpeed 40ms
Set CursorBlink false

# Work in a scratch directory so the prompt and paths stay short on screen.
Hide
Type "export REPO=$PWD PATH=$PWD/bin:$PATH PS1='\[\e[1;35m\]❯\[\e[0m\] '"
Enter
Type "cd $(mktemp -d) && cp $REPO/docs/demo/deployment.yaml . && cp $REPO/tests/test.tf main.tf && cp $REPO/tests/test.csv sensors.csv && cp $REPO/tests/test.msgpack user.msgpack && clear"
Enter
Show

Sleep 500ms

# 1. Interactive mode: the preview re-renders on every keystroke.
# Fast typing keeps transient invalid queries from flashing on screen.
Type "qq -I deployment.yaml"
Sleep 300ms
Enter
Sleep 1.5s
# The editor starts pre-filled with ".".
Type@120ms "spec"
Sleep 1.2s
Type@25ms ".template.spec.containers"
Sleep 1.5s
Type@25ms " | map({name, image})"
Sleep 2.5s
Escape
Sleep 2.5s
Type "clear"
Enter

# 2. Any format to any format.
Type "qq '.metadata' deployment.yaml -o toml"
Sleep 300ms
Enter
Sleep 2s
Type "qq '.module.vpc[0] | {name, cidr, azs}' main.tf -o yaml"
Sleep 300ms
Enter
Sleep 2.5s
Type "clear"
Enter

# 3. Full jq on tabular data.
Type "qq 'map(select(.Temperature > 23)) | map({Location, Temperature})' sensors.csv -o csv"
Sleep 300ms
Enter
Sleep 2.5s

# 4. Binary formats become readable.
Type "cat user.msgpack | qq -i msgpack -o yaml"
Sleep 300ms
Enter
Sleep 2.5s
Type "clear"
Enter

# 5. Compose with unix tools via gron.
Type "qq deployment.yaml -o gron | grep image | qq -i gron -o yaml"
Sleep 300ms
Enter
Sleep 4s
