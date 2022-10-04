print("Hello World")

# [PTC-W0010]: "File opened without the with statement"
f = open("file.txt")
f.write("Hello World")

if True:
    breakpoint()  # [PTC-W0014]: 2 "Debugger activation detected"
    print("Foo bar")
