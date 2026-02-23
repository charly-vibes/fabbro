function parseGitHubUrl(url) {
  const match = url.match(
    /^https?:\/\/github\.com\/([^/]+)\/([^/]+)\/blob\/([^/]+)\/(.+)/
  );
  if (!match) return null;
  return { owner: match[1], repo: match[2], branch: match[3], path: match[4] };
}

async function fetchGitHub(owner, repo, branch, path) {
  const res = await fetch(
    `https://api.github.com/repos/${owner}/${repo}/contents/${path}?ref=${branch}`,
    { headers: { Accept: "application/vnd.github.v3.raw" } }
  );

  if (res.status === 404) {
    throw new Error(
      "File not found. Check the URL and make sure the repository is public."
    );
  }
  if (res.status === 403) {
    const remaining = res.headers.get("x-ratelimit-remaining");
    if (remaining === "0") {
      throw new Error(
        "GitHub API rate limit reached (60 requests/hour). Try pasting text directly."
      );
    }
    throw new Error(
      "Access denied. This may be a private repository. Try pasting text directly."
    );
  }
  if (!res.ok) {
    throw new Error(`GitHub API error: ${res.status}`);
  }

  return res.text();
}

async function fetchMarkdown(url) {
  let res;
  try {
    res = await fetch(url, {
      headers: { Accept: "text/markdown, text/html" },
    });
  } catch (err) {
    if (err instanceof TypeError) {
      throw new Error(
        "This URL blocked cross-origin requests. Try a GitHub file URL or paste text directly."
      );
    }
    throw err;
  }

  if (!res.ok) {
    throw new Error(`Could not fetch URL: ${res.status}`);
  }

  return res.text();
}

export async function fetchContent(url) {
  const gh = parseGitHubUrl(url);

  if (gh) {
    const content = await fetchGitHub(gh.owner, gh.repo, gh.branch, gh.path);
    return {
      content,
      source: `${gh.owner}/${gh.repo}/${gh.path}`,
      filename: gh.path.split("/").pop(),
    };
  }

  const content = await fetchMarkdown(url);
  return {
    content,
    source: url,
    filename: new URL(url).pathname.split("/").pop() || "untitled",
  };
}
