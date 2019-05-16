# PoC for remote cluster diagnostics

This is a proof of concept for diagnosing nodes in a cluster. It establishes an
SSH connection to a jump host, copies itself (i.e. a version of itself compiled
for the target platform and OS) to the jump host and from there does the same in
that it copies itself to the target nodes and runs the diagnosis on them,
gathering the results on stdout. After that it removes itself from all hosts.

# Running

Compile the tool for your machine's arch and for the target nodes' arch:
```sh
$ go build && GOOS=linux GOARCH=amd64 go build -o cluster-diagnostics.linux-amd64
```

Then start it up:
```sh
$ ./cluster-diagnostics makkes 1.2.3.4:22 centos 10.0.0.1:22 10.0.0.2:22 10.0.0.3:22
```

This uses the host `1.2.3.4` (and user `makkes`) as jump host and then diagnoses
the other hosts by logging into them using SSH with the user `centos`.

# Assumptions

The tool assumes that you have an SSH agent process running on your machine that
has the keys available to log into all the given hosts.
