import logging
import subprocess
from pathlib import Path
from typing import Sequence

from ._download import download_repos
from .config import load_config, load_repos
from .logging_config import setup_logging

logger = logging.getLogger(__name__)
logger.addHandler(logging.StreamHandler())


def clone(urls: Sequence[str]) -> None:
    """github clone

    output_dir
    mirror_url
    git_config
    """

    repos = load_repos(urls)
    if len(repos) == 0:
        logger.info("no repo to clone")
        return

    config = load_config().clone

    for repo in repos:
        repo_url = f"{config.mirror_url}{repo}.git"
        local_dst = Path(config.output_dir) / str(repo)

        cmd = ["git", "clone", repo_url, str(local_dst), *config.git_config]
        msg = f"Running: {' '.join(cmd)}"
        logging.info(msg)

        subprocess.run(cmd)


def download(urls: Sequence[str]):
    """github download

    output_dir
    mirrors
    """

    repos = load_repos(urls)
    if len(repos) == 0:
        logger.info("no repo to download")
        return

    config = load_config().download
    output_dir = Path(config.output_dir)
    output_dir.mkdir(exist_ok=True)

    download_repos(repos, config)


def gen_main(fn):
    def main():
        import argparse

        parser = argparse.ArgumentParser(
            description=fn.__doc__, formatter_class=argparse.RawTextHelpFormatter
        )

        parser.add_argument("url", nargs="*", help="GitHub repository URL")
        parser.add_argument(
            "--debug", action="store_true", help="Set log level as DEBUG"
        )
        args = parser.parse_args()
        urls = args.url
        debug = args.debug

        setup_logging(debug)

        logger.debug(f"{args=}")

        try:
            fn(urls)
        except Exception as e:
            logger.error(f"Error: {e}")

    return main


main_cl = gen_main(clone)
main_dl = gen_main(download)
