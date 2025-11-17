from dataclasses import dataclass, field
from pathlib import Path
from typing import Sequence

from mashumaro.mixins.toml import DataClassTOMLMixin

from .parser import Repo, parse_url
from .survey import survey_cache

CONFIG_FILE_PATH = Path("gh2.toml")


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
    config_file_path = CONFIG_FILE_PATH

    if not config_file_path.exists():
        return Config()
    content = config_file_path.read_text(encoding="utf-8")
    return Config.from_toml(content)


RepoCacheFileName = Path("gh-repos")


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
        if not RepoCacheFileName.exists():
            return []
        cached_repos = RepoCacheFileName.read_text().strip().splitlines()
        urls = survey_cache(cached_repos)

    repos = [repo for repo in map(parse_url, urls) if repo is not None]
    return repos


def update_repos(repos: Sequence[Repo]):
    cached_repos = load_repos([])
    new_repos = set(cached_repos).union(repos)
    new_repos = map(str, new_repos)
    new_repos = sorted(new_repos, key=str.lower)
    RepoCacheFileName.write_text("\n".join(new_repos))
