import asyncio
import logging
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
from .rest import Asset, Release, survey_releases

logger = logging.getLogger(__name__)


def download_repos(
    repos: Sequence[Repo],
    config: DownloadConfig,
):
    logger.debug("into asyncio runtime")
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

        if proxy is None:
            logger.info("no proxy")
        else:
            logger.info(f"proxy={proxy('')}")

        for result in results:
            repo, releases = await result
            logger.info(f"{repo} has {len(releases)} releases")
            assets = survey_releases(str(repo), releases)
            logger.info(f"select {len(assets)} assets")

            download_tasks.extend(
                download_asset(client, asset, output_dir, proxy) for asset in assets
            )

        await asyncio.gather(*download_tasks)


async def get_releases(
    client: Session,
    repo: Repo,
) -> Tuple[Repo, Sequence[Release]]:
    async with client.get(repo.releases_url) as resp:
        if resp.ok:
            content = await resp.read()
        else:
            logger.info("request failed, use gh cli")
            cmd = ["gh", "api", f"repos/{repo}/releases"]
            content = subprocess.run(cmd, capture_output=True).stdout

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
        logger.debug(f"test proxy {proxy('')}")
        async with client.head(proxy(test_url)) as resp:
            resp.raise_for_status()
        logger.debug(f"proxy in access: {proxy('')}")
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
    # TODO: multi-parts download, progress bar.
    async with client.get(url) as resp:
        content = await resp.read()
    path = output_dir.joinpath(asset.name)
    async with aiofiles.open(path, "wb") as fp:
        await fp.write(content)
    logger.info(f"downloaded {path}")
