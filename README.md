# :milkyway: Galaxy Cloner

The stars are born, grow and die. Let's keep a copy before their death.

# Why

If you're like me and star projects as a bookmark, reference, help for communication, etc, at some point, you go back in it to check how is the project going.
The more you wait, the more you will have bad suprise and some project will just disappear.
So I decided to have this tool to brutefork my way into the stars to be able to have a reference to the code, or maybe use it in a future project.

# Usage

```
export GALAXY_CLONER_GITHUB_TOKEN=XXX
# Where will all the clone go (it must be a github org)
export GALAXY_CLONER_DEST_ORG=YYY
# Be careful with that setting, it can get you rate limited.
# Without setting it, it will use runtime.NumCPU(), which can be a problem.
export GALAXY_CLONER_PARALLEL=1
go run main.go
```


# Limitations

* I want to make a daemon out of it so it can listen to new star notifications and only fork those.
* It doesn't handle name conflict. For now, the conflict are just ignored because the repo "already exists".
