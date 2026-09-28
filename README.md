# Mattermost Spoiler Plugin

Hide plot twists, puzzle answers and surprise announcements behind a spoiler. Readers choose when to reveal it.

![preview](https://github.com/svelle/mattermost-spoiler-plugin/raw/master/preview.png)

## How to use

In any channel or thread, type:

```
/spoiler The butler did it
```

Spoilers support Markdown, and they can span several lines.

- **Web and desktop:** the spoiler is blurred and marked with a **Spoiler · Click to reveal** label. Click it to reveal the message. Hover over it and press the eye icon to hide it again.
- **Mobile:** the post shows a **Reveal spoiler** button. Tap it and the message appears right below the spoiler, visible only to you. Tap the revealed message to dismiss it.

The hidden text never appears in the post message. Push notifications, email notifications, search results and thread previews show only "🙈 Spoiler — hidden until revealed".

## Installation

1. Download the latest release from the [releases page](https://github.com/svelle/mattermost-spoiler-plugin/releases).
2. In Mattermost, go to **System Console > Plugins > Plugin Management** and upload the `.tar.gz` file.
3. Enable the plugin.

Requires Mattermost Server 10.11 or later. Tested against Mattermost 11.10 and 12.0.

## Upgrading from 1.x

Each spoiler posted with 1.x has the server's full Site URL saved in its **Show** button. Tapping it only works if the Mattermost server can reach itself at that address. Behind a reverse proxy, in Docker, or after a Site URL change, those old buttons fail with "Action integration error."

Spoilers posted with 2.0 use a relative button URL that the server handles internally, so they work regardless of the Site URL.

## Development

You need Go 1.26.7+ and Node.js (see `.nvmrc`).

```
make            # lint, test and build the bundle into dist/
make test       # run server and webapp tests
make deploy     # build and install on a local server (set MM_SERVICESETTINGS_SITEURL and MM_ADMIN_TOKEN)
```

The build tooling follows the [Mattermost plugin starter template](https://github.com/mattermost/mattermost-plugin-starter-template).
