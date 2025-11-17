import asyncio
import subprocess
from pathlib import Path
from typing import Sequence, Tuple

import aiofiles
import aiohttp
import orjson
from aiohttp import ClientSession as Session
from fake_useragent import FakeUserAgent

from ..config import DownloadConfig
from ..parser import Repo
from ..survey import survey_releases
from .rest import Asset, Release


def download_repos(repos: Sequence[Repo], config: DownloadConfig):
    asyncio.run(_download_repos(repos, config))


async def _download_repos(repos: Sequence[Repo], config: DownloadConfig):
    async with aiohttp.ClientSession(
        headers={"User-Agent": str(FakeUserAgent().random)},
    ) as client:
        tasks = (get_releases(client, repo) for repo in repos)
        results = asyncio.as_completed(tasks)

        download_tasks = []
        output_dir = Path(config.output_dir)
        for result in results:
            repo, releases = await result
            assets = survey_releases(repo, releases)

            download_tasks.extend(
                download_asset(client, asset, output_dir, None) for asset in assets
            )

        await asyncio.gather(*download_tasks, return_exceptions=True)


def _get_releases_gh_api(repo: Repo) -> bytes:
    url = f"repos/{repo}/releases"
    cmd = ["gh", "api", url]
    data = subprocess.run(cmd, capture_output=True).stdout
    return data


async def get_releases(
    client: Session,
    repo: Repo,
) -> Tuple[Repo, Sequence[Release]]:
    async with client.get(repo.releases_url) as resp:
        if resp.status == 200:
            content = await resp.read()
        else:
            content = _get_releases_gh_api(repo)

    data = orjson.loads(content)
    ret = map(Release.from_dict, data)
    return repo, list(ret)


async def download_asset(
    client: Session,
    asset: Asset,
    output_dir: Path,
    proxy,
):
    url = asset.browser_download_url
    path = output_dir.joinpath(asset.name)
    async with client.get(url) as resp:
        content = await resp.read()
    async with aiofiles.open(path, "wb") as fp:
        await fp.write(content)
