#!/usr/bin/env python3
"""Collect release context from Git and GitHub PR metadata."""

from __future__ import annotations

import argparse
import json
import re
import shutil
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Any


SEVERITY_ORDER = {"none": 0, "patch": 1, "minor": 2, "major": 3}
SQUASH_PR_RE = re.compile(r"\(#(\d+)\)\s*$")
MERGE_PR_RE = re.compile(r"^Merge pull request #(\d+)\b")
BREAKING_RE = re.compile(r"\bBREAKING(?: CHANGE)?\b|breaking change", re.I)


@dataclass
class CommandResult:
    stdout: str
    stderr: str
    returncode: int


def run(cmd: list[str], check: bool = True, timeout: int | None = None) -> CommandResult:
    try:
        proc = subprocess.run(cmd, text=True, capture_output=True, timeout=timeout)
    except subprocess.TimeoutExpired as exc:
        stdout = exc.stdout if isinstance(exc.stdout, str) else ""
        stderr = exc.stderr if isinstance(exc.stderr, str) else ""
        stderr = stderr or f"command timed out after {timeout}s: {' '.join(cmd)}"
        return CommandResult(stdout, stderr, 124)
    if check and proc.returncode != 0:
        print(proc.stderr.strip() or proc.stdout.strip(), file=sys.stderr)
        raise SystemExit(proc.returncode)
    return CommandResult(proc.stdout, proc.stderr, proc.returncode)


def git(*args: str, check: bool = True) -> CommandResult:
    return run(["git", *args], check=check)


def repo_root() -> Path:
    return Path(git("rev-parse", "--show-toplevel").stdout.strip())


def latest_reachable_tag(ref: str) -> str | None:
    result = git("describe", "--tags", "--abbrev=0", ref, check=False)
    if result.returncode != 0:
        return None
    return result.stdout.strip() or None


def remote_tag_exists(tag: str) -> bool:
    result = git("ls-remote", "--tags", "origin", f"refs/tags/{tag}", check=False)
    return bool(result.stdout.strip())


def local_tag_exists(tag: str) -> bool:
    return git("rev-parse", "-q", "--verify", f"refs/tags/{tag}", check=False).returncode == 0


def parse_version(tag: str) -> tuple[str, list[int]]:
    prefix = "v" if tag.startswith("v") else ""
    raw = tag[1:] if prefix else tag
    parts = raw.split(".")
    if not parts or not all(part.isdigit() for part in parts):
        raise ValueError(f"tag is not numeric-version shaped: {tag}")
    nums = [int(part) for part in parts]
    while len(nums) < 3:
        nums.append(0)
    return prefix, nums


def bump_tag(tag: str, impact: str) -> str | None:
    if impact == "none":
        return None
    prefix, nums = parse_version(tag)
    if impact == "major":
        nums[0] += 1
        for i in range(1, len(nums)):
            nums[i] = 0
    elif impact == "minor":
        nums[1] += 1
        for i in range(2, len(nums)):
            nums[i] = 0
    elif impact == "patch":
        nums[2] += 1
        for i in range(3, len(nums)):
            nums[i] = 0
    else:
        raise ValueError(f"unknown impact: {impact}")
    return prefix + ".".join(str(n) for n in nums[: max(3, len(nums))])


def commits_since(tag: str | None, ref: str) -> list[dict[str, str]]:
    rev = f"{tag}..{ref}" if tag else ref
    result = git("log", "--first-parent", "--format=%H%x1f%s%x1f%b%x1e", rev)
    commits: list[dict[str, str]] = []
    for record in result.stdout.strip("\x1e\n").split("\x1e"):
        if not record.strip():
            continue
        fields = record.strip("\n").split("\x1f", 2)
        if len(fields) < 3:
            continue
        oid, subject, body = fields
        commits.append({"oid": oid, "subject": subject, "body": body.strip()})
    return commits


def pr_numbers_from_commits(commits: list[dict[str, str]]) -> list[int]:
    nums: list[int] = []
    seen: set[int] = set()
    for commit in commits:
        subject = commit["subject"]
        matches = list(SQUASH_PR_RE.finditer(subject)) + list(MERGE_PR_RE.finditer(subject))
        for match in matches:
            num = int(match.group(1))
            if num not in seen:
                nums.append(num)
                seen.add(num)
    return nums


def gh_prs(nums: list[int]) -> list[dict[str, Any]]:
    if not nums:
        return []
    fields = (
        "number,title,state,mergedAt,author,labels,baseRefName,headRefName,"
        "url,additions,deletions"
    )
    result = run(
        ["gh", "pr", "list", "--state", "all", "--limit", "1000", "--json", fields],
        check=False,
        timeout=30,
    )
    if result.returncode != 0:
        return [
            {
                "number": num,
                "metadata_error": result.stderr.strip() or result.stdout.strip(),
            }
            for num in nums
        ]
    by_number = {int(pr["number"]): pr for pr in json.loads(result.stdout)}
    prs: list[dict[str, Any]] = []
    for num in nums:
        prs.append(
            by_number.get(
                num,
                {
                    "number": num,
                    "metadata_error": "PR not returned by gh pr list --state all --limit 1000",
                },
            )
        )
    return prs


def prefix_of(title: str) -> str:
    m = re.match(r"^\s*([a-zA-Z]+)(?:\([^)]+\))?(!)?:", title)
    if m:
        return m.group(1).lower()
    first = title.strip().split(" ", 1)[0].lower().strip("[]():")
    return first


def all_files_match(files: list[dict[str, Any]], prefixes: tuple[str, ...]) -> bool:
    if not files:
        return False
    paths = [str(f.get("path", "")) for f in files]
    return all(path.startswith(prefixes) for path in paths)


def classify_pr(pr: dict[str, Any]) -> tuple[str, str]:
    if pr.get("metadata_error"):
        return "patch", "GitHub metadata unavailable; requires manual review"

    title = str(pr.get("title") or "")
    body = str(pr.get("body") or "")
    files = pr.get("files") or []
    labels = " ".join(str(label.get("name", "")) for label in pr.get("labels") or [])
    text = f"{title}\n{body}\n{labels}"
    prefix = prefix_of(title)

    if BREAKING_RE.search(text) or "!" in title.split(":", 1)[0]:
        return "major", "breaking-change marker"
    if prefix == "feat":
        return "minor", "feature PR"
    if title.lower().startswith("add ") or title.lower().startswith("[radiusd]"):
        return "minor", "new behavior or protocol surface"
    if prefix in {"fix", "perf", "security"} or title.lower().startswith("fix "):
        return "patch", "fix/security/performance PR"
    if prefix == "chore" and ("deps" in title.lower() or "bump " in title.lower()):
        return "patch", "dependency update"
    if all_files_match(files, ("docs/", ".github/", ".agents/")) and prefix in {
        "docs",
        "ci",
        "test",
        "chore",
    }:
        return "none", "docs/CI/test/agent-only change"
    if prefix in {"docs", "ci", "test"}:
        return "none", f"{prefix} PR"
    if prefix in {"refactor", "chore"}:
        return "patch", "non-feature code maintenance; verify manually"
    return "patch", "unclassified merged PR; verify manually"


def classify_commit(commit: dict[str, str]) -> tuple[str, str]:
    subject = commit["subject"]
    text = f"{subject}\n{commit['body']}"
    prefix = prefix_of(subject)
    if BREAKING_RE.search(text) or "!" in subject.split(":", 1)[0]:
        return "major", "breaking-change marker"
    if prefix == "feat":
        return "minor", "feature commit"
    if prefix in {"fix", "perf", "security"} or subject.lower().startswith("fix "):
        return "patch", "fix/security/performance commit"
    if prefix in {"docs", "ci", "test", "chore"}:
        return "none", f"{prefix} commit"
    return "patch", "unclassified commit; verify manually"


def max_impact(items: list[str]) -> str:
    if not items:
        return "none"
    return max(items, key=lambda item: SEVERITY_ORDER[item])


def build_context(ref: str, do_fetch: bool) -> dict[str, Any]:
    if do_fetch:
        git("fetch", "origin", "main", "--tags", "--prune")

    root = repo_root()
    target_sha = git("rev-parse", ref).stdout.strip()
    previous_tag = latest_reachable_tag(ref)
    commits = commits_since(previous_tag, ref)
    pr_nums = pr_numbers_from_commits(commits)

    prs: list[dict[str, Any]] = []
    if shutil.which("gh"):
        for pr in gh_prs(pr_nums):
            impact, reason = classify_pr(pr)
            pr["impact"] = impact
            pr["impact_reason"] = reason
            prs.append(pr)
    else:
        prs = [
            {
                "number": num,
                "metadata_error": "gh CLI not found",
                "impact": "patch",
                "impact_reason": "GitHub metadata unavailable; requires manual review",
            }
            for num in pr_nums
        ]

    commits_without_pr: list[dict[str, Any]] = []
    pr_set = set(pr_nums)
    for commit in commits:
        subject = commit["subject"]
        commit_prs = {
            int(match.group(1))
            for match in list(SQUASH_PR_RE.finditer(subject)) + list(MERGE_PR_RE.finditer(subject))
        }
        if commit_prs & pr_set:
            continue
        impact, reason = classify_commit(commit)
        commits_without_pr.append({**commit, "impact": impact, "impact_reason": reason})

    impacts = [pr["impact"] for pr in prs] + [c["impact"] for c in commits_without_pr]
    recommendation = max_impact(impacts)
    next_tag = None
    tag_available = None
    version_error = None
    if previous_tag and recommendation != "none":
        try:
            next_tag = bump_tag(previous_tag, recommendation)
            tag_available = not (local_tag_exists(next_tag) or remote_tag_exists(next_tag))
        except ValueError as exc:
            version_error = str(exc)

    return {
        "repo": str(root),
        "ref": ref,
        "target_sha": target_sha,
        "previous_tag": previous_tag,
        "commit_count": len(commits),
        "pr_count": len(prs),
        "prs": prs,
        "commits_without_pr": commits_without_pr,
        "recommended_impact": recommendation,
        "suggested_next_tag": next_tag,
        "suggested_tag_available": tag_available,
        "version_error": version_error,
    }


def md(context: dict[str, Any]) -> str:
    lines = [
        "# Release context",
        "",
        f"- Repository: `{context['repo']}`",
        f"- Target: `{context['ref']}` @ `{context['target_sha']}`",
        f"- Previous tag: `{context['previous_tag'] or 'none'}`",
        f"- Commits since previous tag: {context['commit_count']}",
        f"- PRs since previous tag: {context['pr_count']}",
        f"- Recommended impact: `{context['recommended_impact']}`",
    ]
    if context.get("suggested_next_tag"):
        available = "yes" if context.get("suggested_tag_available") else "no"
        lines.append(f"- Suggested next tag: `{context['suggested_next_tag']}` (available: {available})")
    if context.get("version_error"):
        lines.append(f"- Version error: {context['version_error']}")

    lines.extend(["", "## PRs"])
    if context["prs"]:
        for pr in context["prs"]:
            title = pr.get("title") or "(metadata unavailable)"
            url = pr.get("url") or ""
            suffix = f" - {url}" if url else ""
            lines.append(
                f"- #{pr['number']} `{pr['impact']}`: {title} "
                f"({pr['impact_reason']}){suffix}"
            )
    else:
        lines.append("- none")

    lines.extend(["", "## Commits Without PR"])
    if context["commits_without_pr"]:
        for commit in context["commits_without_pr"]:
            lines.append(
                f"- `{commit['oid'][:12]}` `{commit['impact']}`: "
                f"{commit['subject']} ({commit['impact_reason']})"
            )
    else:
        lines.append("- none")
    return "\n".join(lines) + "\n"


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--ref", default="origin/main", help="Release target ref, default: origin/main")
    parser.add_argument("--fetch", action="store_true", help="Fetch origin/main and tags first")
    parser.add_argument("--format", choices=["markdown", "json"], default="markdown")
    args = parser.parse_args()

    context = build_context(args.ref, args.fetch)
    if args.format == "json":
        print(json.dumps(context, ensure_ascii=False, indent=2))
    else:
        print(md(context), end="")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
