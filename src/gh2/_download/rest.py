from dataclasses import dataclass
from datetime import datetime
from typing import Sequence

from mashumaro.mixins.orjson import DataClassORJSONMixin


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
