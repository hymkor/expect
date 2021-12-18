local pid = assert(spawn("notepad"))
echo("wait ".. pid)
assert(wait(pid))
echo("done ".. pid)
