import logging
from dataclasses import dataclass, field
from pathlib import Path
from typing import Sequence, Set

from mashumaro.mixins.toml import DataClassTOMLMixin
from survey import routines

from .parser import Repo, parse_url

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


CONFIG_FILE_PATH = Path("gh2.toml")


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


def load_repos(urls: Sequence[str]) -> Sequence[Repo]:
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

    if len(urls) != 0:
        repos = [repo for repo in map(parse_url, urls) if repo is not None]
        if len(repos) > 0:
            _update_cache(repos)
        return repos

    cached_repos = _load_cache()
    if len(cached_repos) == 0:
        logger.debug("no url input and no cached repo found")
        return []

    logger.debug("no url input, use cached repo")
    urls = survey_cache(cached_repos)
    repos = [repo for repo in map(parse_url, urls) if repo is not None]

    return repos


REPO_CACHE_FILE_PATH = Path("gh-repos")


# repo cache file format:
#    owner1/name1
#    owner2/name2
#    ...
def _load_cache() -> Sequence[str]:
    cache_path = REPO_CACHE_FILE_PATH
    if not cache_path.exists():
        logger.debug("no url input and no cached repo found")
        return []
    cached_repos = cache_path.read_text().strip().splitlines()
    return cached_repos


def _update_cache(repos: Sequence[Repo]):
    new_repos = map(str, repos)
    new_repos = set(new_repos).union(_load_cache())
    new_repos = sorted(new_repos, key=str.lower)
    REPO_CACHE_FILE_PATH.write_text("\n".join(new_repos))
