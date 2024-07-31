Output demo.gif
Set TypingSpeed 0.125
Set FontSize 30

Type@10ms "# qq is a jack of all configuration formats and a master of none. But it's still pretty good. And has a lot unique features."
Sleep 1
Enter
Type@10ms "# let's start with a simple example."
Sleep 1
Enter
Type 'clear'
Enter
Sleep .25
Type 'curl -Ls https://lobste.rs | qq -i html -I'
Enter
Sleep 1
Type 'html.body.div.ol.li[].a."@href" | split("/")[3]'
Sleep 1
Enter
Sleep 1
Type@10ms "# Let's perform a demo with a Terraform module."
Enter
Type "qq '.module' tests/test.tf -I"
Enter
Sleep 1
Tab
Sleep .5
Tab
Sleep 1
Enter
Enter
Type@10ms "# You can also output the results between included formats."
Enter
Sleep 1
Type "qq '.module' tests/test.tf -o toml"
Enter
Sleep 1.5
