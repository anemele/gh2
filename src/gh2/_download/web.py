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
from rich.progress import Progress, TaskID
from tenacity import (
    retry,
    retry_if_exception_type,
    stop_after_attempt,
    wait_exponential,
)

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


MAX_CONCURRENCY = 4
MAX_CHUNK_SIZE = 10 << 20  # 10 MB
MIN_CHUNK_SIZE = 1 << 17  # 128 KB
FIRST_CHUNK_SIZE = 1 << 20
FIRST_SPEED = FIRST_CHUNK_SIZE


async def _download_repos(
    repos: Sequence[Repo],
    config: DownloadConfig,
):
    async with aiohttp.ClientSession(
        headers={"User-Agent": str(FakeUserAgent().random)},
    ) as client:
        tasks = (_get_releases(client, repo) for repo in repos)
        results = asyncio.as_completed(tasks)

        output_dir = Path(config.output_dir)
        proxy = await _get_proxy(client, config.mirrors)

        if proxy is None:
            logger.info("no proxy")
        else:
            logger.info(f"proxy={proxy('')}")

        assets = []
        for result in results:
            repo, releases = await result
            logger.info(f"{repo} has {len(releases)} releases")
            assets2 = survey_releases(str(repo), releases)
            logger.info(f"select {len(assets2)} assets")
            assets.extend(assets2)

        with Progress(transient=True) as progress:
            semaphore = asyncio.Semaphore(MAX_CONCURRENCY)
            download_tasks = (
                _download_asset(
                    client,
                    asset,
                    output_dir,
                    proxy,
                    progress=progress,
                    semaphore=semaphore,
                )
                for asset in assets
            )

            await asyncio.gather(*download_tasks)


async def _get_releases(
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
async def _get_proxy(
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
            task = tg.create_task(test_proxy(client, proxy))
            tasks.append(task)

        done, pending = await asyncio.wait(tasks, return_when=asyncio.FIRST_COMPLETED)

        for p in pending:
            p.cancel()

        for d in done:
            return await d


async def _download_asset(
    client: Session,
    asset: Asset,
    output_dir: Path,
    proxy: Proxy | None,
    *,
    progress: Progress,
    semaphore: asyncio.Semaphore,
):
    url = asset.browser_download_url
    if proxy is not None:
        url = proxy(url)

    # TODO: multi-parts download, progress bar.

    # GET with range

    async with semaphore:
        size = asset.size
        task_id = progress.add_task(asset.name, total=size)

        path = output_dir.joinpath(asset.name)
        async with aiofiles.open(path, "wb") as fp:
            start = 0
            chunk_size = FIRST_CHUNK_SIZE
            speed = FIRST_SPEED
            loop = asyncio.get_running_loop()
            while start < size - 1:
                end = start + chunk_size - 1

                start_time = loop.time()
                await _fetch_segment(client, url, fp, start, end, progress, task_id)
                end_time = loop.time()

                # 动态调整块大小
                speed2 = chunk_size / (end_time - start_time)
                if speed2 >= speed * 2:
                    speed = speed2
                    chunk_size <<= 1
                    chunk_size = min(chunk_size, MAX_CHUNK_SIZE)
                elif speed2 <= speed / 2:
                    speed = speed2
                    chunk_size >>= 1
                    chunk_size = max(chunk_size, MIN_CHUNK_SIZE)

                start = end + 1

            if start < size:
                await _fetch_segment(
                    client, url, fp, start, size - 1, progress, task_id
                )

    progress.remove_task(task_id)

    logger.info(f"downloaded {path}")


# 网络环境复杂，经常因此超时导致出错
# AI 建议使用 tenacity 库提供的 retry 方法
# 这里设置超时重试 3 次，如果还是失败，说明网络环境不适合下载
@retry(
    stop=stop_after_attempt(3),
    wait=wait_exponential(multiplier=1, min=2, max=10),
    retry=retry_if_exception_type((asyncio.TimeoutError,)),
    reraise=True,
)
async def _fetch_segment(
    client: Session,
    url: str,
    fp,
    start: int,
    end: int,
    progress: Progress,
    task_id: TaskID,
):
    async with client.get(url, headers={"Range": f"bytes={start}-{end}"}) as resp:
        resp.raise_for_status()
        data = await resp.read()

    await fp.write(data)
    progress.update(task_id, advance=end - start)
