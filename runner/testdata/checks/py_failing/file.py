print("Hello World")

f = open("file.txt")
f.write("Hello World")

if True:
    breakpoint()  # [PTC-W0014]: 4 "Debugger activation detected"
    print("Foo bar")

# [PY-DUMMY]: "Something"
print("Baz")