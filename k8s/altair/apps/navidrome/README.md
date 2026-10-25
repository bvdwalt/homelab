# Navidrome

## Per-user Favourites / Recently Played playlists

Navidrome's own web UI can't author smart playlist (`.nsp`) criteria — only
manual file drop. Smart playlists are also single-owner: a `.nsp` file's
tracks are computed once against one fixed account's star/play data, so a
shared file shows the same list to every user regardless of who's logged in.
The `"owner"` key in an `.nsp` file is ignored.

To get a real per-user "Favourites" and "Recently Played" playlist for each
account, use [Feishin](https://github.com/jeffvli/feishin) (a Navidrome
client with a smart playlist builder) and create the playlists individually
while logged in as each user:

- **Favourites**: field `Is Favorite`, operator `Is`, value `true`
- **Recently Played**: field `Date Last Played`, operator `Is in the last`,
  value e.g. `30d`; sort by `Date Last Played`, descending (not by Album —
  that's the default and just orders results alphabetically)

Each playlist created this way is owned by whichever user was logged in when
it was created, so it stays correctly per-user.

## Lyrics plugin

[navidrome-lyrics-plugin](https://github.com/J0R6IT0/navidrome-lyrics-plugin)
is fetched by the `fetch-lyrics-plugin` init container (pinned version and
sha256) into `/data/plugins`.

- **Setup**: as admin, enable and approve "Lyrics" under Settings > Plugins.
- **Other providers**: add their hosts to the egress rule first (only
  `lrclib.net` and `api.lyrics.ovh` are allowed).
- **Upgrading**: Renovate bumps `PLUGIN_VERSION` and `PLUGIN_SHA256` together.
- The web UI doesn't show plugin lyrics; use a client such as Feishin.
