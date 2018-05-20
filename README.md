# rails-log-parser
A parser for Rails logs written in Go

I found the logs for a long running application (I didn't apply logrotation back then) and I thought it could be useful to parse them and try to infer useful information with it.

My idea is to parse them and publish to an ES instance.

It is intended as an exercise in go, but let's see where does it get!
