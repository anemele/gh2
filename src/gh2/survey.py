from typing import Sequence, Set

from survey import routines

from ._download import Asset, Release
from .parser import Repo


def survey_releases(repo: Repo, releases: Sequence[Release]) -> Sequence[Asset]:
    # why pyright infers this as None? its real type is Set[int].
    indexes: Set[int] = routines.basket(  # type: ignore
        f"select releases of {repo}: ",
        options=map(str, releases),
    )
    selected_releases = (releases[index] for index in indexes)

    selected_assets = []
    for release in selected_releases:
        assets = release.assets
        indexes = routines.basket(  # type: ignore
            f"select asset ({len(assets)}): ",
            options=map(str, assets),
        )
        selected_assets.extend(assets[index] for index in indexes)

    return selected_assets


def survey_cache(repos: Sequence[str]) -> Sequence[str]:
    indexes: Set[int] = routines.basket(  # type: ignore
        "select repos: ",
        options=repos,
    )
    selected_repos = [repos[index] for index in indexes]

    return selected_repos
