package pkg

/*
  {
    "url": "https://api.github.com/repos/kheina/OpenConMan/releases/246894978",
    "assets_url": "https://api.github.com/repos/kheina/OpenConMan/releases/246894978/assets",
    "upload_url": "https://uploads.github.com/repos/kheina/OpenConMan/releases/246894978/assets{?name,label}",
    "html_url": "https://github.com/kheina/OpenConMan/releases/tag/0.0.1%2Bdev-1757712876",
    "id": 246894978,
    "author": {
      "login": "github-actions[bot]",
      "id": 41898282,
      "node_id": "MDM6Qm90NDE4OTgyODI=",
      "avatar_url": "https://avatars.githubusercontent.com/in/15368?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/github-actions%5Bbot%5D",
      "html_url": "https://github.com/apps/github-actions",
      "followers_url": "https://api.github.com/users/github-actions%5Bbot%5D/followers",
      "following_url": "https://api.github.com/users/github-actions%5Bbot%5D/following{/other_user}",
      "gists_url": "https://api.github.com/users/github-actions%5Bbot%5D/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/github-actions%5Bbot%5D/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/github-actions%5Bbot%5D/subscriptions",
      "organizations_url": "https://api.github.com/users/github-actions%5Bbot%5D/orgs",
      "repos_url": "https://api.github.com/users/github-actions%5Bbot%5D/repos",
      "events_url": "https://api.github.com/users/github-actions%5Bbot%5D/events{/privacy}",
      "received_events_url": "https://api.github.com/users/github-actions%5Bbot%5D/received_events",
      "type": "Bot",
      "user_view_type": "public",
      "site_admin": false
    },
    "node_id": "RE_kwDOPYWHg84Ot1GC",
    "tag_name": "0.0.1+dev-1757712876",
    "target_commitish": "main",
    "name": "0.0.1+dev-1757712876",
    "draft": false,
    "immutable": false,
    "prerelease": true,
    "created_at": "2025-08-04T23:03:49Z",
    "updated_at": "2025-09-12T21:35:08Z",
    "published_at": "2025-09-12T21:35:06Z",
    "assets": [
      {
        "url": "https://api.github.com/repos/kheina/OpenConMan/releases/assets/292652831",
        "id": 292652831,
        "node_id": "RA_kwDOPYWHg84RcYcf",
        "name": "conman",
        "label": "",
        "uploader": {
          "login": "github-actions[bot]",
          "id": 41898282,
          "node_id": "MDM6Qm90NDE4OTgyODI=",
          "avatar_url": "https://avatars.githubusercontent.com/in/15368?v=4",
          "gravatar_id": "",
          "url": "https://api.github.com/users/github-actions%5Bbot%5D",
          "html_url": "https://github.com/apps/github-actions",
          "followers_url": "https://api.github.com/users/github-actions%5Bbot%5D/followers",
          "following_url": "https://api.github.com/users/github-actions%5Bbot%5D/following{/other_user}",
          "gists_url": "https://api.github.com/users/github-actions%5Bbot%5D/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/github-actions%5Bbot%5D/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/github-actions%5Bbot%5D/subscriptions",
          "organizations_url": "https://api.github.com/users/github-actions%5Bbot%5D/orgs",
          "repos_url": "https://api.github.com/users/github-actions%5Bbot%5D/repos",
          "events_url": "https://api.github.com/users/github-actions%5Bbot%5D/events{/privacy}",
          "received_events_url": "https://api.github.com/users/github-actions%5Bbot%5D/received_events",
          "type": "Bot",
          "user_view_type": "public",
          "site_admin": false
        },
        "content_type": "application/octet-stream",
        "state": "uploaded",
        "size": 24672384,
        "digest": "sha256:2fe56143b0b4ee764c23888ff7010f945b8638b9e142bf2170aae9d54f6507a7",
        "download_count": 2,
        "created_at": "2025-09-12T21:35:06Z",
        "updated_at": "2025-09-12T21:35:08Z",
        "browser_download_url": "https://github.com/kheina/OpenConMan/releases/download/0.0.1%2Bdev-1757712876/conman"
      }
    ],
    "tarball_url": "https://api.github.com/repos/kheina/OpenConMan/tarball/0.0.1+dev-1757712876",
    "zipball_url": "https://api.github.com/repos/kheina/OpenConMan/zipball/0.0.1+dev-1757712876",
    "body": "todo\n\n\n**Full Changelog**: https://github.com/kheina/OpenConMan/commits/0.0.1+dev-1757712876"
  },
*/

type ReleaseAssets struct {
	Url  string `json:"url"`
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}

type GithubRelease struct {
	// we don't need all of the keys
	Id         int             `json:"id"`
	Url        string          `json:"url"`
	Assets     []ReleaseAssets `json:"assets"`
	Name       string          `json:"name"`
	TagName    string          `json:"tag_name"`
	Prerelease bool            `json:"prerelease"`
}
