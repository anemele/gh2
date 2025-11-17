"""github download

mirror
"""

from pathlib import Path
from typing import Sequence

from ._download import download_repos
from .config import DownloadConfig, load_config, load_repos


def download(urls: Sequence[str], config: DownloadConfig):
    output_dir = Path(config.output_dir)
    output_dir.mkdir(exist_ok=True)

    repos = load_repos(urls)
    if len(repos) == 0:
        return

    download_repos(repos, config)


def main():
    import argparse

    parser = argparse.ArgumentParser(
        description=__doc__, formatter_class=argparse.RawTextHelpFormatter
    )

    parser.add_argument("url", nargs="*", help="GitHub repository URL")

    args = parser.parse_args()

    config = load_config().download
    download(args.url, config)


if __name__ == "__main__":
    main()
