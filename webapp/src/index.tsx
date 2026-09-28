import manifest from 'manifest';

import SpoilerPost, {SPOILER_POST_TYPE} from 'components/spoiler_post';

import type {PluginRegistry} from 'types/mattermost-webapp';

export default class Plugin {
    public async initialize(registry: PluginRegistry) {
        registry.registerPostTypeComponent(SPOILER_POST_TYPE, SpoilerPost);
    }
}

declare global {
    interface Window {
        registerPlugin(pluginId: string, plugin: Plugin): void;
        PostUtils: {
            formatText(text: string, options?: Record<string, unknown>): string;
            messageHtmlToComponent(html: string, options?: Record<string, unknown>): React.ReactNode;
        };
    }
}

window.registerPlugin(manifest.id, new Plugin());
