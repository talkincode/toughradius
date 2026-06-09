#!/usr/bin/env bash

set -euo pipefail

DAYS=30
PR_LIMIT=200
RUN_LIMIT=500
REPO=""
JSON_OUTPUT=""
MARKDOWN_OUTPUT=""

usage() {
  cat <<'USAGE'
Usage: scripts/agent-roadmap-quality-metrics.sh [options]

Collect quality metrics for merged PRs labeled `agent-roadmap`.

Options:
  --days <n>              Time window in days (default: 30)
  --pr-limit <n>          Max closed PRs fetched (default: 200)
  --run-limit <n>         Max CI workflow runs fetched (default: 500)
  --repo <owner/name>     GitHub repository (default: current gh repo)
  --json-output <path>    Write full JSON report to file
  --markdown-output <path> Write markdown summary to file
  -h, --help              Show this help message
USAGE
}

while (($# > 0)); do
  case "$1" in
    --days)
      DAYS="$2"
      shift 2
      ;;
    --pr-limit)
      PR_LIMIT="$2"
      shift 2
      ;;
    --run-limit)
      RUN_LIMIT="$2"
      shift 2
      ;;
    --repo)
      REPO="$2"
      shift 2
      ;;
    --json-output)
      JSON_OUTPUT="$2"
      shift 2
      ;;
    --markdown-output)
      MARKDOWN_OUTPUT="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if ! [[ "$DAYS" =~ ^[0-9]+$ ]] || [ "$DAYS" -lt 1 ]; then
  echo "--days must be a positive integer" >&2
  exit 1
fi
if ! [[ "$PR_LIMIT" =~ ^[0-9]+$ ]] || [ "$PR_LIMIT" -lt 1 ]; then
  echo "--pr-limit must be a positive integer" >&2
  exit 1
fi
if ! [[ "$RUN_LIMIT" =~ ^[0-9]+$ ]] || [ "$RUN_LIMIT" -lt 1 ]; then
  echo "--run-limit must be a positive integer" >&2
  exit 1
fi

for cmd in gh jq git; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "Missing required command: $cmd" >&2
    exit 1
  fi
done

if [ -z "$REPO" ]; then
  REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner)
fi

compute_since_iso() {
  local days="$1"
  if date -u -d "-$days days" '+%Y-%m-%dT%H:%M:%SZ' >/dev/null 2>&1; then
    date -u -d "-$days days" '+%Y-%m-%dT%H:%M:%SZ'
    return
  fi

  if date -u -v-"$days"d '+%Y-%m-%dT%H:%M:%SZ' >/dev/null 2>&1; then
    date -u -v-"$days"d '+%Y-%m-%dT%H:%M:%SZ'
    return
  fi

  echo "Unable to compute since timestamp with system date command" >&2
  exit 1
}

SINCE_ISO=$(compute_since_iso "$DAYS")
GENERATED_AT=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

PRS_JSON=$(gh pr list \
  --repo "$REPO" \
  --state closed \
  --label agent-roadmap \
  --limit "$PR_LIMIT" \
  --json number,title,state,createdAt,closedAt,mergedAt,headRefName,mergeCommit,url)

MERGED_PRS_JSON=$(jq --arg since "$SINCE_ISO" '[
  .[]
  | select(.state == "MERGED")
  | select(.mergedAt != null and .mergedAt >= $since)
]' <<<"$PRS_JSON")

CI_RUNS_JSON=$(gh run list \
  --repo "$REPO" \
  --workflow CI \
  --event pull_request \
  --limit "$RUN_LIMIT" \
  --json databaseId,headBranch,createdAt,status,conclusion,attempt,url)

MAIN_REF="origin/main"
if ! git rev-parse --verify "$MAIN_REF" >/dev/null 2>&1; then
  MAIN_REF="main"
fi

REVERTED_SHA_JSON=$(git log "$MAIN_REF" --format=%B 2>/dev/null \
  | sed -nE 's/^This reverts commit ([0-9a-f]{40})\.?$/\1/p' \
  | sort -u \
  | jq -Rsc 'split("\n") | map(select(length > 0))')

REPORT_JSON=$(jq -n \
  --arg repo "$REPO" \
  --arg generatedAt "$GENERATED_AT" \
  --arg since "$SINCE_ISO" \
  --argjson days "$DAYS" \
  --argjson mergedPRs "$MERGED_PRS_JSON" \
  --argjson ciRuns "$CI_RUNS_JSON" \
  --argjson revertedSHAs "$REVERTED_SHA_JSON" '
  def pct($n; $d):
    if $d == 0 then 0 else ((($n * 10000) / $d) | round) / 100 end;

  def runs_for_pr($pr):
    $ciRuns
    | map(
        select(.headBranch == $pr.headRefName)
        | select(.status == "completed")
        | select(.createdAt >= $pr.createdAt and .createdAt <= $pr.closedAt)
      );

  ($mergedPRs
    | map(
        . as $pr
        | (runs_for_pr($pr)) as $runs
        | ($pr.mergeCommit.oid // "") as $mergeOID
        | $pr + {
            ci: {
              totalRuns: ($runs | length),
              successRuns: ($runs | map(select(.conclusion == "success")) | length),
              failedRuns: ($runs | map(select(.conclusion != "success")) | length),
              rerunRuns: ($runs | map(select((.attempt // 1) > 1)) | length)
            },
            reverted: ($mergeOID != "" and ($revertedSHAs | index($mergeOID) != null))
          }
      )) as $items
  | ($items | map(.ci.totalRuns) | add // 0) as $ciTotal
  | ($items | map(.ci.successRuns) | add // 0) as $ciSuccess
  | ($items | map(.ci.failedRuns) | add // 0) as $ciFailed
  | ($items | map(.ci.rerunRuns) | add // 0) as $ciRerun
  | ($items | map(select(.reverted)) | length) as $rollbackCount
  | {
      repo: $repo,
      generatedAt: $generatedAt,
      window: {
        days: $days,
        since: $since
      },
      metricDefinition: {
        ciPassRate: "completed CI workflow runs tied to merged agent-roadmap PRs in window",
        rollbackRate: "merged agent-roadmap PRs whose merge commit is later reverted on main"
      },
      mergedPRCount: ($items | length),
      ci: {
        totalRuns: $ciTotal,
        successRuns: $ciSuccess,
        failedRuns: $ciFailed,
        rerunRuns: $ciRerun,
        prsWithNoCompletedRuns: ($items | map(select(.ci.totalRuns == 0)) | length),
        prsWithFailedRuns: ($items | map(select(.ci.failedRuns > 0)) | length),
        passRatePct: pct($ciSuccess; $ciTotal)
      },
      rollback: {
        revertedPRCount: $rollbackCount,
        ratePct: pct($rollbackCount; ($items | length))
      },
      prs: $items
    }
')

if [ -n "$JSON_OUTPUT" ]; then
  mkdir -p "$(dirname "$JSON_OUTPUT")"
  printf '%s\n' "$REPORT_JSON" > "$JSON_OUTPUT"
fi

SUMMARY_MD=$(
  jq -r '
    def yn($v): if $v then "yes" else "no" end;

    "# Agent Roadmap Quality Metrics\n\n" +
    "- Repo: `" + .repo + "`\n" +
    "- Window: last " + (.window.days|tostring) + " days (since `" + .window.since + "`)\n" +
    "- Generated at: `" + .generatedAt + "`\n\n" +
    "## Summary\n\n" +
    "| Metric | Value |\n" +
    "| --- | --- |\n" +
    "| Merged agent-roadmap PRs | " + (.mergedPRCount|tostring) + " |\n" +
    "| CI pass rate | " + (.ci.passRatePct|tostring) + "% (" + (.ci.successRuns|tostring) + "/" + (.ci.totalRuns|tostring) + " completed runs) |\n" +
    "| CI rerun runs (`attempt > 1`) | " + (.ci.rerunRuns|tostring) + " |\n" +
    "| PRs with failed CI runs | " + (.ci.prsWithFailedRuns|tostring) + " |\n" +
    "| PRs with no completed CI run | " + (.ci.prsWithNoCompletedRuns|tostring) + " |\n" +
    "| Rollback rate | " + (.rollback.ratePct|tostring) + "% (" + (.rollback.revertedPRCount|tostring) + "/" + (.mergedPRCount|tostring) + " merged PRs) |\n\n" +
    "## Definitions\n\n" +
    "- CI pass rate: " + .metricDefinition.ciPassRate + ".\n" +
    "- Rollback rate: " + .metricDefinition.rollbackRate + ".\n\n" +
    "## PR Breakdown\n\n" +
    "| PR | CI success/total | Rerun runs | Reverted |\n" +
    "| --- | --- | --- | --- |\n" +
    (
      if (.prs | length) == 0 then
        "| (none) | 0/0 | 0 | no |\n"
      else
        (.prs
          | sort_by(.number)
          | reverse
          | map(
              "| [#" + (.number|tostring) + "](" + .url + ") | " +
              (.ci.successRuns|tostring) + "/" + (.ci.totalRuns|tostring) + " | " +
              (.ci.rerunRuns|tostring) + " | " + yn(.reverted) + " |"
            )
          | join("\n")
        ) + "\n"
      end
    )
  ' <<<"$REPORT_JSON"
)

if [ -n "$MARKDOWN_OUTPUT" ]; then
  mkdir -p "$(dirname "$MARKDOWN_OUTPUT")"
  printf '%s\n' "$SUMMARY_MD" > "$MARKDOWN_OUTPUT"
fi

printf '%s\n' "$SUMMARY_MD"
