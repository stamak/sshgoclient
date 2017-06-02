# It is a simple ssh client written in Go (Golang)

Advantage over usual ssh client - it runs command on list of hosts
simultaneously (thanks Go build-in concurrency).


Usage: `./sshgoclient -cmd 'echo test; ping -c1 localhost' -hosts 'host1 host2'`
