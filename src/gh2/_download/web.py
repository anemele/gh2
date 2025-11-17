import asyncio
import subprocess
from pathlib import Path
from typing import Callable, Sequence, Tuple

import aiofiles
import aiohttp
import orjson
from aiohttp import ClientSession as Session
from fake_useragent import FakeUserAgent

from ..config import DownloadConfig
from ..parser import Repo
from ..survey import survey_releases
from .rest import Asset, Release


def download_repos(
    repos: Sequence[Repo],
    config: DownloadConfig,
):
    asyncio.run(_download_repos(repos, config))


async def _download_repos(
    repos: Sequence[Repo],
    config: DownloadConfig,
):
    async with aiohttp.ClientSession(
        headers={"User-Agent": str(FakeUserAgent().random)},
    ) as client:
        tasks = (get_releases(client, repo) for repo in repos)
        results = asyncio.as_completed(tasks)

        download_tasks = []
        output_dir = Path(config.output_dir)
        proxy = await get_proxy(client, config.mirrors)
        for result in results:
            repo, releases = await result
            assets = survey_releases(repo, releases)

            download_tasks.extend(
                download_asset(client, asset, output_dir, proxy) for asset in assets
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


type Proxy = Callable[[str], str]


# 代理链接一般是添加前缀，有的保留 github.com 部分，有的不保留，
# 这部分留给用户自定义，这里统一去除 https://github.com 前缀，
# 如果需要保留，则用户自己添加。
# 例如代理地址 https://a.b/https://github.com/
async def get_proxy(
    client: Session,
    mirrors: Sequence[str],
) -> Proxy | None:
    def gen_proxy(mirror: str):
        # 不带结尾的 /
        def inner(url: str) -> str:
            url = url.removeprefix("https://github.com/")
            return f"{mirror}/{url}"

        return inner

    # 任意 asset 的下载链接
    test_url = "https://github.com/cli/cli/releases/download/v2.50.0/gh_2.50.0_windows_arm64.zip"

    async def test_proxy(client: Session, proxy: Proxy):
        async with client.head(proxy(test_url)) as resp:
            resp.raise_for_status()
        return proxy

    tasks = []
    async with asyncio.TaskGroup() as tg:
        for mirror in mirrors:
            proxy = gen_proxy(mirror.removesuffix("/"))
            tasks.append(tg.create_task(test_proxy(client, proxy)))

        done, pending = await asyncio.wait(tasks, return_when=asyncio.FIRST_COMPLETED)

        for p in pending:
            p.cancel()

        for d in done:
            return await d


async def download_asset(
    client: Session,
    asset: Asset,
    output_dir: Path,
    proxy: Proxy | None,
):
    url = asset.browser_download_url
    if proxy is not None:
        url = proxy(url)
    path = output_dir.joinpath(asset.name)
    async with client.get(url) as resp:
        content = await resp.read()
    async with aiofiles.open(path, "wb") as fp:
        await fp.write(content)
