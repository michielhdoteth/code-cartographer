"""
Git Integration for Code Cartographer

Tracks code evolution, history, and changes using git.
"""

import os
import subprocess
import json
from datetime import datetime
from typing import Dict, List, Optional, Any, Tuple
from dataclasses import dataclass, field


@dataclass
class GitCommit:
    """Represents a git commit."""
    hash: str
    short_hash: str
    message: str
    author: str
    author_email: str
    date: str
    timestamp: float
    parent_hash: Optional[str] = None
    files_changed: int = 0
    insertions: int = 0
    deletions: int = 0


@dataclass
class GitFileChange:
    """Represents a file change in a commit."""
    path: str
    change_type: str  # 'A'dded, 'M'odified, 'D'eleted, 'R'enamed
    old_path: Optional[str] = None
    diff: str = ""
    insertions: int = 0
    deletions: int = 0


@dataclass  
class GitBlameInfo:
    """Represents blame information for a line."""
    commit_hash: str
    author: str
    date: str
    line_number: int
    content: str


class GitIntegration:
    """Git integration for code cartographer."""

    def __init__(self, repo_path: str):
        self.repo_path = os.path.abspath(repo_path)
        self.git_dir = os.path.join(self.repo_path, ".git")

    def is_git_repo(self) -> bool:
        """Check if the path is a git repository."""
        return os.path.isdir(self.git_dir) or os.path.isfile(os.path.join(self.git_dir, "HEAD"))

    def _run_git(self, *args) -> Tuple[str, str, int]:
        """Run a git command and return stdout, stderr, returncode."""
        try:
            result = subprocess.run(
                ["git"] + list(args),
                cwd=self.repo_path,
                capture_output=True,
                text=True,
                timeout=30
            )
            return result.stdout, result.stderr, result.returncode
        except FileNotFoundError:
            return "", "git not found", 1
        except subprocess.TimeoutExpired:
            return "", "git command timed out", 1

    def get_commits(self, limit: int = 100, since: Optional[str] = None) -> List[GitCommit]:
        """Get recent commits."""
        cmd = ["log", "--pretty=format:%H|%h|%s|%an|%ae|%ai|%ad", f"-n{limit}"]
        if since:
            cmd.extend(["--since", since])

        stdout, _, rc = self._run_git(*cmd)
        if rc != 0 or not stdout:
            return []

        commits = []
        for line in stdout.strip().split("\n"):
            if not line:
                continue
            parts = line.split("|")
            if len(parts) >= 7:
                commit = GitCommit(
                    hash=parts[0],
                    short_hash=parts[1],
                    message=parts[2],
                    author=parts[3],
                    author_email=parts[4],
                    date=parts[5],
                    timestamp=self._parse_timestamp(parts[5]),
                )
                commits.append(commit)

        return commits

    def get_commit(self, commit_hash: str) -> Optional[GitCommit]:
        """Get a specific commit."""
        stdout, _, rc = self._run_git("show", "--pretty=format:%H|%h|%s|%an|%ae|%ai|%ad", "-n1", commit_hash)
        if rc != 0 or not stdout:
            return None

        parts = stdout.strip().split("|")
        if len(parts) >= 7:
            return GitCommit(
                hash=parts[0],
                short_hash=parts[1],
                message=parts[2],
                author=parts[3],
                author_email=parts[4],
                date=parts[5],
                timestamp=self._parse_timestamp(parts[5]),
            )
        return None

    def get_file_changes(self, commit_hash: str) -> List[GitFileChange]:
        """Get file changes for a commit."""
        stdout, _, rc = self._run_git("show", "--name-status", "--pretty=format:", commit_hash)
        if rc != 0:
            return []

        changes = []
        for line in stdout.strip().split("\n"):
            if not line or line.startswith("..."):
                continue
            parts = line.split("\t")
            if len(parts) >= 2:
                change_type = parts[0]
                path = parts[1]
                old_path = None
                if change_type.startswith("R") and len(parts) >= 3:
                    old_path = parts[2]
                    change_type = "R"
                
                changes.append(GitFileChange(
                    path=path,
                    change_type=change_type,
                    old_path=old_path,
                ))

        return changes

    def get_file_history(self, file_path: str, limit: int = 20) -> List[GitCommit]:
        """Get commit history for a specific file."""
        stdout, _, rc = self._run_git("log", "--pretty=format:%H|%h|%s|%an|%ae|%ai|%ad", 
                                       f"-n{limit}", "--", file_path)
        if rc != 0 or not stdout:
            return []

        commits = []
        for line in stdout.strip().split("\n"):
            if not line:
                continue
            parts = line.split("|")
            if len(parts) >= 7:
                commit = GitCommit(
                    hash=parts[0],
                    short_hash=parts[1],
                    message=parts[2],
                    author=parts[3],
                    author_email=parts[4],
                    date=parts[5],
                    timestamp=self._parse_timestamp(parts[5]),
                )
                commits.append(commit)

        return commits

    def get_diff(self, commit_hash: str, file_path: Optional[str] = None) -> str:
        """Get diff for a commit (optionally for a specific file)."""
        cmd = ["show", "--pretty=format:", commit_hash]
        if file_path:
            cmd.extend(["--", file_path])
        
        stdout, _, rc = self._run_git(*cmd)
        return stdout if rc == 0 else ""

    def get_diff_between(self, from_commit: str, to_commit: str, 
                        file_path: Optional[str] = None) -> str:
        """Get diff between two commits."""
        cmd = ["diff", f"{from_commit}..{to_commit}"]
        if file_path:
            cmd.extend(["--", file_path])

        stdout, _, rc = self._run_git(*cmd)
        return stdout if rc == 0 else ""

    def get_current_branch(self) -> str:
        """Get the current branch name."""
        stdout, _, rc = self._run_git("rev-parse", "--abbrev-ref", "HEAD")
        return stdout.strip() if rc == 0 else "unknown"

    def get_branches(self) -> List[str]:
        """Get all branch names."""
        stdout, _, rc = self._run_git("branch", "-a", "--format=%(refname:short)")
        if rc != 0:
            return []
        return [b.strip() for b in stdout.strip().split("\n") if b.strip()]

    def get_tags(self) -> List[str]:
        """Get all tag names."""
        stdout, _, rc = self._run_git("tag", "--format=%(refname:short)")
        if rc != 0:
            return []
        return [t.strip() for t in stdout.strip().split("\n") if t.strip()]

    def get_active_authors(self, limit: int = 10) -> List[Dict]:
        """Get most active authors by commit count."""
        stdout, _, rc = self._run_git("shortlog", "-s", "-n", f"-n{limit}", "--all")
        if rc != 0:
            return []

        authors = []
        for line in stdout.strip().split("\n"):
            parts = line.strip().split(None, 2)
            if len(parts) >= 2:
                authors.append({
                    "commits": int(parts[0]),
                    "name": parts[1] if len(parts) < 3 else parts[2],
                })
        return authors

    def get_code_churn(self, days: int = 30) -> Dict[str, Dict]:
        """Get code churn statistics for recent changes."""
        since_date = datetime.now().replace(
            hour=0, minute=0, second=0, microsecond=0
        )
        since_date = since_date.fromordinal(since_date.toordinal() - days)

        stdout, _, rc = self._run_git(
            "log", "--since", since_date.strftime("%Y-%m-%d"),
            "--pretty=format:%h|%an|%ad|%s",
            "--numstat"
        )
        
        if rc != 0:
            return {}

        churn = {}
        lines = stdout.strip().split("\n")
        i = 0
        while i < len(lines):
            if "|" in lines[i] and "|" in lines[i + 1]:
                parts = lines[i].split("|")
                if len(parts) >= 3:
                    commit_hash = parts[0]
                    author = parts[1]
                    date = parts[2]
                    
                    j = i + 1
                    file_changes = []
                    while j < len(lines) and "|" not in lines[j]:
                        stat_parts = lines[j].split()
                        if len(stat_parts) >= 2:
                            try:
                                insertions = int(stat_parts[0])
                                deletions = int(stat_parts[1])
                                if j + 1 < len(lines):
                                    file_path = lines[j + 1].strip()
                                    if file_path:
                                        file_changes.append({
                                            "path": file_path,
                                            "insertions": insertions,
                                            "deletions": deletions,
                                        })
                                j += 2
                            except (ValueError, IndexError):
                                j += 1
                        else:
                            j += 1

                    if author not in churn:
                        churn[author] = {
                            "commits": 0,
                            "insertions": 0,
                            "deletions": 0,
                            "files": set()
                        }
                    churn[author]["commits"] += 1
                    for fc in file_changes:
                        churn[author]["insertions"] += fc["insertions"]
                        churn[author]["deletions"] += fc["deletions"]
                        churn[author]["files"].add(fc["path"])
                
                i = j
            else:
                i += 1

        # Convert sets to counts
        for author in churn:
            churn[author]["file_count"] = len(churn[author]["files"])
            del churn[author]["files"]

        return churn

    def get_heatmap_data(self, year: Optional[int] = None) -> List[Dict]:
        """Get commit heatmap data (commits by date)."""
        if year is None:
            year = datetime.now().year

        stdout, _, rc = self._run_git(
            "log", "--since", f"{year}-01-01",
            "--until", f"{year+1}-01-01",
            "--pretty=format:%ad", "--date=short"
        )
        
        if rc != 0:
            return []

        dates = {}
        for line in stdout.strip().split("\n"):
            if line:
                dates[line] = dates.get(line, 0) + 1

        return [{"date": d, "count": c} for d, c in dates.items()]

    def get_stats_summary(self) -> Dict[str, Any]:
        """Get a summary of git statistics."""
        commits, _, _ = self._run_git("rev-list", "--count", "--all")
        files, _, _ = self._run_git("ls-files")
        branches = self.get_branches()
        tags = self.get_tags()
        authors = self.get_active_authors(5)
        
        return {
            "total_commits": int(commits.strip()) if commits.strip().isdigit() else 0,
            "total_files": len([f for f in files.strip().split("\n") if f]) if files else 0,
            "branches": len(branches),
            "current_branch": self.get_current_branch(),
            "tags": len(tags),
            "top_authors": authors,
            "is_git_repo": self.is_git_repo(),
        }

    def export_git_data(self, output_path: Optional[str] = None) -> Dict[str, Any]:
        """Export all git data for visualization."""
        stats = self.get_stats_summary()
        commits = self.get_commits(limit=500)
        heatmap = self.get_heatmap_data()
        churn = self.get_code_churn(30)

        git_data = {
            "generated_at": datetime.now().isoformat(),
            "stats": stats,
            "recent_commits": [
                {
                    "hash": c.hash,
                    "short_hash": c.short_hash,
                    "message": c.message,
                    "author": c.author,
                    "date": c.date,
                    "timestamp": c.timestamp,
                }
                for c in commits[:100]
            ],
            "heatmap": heatmap,
            "code_churn_30d": churn,
        }

        if output_path:
            with open(output_path, "w", encoding="utf-8") as f:
                json.dump(git_data, f, indent=2)

        return git_data

    def _parse_timestamp(self, date_str: str) -> float:
        """Parse ISO date string to timestamp."""
        try:
            return datetime.fromisoformat(date_str).timestamp()
        except ValueError:
            return 0

    def blame_file(self, file_path: str) -> List[GitBlameInfo]:
        """Get blame information for a file."""
        stdout, _, rc = self._run_git("blame", "--line-porcelain", file_path)
        if rc != 0 or not stdout:
            return []

        blame_lines = []
        lines = stdout.split("\n")
        i = 0
        current_commit = None
        current_author = None
        current_date = None

        while i < len(lines):
            line = lines[i]
            if line.startswith("author "):
                current_author = line[7:]
            elif line.startswith("author-time "):
                try:
                    current_date = datetime.fromtimestamp(int(line[12:])).isoformat()
                except ValueError:
                    current_date = ""
            elif line.startswith("author-mail "):
                if current_author and current_author in line:
                    current_author = line[12:].strip("<").strip(">")
            elif line.startswith("summary "):
                if current_commit:
                    blame_lines.append(GitBlameInfo(
                        commit_hash=current_commit,
                        author=current_author or "unknown",
                        date=current_date or "",
                        line_number=len(blame_lines) + 1,
                        content="",
                    ))
            elif line and not line.startswith("filename") and not line.startswith("boundary"):
                # This is a content line
                if current_commit and i > 0:
                    blame_lines.append(GitBlameInfo(
                        commit_hash=current_commit,
                        author=current_author or "unknown",
                        date=current_date or "",
                        line_number=len(blame_lines) + 1,
                        content=line,
                    ))
            elif line.startswith("commit "):
                current_commit = line[7:]
            elif line.startswith("filename "):
                # Skip filename line, next line is content
                i += 1
                continue

            i += 1

        return blame_lines


def get_git_integration(repo_path: str) -> Optional[GitIntegration]:
    """Get git integration for a repository if it's a git repo."""
    git = GitIntegration(repo_path)
    if git.is_git_repo():
        return git
    return None
