from dataclasses import dataclass
from datetime import datetime
from typing import Sequence, Set

from mashumaro.mixins.orjson import DataClassORJSONMixin
from survey import routines


def human_size(size: float) -> str:
    if size < 1024:
        return f"{size} B"
    for unit in "KMG":
        size /= 1024
        if size < 1024:
            return f"{size:.2f} {unit}B"
    return f"{size:.2f} TB"


@dataclass
class Asset:
    name: str
    label: str
    content_type: str
    size: int
    created_at: datetime
    updated_at: datetime
    browser_download_url: str

    def __str__(self) -> str:
        date = self.created_at.strftime("%Y-%m-%d")
        size = human_size(self.size)
        return f"{self.name} ({date}, {size})"


@dataclass
class Release(DataClassORJSONMixin):
    tag_name: str
    name: str
    body: str
    draft: bool
    prerelease: bool
    created_at: datetime
    published_at: datetime
    assets: Sequence[Asset]

    def __str__(self) -> str:
        # 这里要使用 PublishedAt 而不是 CreatedAt，因为 release 可以更新
        # 例如 x64dbg/x64dbg 以 snapshot 发布，它的 created_at 不会变
        date = self.published_at.strftime("%Y-%m-%d")
        return f"{self.name} ({date})"


def survey_releases(repo: str, releases: Sequence[Release]) -> Sequence[Asset]:
    if len(releases) == 0:
        return []

    # why pyright infers this as None? its real type is Set[int].
    indexes: Set[int] = routines.basket(  # type: ignore
        f"select releases of {repo}: ",
        options=map(str, releases),
    )
    selected_releases = (releases[index] for index in indexes)

    selected_assets = []
    for release in selected_releases:
        assets = release.assets

        if len(assets) == 0:
            continue

        indexes = routines.basket(  # type: ignore
            f"select asset ({len(assets)}): ",
            options=map(str, assets),
        )
        selected_assets.extend(assets[index] for index in indexes)

    return selected_assets
