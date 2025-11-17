from gh2.parser import parse_url


def f(url: str) -> str:
    repo = parse_url(url)
    if repo is None:
        return ""
    return str(repo)


def test_1():
    assert f("x/y") == "x/y"
    assert f("x/y.git") == "x/y"
    assert f("/x/y") == "x/y"
    assert f("/x/y.git") == "x/y"
    assert f("https://github.com/x/y") == "x/y"
    assert f("https://github.com/x/y.git") == "x/y"
    assert f("git@github.com:x/y") == "x/y"
    assert f("git@github.com:x/y.git") == "x/y"
    assert f("https://github.mirror/x/y") == "x/y"
    assert f("https://github.mirror/x/y.git") == "x/y"
    assert f("https://a.b.c/x/y") == "x/y"
    assert f("https://a.b.c/x/y.git") == "x/y"


def test_2():
    assert f("a-b/c_d") == "a-b/c_d"
    assert f("a-b/c_d.git") == "a-b/c_d"
    assert f("a-b/c_d.git.git") == "a-b/c_d.git"
    assert f("a-b/c_d.github.io") == "a-b/c_d.github.io"
    assert f("a-b/c_d.github.io.git") == "a-b/c_d.github.io"
    assert f("a-b/c_d.github.io.git.git") == "a-b/c_d.github.io.git"


def test_3():
    assert parse_url("xy") is None
    assert parse_url("https://github.com/xy") is None


def test_4():
    assert f("a/b/c/x/y") == "a/b"
    assert f("a/b.git/c/x/y") == "a/b.git"
    assert f("a/b/c/x/y.git") == "a/b"
    assert f("a/b.git/c/x/y.git") == "a/b.git"
    assert f("a/b.git/c.git/x.git/y.git") == "a/b.git"
    assert f("/a/b.git/c.git/x.git/y.git") == "a/b.git"
    assert f("https://github.com/x/y/issues") == "x/y"
    assert f("https://github.com/x/y.git/issues") == "x/y.git"
    assert f("https://github.com/x/y/issues.git") == "x/y"
    assert f("https://github.com/x/y.git/issues.git") == "x/y.git"
    assert f("https://github.com/x/y/releases/tag/v1.0") == "x/y"
    assert f("https://github.com/x/y.git/releases/tag/v1.0") == "x/y.git"
    assert f("https://github.com/x/y/releases/tag/v1.0.git") == "x/y"


def test_5():
    assert parse_url("https://github.com/[]/{}") is None
    assert parse_url("git@github.com:[]/{}") is None
    assert parse_url("git@github.com:/[]/{}") is None


def test_6():
    assert parse_url("https:///x/y") is None
    assert parse_url("git@:/x/y") is None
