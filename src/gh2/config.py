import logging
from dataclasses import dataclass, field
from pathlib import Path
from typing import Sequence, Set

from mashumaro.mixins.toml import DataClassTOMLMixin
from survey import routines

from .parser import Repo, parse_url

CONFIG_FILE_PATH = Path("gh2.toml")

logger = logging.getLogger(__name__)


@dataclass
class CloneConfig:
    output_dir: str = field(default=".")
    mirror_url: str = field(default="https://github.com/")
    git_config: Sequence[str] = field(default_factory=list)


@dataclass
class DownloadConfig:
    output_dir: str = field(default=".")
    mirrors: Sequence[str] = field(default_factory=list)


@dataclass
class Config(DataClassTOMLMixin):
    clone: CloneConfig = field(default_factory=CloneConfig)
    download: DownloadConfig = field(default_factory=DownloadConfig)


def load_config() -> Config:
    cfg_path = CONFIG_FILE_PATH

    if not cfg_path.exists():
        logger.info(f"not found {cfg_path}, use default config")
        return Config()
    content = cfg_path.read_text(encoding="utf-8")
    logger.info(f"use config from {cfg_path}")
    config = Config.from_toml(content)
    logger.debug(f"{config=}")
    return config


def survey_cache(repos: Sequence[str]) -> Sequence[str]:
    indexes: Set[int] = routines.basket(  # type: ignore
        "select repos: ",
        options=repos,
    )
    selected_repos = [repos[index] for index in indexes]

    return selected_repos


REPO_CACHE_FILE_PATH = Path("gh-repos")


def load_repos(urls: Sequence[str]) -> Sequence[Repo]:
    """
    repo cache file format:
    owner1/name1
    owner2/name2
    ...
    """
    # 输入 url 为空，怎么处理？
    # 加载缓存
    #
    # 此处先跳过，因为加载缓存需要 TUI 程序
    # Python 暂时不知有哪些合适的库
    #
    # 感谢大佬！有一个 survey 库
    # https://github.com/Exahilosys/survey
    # 正好是参考了 golang 的 survey 库
    # https://github.com/AlecAivazis/survey
    # 参考文档 https://survey.readthedocs.io/reference.html

    if len(urls) == 0:
        if not REPO_CACHE_FILE_PATH.exists():
            logger.debug("no url input and no cached repo found")
            return []
        cached_repos = REPO_CACHE_FILE_PATH.read_text().strip().splitlines()
        urls = survey_cache(cached_repos)
        logger.debug("no url input, use cached repo")

    repos = [repo for repo in map(parse_url, urls) if repo is not None]
    return repos


def update_repos(repos: Sequence[Repo]):
    cached_repos = load_repos([])
    new_repos = set(cached_repos).union(repos)
    new_repos = map(str, new_repos)
    new_repos = sorted(new_repos, key=str.lower)
    REPO_CACHE_FILE_PATH.write_text("\n".join(new_repos))
