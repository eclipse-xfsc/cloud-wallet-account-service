import os
import pathlib

TEMPLATE_NAME = "microservice-template"
EXCLUDE = [".git", ".idea", ".DS_Store", "go.sum"] + [os.path.basename(__file__)]

current = os.getcwd()
project_name = current.split("/")[-1]

print("Scanning files...")

path = pathlib.Path(current)
for file in path.rglob("*"):
    if not file.is_file():
        continue
    filename = str(file)
    if all(item not in filename for item in EXCLUDE):
        with open(filename, "r") as f:
            lines = f.readlines()
            for i, line in enumerate(lines):
                line = line.replace(TEMPLATE_NAME, project_name)
                lines[i] = line

        with open(filename, "w") as f:
            f.writelines(lines)

print(f"{TEMPLATE_NAME} references replaced with {project_name}")
