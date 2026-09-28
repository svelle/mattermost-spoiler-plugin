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
- **Mobile:** the post shows a **Reveal spoiler** button. Tap it to open the message in a dialog that only you can see.

The hidden text never appears in the post message. Push notifications, email notifications, search results and thread previews show only "🙈 Spoiler — hidden until revealed".

## Installation

1. Download the latest release from the [releases page](https://github.com/svelle/mattermost-spoiler-plugin/releases).
2. In Mattermost, go to **System Console > Plugins > Plugin Management** and upload the `.tar.gz` file.
3. Enable the plugin.

Requires Mattermost Server 10.11 or later. Tested against Mattermost 12.0.

## Upgrading from 1.x

Spoilers posted with 1.x still work. Their **Show** button now opens the same reveal flow as new spoilers. New spoilers use a relative button URL. This fixes the mobile **Show** button on servers where the Site URL isn't reachable from the server itself, for example behind a reverse proxy or when Mattermost runs under a subpath.

## Development

You need Go 1.26.7+ and Node.js (see `.nvmrc`).

```
make            # lint, test and build the bundle into dist/
make test       # run server and webapp tests
make deploy     # build and install on a local server (set MM_SERVICESETTINGS_SITEURL and MM_ADMIN_TOKEN)
```

The build tooling follows the [Mattermost plugin starter template](https://github.com/mattermost/mattermost-plugin-starter-template).
