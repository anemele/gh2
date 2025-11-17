"""github clone

output_dir
mirror
"""

import subprocess
from pathlib import Path
from typing import Sequence

from .config import CloneConfig, load_config, load_repos


def clone(urls: Sequence[str], config: CloneConfig) -> None:
    repos = load_repos(urls)

    for repo in repos:
        repo_url = f"{config.mirror_url}{repo}.git"
        local_dst = Path(config.output_dir) / str(repo)

        cmd = ["git", "clone", repo_url, str(local_dst), *config.git_config]
        print(f"Running: {' '.join(cmd)}")

        subprocess.run(cmd)


def main():
    import argparse

    parser = argparse.ArgumentParser(
        description=__doc__, formatter_class=argparse.RawTextHelpFormatter
    )
    parser.add_argument("url", nargs="*", help="GitHub repository URL")

    args = parser.parse_args()

    config = load_config().clone

    try:
        clone(args.url, config)
    except Exception as e:
        print(f"Error: {e}")


if __name__ == "__main__":
    main()
