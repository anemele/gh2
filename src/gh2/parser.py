import re
from dataclasses import dataclass
from typing import Optional


@dataclass
class Repo:
    owner: str
    name: str

    def __str__(self):
        return f"{self.owner}/{self.name}"

    @property
    def releases_url(self):
        return f"https://api.github.com/repos/{self}/releases"


PATTERN1 = re.compile(r"^(?:https://[\w\.\-]+/)|(?:git@[\w\.\-]+:)")
PATTERN2 = re.compile(r"^([\w-]+)/([\w\.-]+)")


def parse_url(url: str) -> Optional[Repo]:
    s = PATTERN1.search(url)
    if s is not None:
        url = url[s.end() :]

    url = url.removeprefix("/")

    s = PATTERN2.search(url)
    if s is None:
        return None

    owner = s.group(1)
    name = s.group(2)
    grp0 = s.group(0)
    if len(grp0) == len(url) or url[len(grp0)] != "/":
        name = name.removesuffix(".git")

    return Repo(owner, name)
