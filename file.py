import os

IGNORE_FILES = [
    "file.py",
    ".gitignore",
    "modules.json",
    "go.mod",
    "go.sum",
    "LICENSE",
    "README.md",
    "design.md"
]

IGNORE_FOLDERS = [
    "doc",
    ".git"
]

OUTPUT_FILE = "output.md"


def should_ignore_file(filename: str) -> bool:
    return filename in IGNORE_FILES


def should_ignore_folder(foldername: str) -> bool:
    return foldername in IGNORE_FOLDERS


def collect_files(root: str):
    for dirpath, dirnames, filenames in os.walk(root):
        # filter folders in-place
        dirnames[:] = [d for d in dirnames if not should_ignore_folder(d)]

        for fname in filenames:
            if should_ignore_file(fname):
                continue
            full_path = os.path.join(dirpath, fname)
            yield full_path


def to_relative(root: str, path: str) -> str:
    return os.path.relpath(path, root)


def main():
    root = os.getcwd()
    sections = []

    for file_path in collect_files(root):
        rel_path = to_relative(root, file_path)

        try:
            with open(file_path, "r", encoding="utf-8", errors="ignore") as f:
                content = f.read()
        except Exception as e:
            content = f"Error reading file: {e}"

        section = f"# {rel_path}\n\n```go\n{content}\n```\n"
        sections.append(section)

    output = "\n".join(sections)

    with open(OUTPUT_FILE, "w", encoding="utf-8") as f:
        f.write(output)

    print(f"Done -> {OUTPUT_FILE}")


if __name__ == "__main__":
    main()