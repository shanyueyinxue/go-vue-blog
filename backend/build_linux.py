"""Cross-compile the backend for Linux (amd64) from Windows.

This script mirrors backend/build.bat, but keeps the backend directory and
Zig wrapper path derived from the script location, so it can be run from
anywhere with ``python backend/build_linux.py``.
"""

from __future__ import annotations

import os
import shutil
import subprocess
import sys
from pathlib import Path


def main() -> int:
    backend_dir = Path(__file__).resolve().parent
    cc_wrapper = backend_dir / "zig-cc-linux.bat"
    output_binary = backend_dir / "blog"

    if not cc_wrapper.is_file():
        print(f"未找到 Zig CC 包装脚本: {cc_wrapper}", file=sys.stderr)
        return 1

    go_bin = shutil.which("go")
    if not go_bin:
        print("未找到 go，请确认 Go 已安装并加入 PATH。", file=sys.stderr)
        return 1

    env = os.environ.copy()
    env.update(
        {
            "GOOS": "linux",
            "GOARCH": "amd64",
            "CGO_ENABLED": "1",
            "CC": str(cc_wrapper),
        }
    )

    command = [go_bin, "build", "-o", str(output_binary), "."]
    print(f"编译目录: {backend_dir}")
    print(f"输出文件: {output_binary}")
    print("执行命令: " + subprocess.list2cmdline(command))

    try:
        subprocess.run(
            command,
            cwd=backend_dir,
            env=env,
            check=True,
        )
    except subprocess.CalledProcessError as exc:
        print(f"编译失败，go build 退出码: {exc.returncode}", file=sys.stderr)
        return exc.returncode

    print("编译完成。")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
