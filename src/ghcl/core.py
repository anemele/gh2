import subprocess

from .config import Config
from .parser import parse_url


def clone(url: str, config: Config) -> None:
    repo = parse_url(url)
    if repo is None:
        print(f"Invalid URL: {url}")
        return None

    repo_url = f"{config.mirror_url}{repo}.git"
    local_dir = repo.name if config.no_owner else str(repo)
    local_dst = config.destiny / local_dir

    cmd = f"git clone {repo_url} {local_dst}"
    if config.git_config:
        cmd += " ".join(config.git_config)

    res = subprocess.run(cmd)
    if res.returncode != 0:
        print(f"Failed to clone {url}")
        return None
