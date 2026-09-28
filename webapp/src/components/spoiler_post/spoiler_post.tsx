import React, {useEffect, useMemo, useRef, useState} from 'react';
import type {MouseEvent} from 'react';

import type {Post} from '@mattermost/types/posts';

import './spoiler_post.scss';

export const SPOILER_POST_TYPE = 'custom_spoiler';

const LABEL_SPOILER = 'Spoiler';
const LABEL_REVEAL_HINT = 'Click to reveal';
const LABEL_REVEAL = 'Reveal spoiler';
const LABEL_HIDE = 'Hide spoiler';

type Props = {
    post: Post;
};

export default function SpoilerPost({post}: Props) {
    const [revealed, setRevealed] = useState(false);
    const contentRef = useRef<HTMLDivElement>(null);

    const spoilerText = typeof post.props?.spoiler_text === 'string' ? post.props.spoiler_text : '';

    const content = useMemo(() => {
        const {formatText, messageHtmlToComponent} = window.PostUtils;
        return messageHtmlToComponent(formatText(spoilerText, {atMentions: true}), {mentionHighlight: false});
    }, [spoilerText]);

    // While hidden, keep links and other interactive content inside the spoiler out of
    // the tab order and away from screen readers.
    useEffect(() => {
        const node = contentRef.current;
        if (!node) {
            return;
        }
        if (revealed) {
            node.removeAttribute('inert');
        } else {
            node.setAttribute('inert', '');
        }
    }, [revealed]);

    // Spoilers also render inside clickable cards such as permalink previews, so keep the
    // click from also opening the linked post.
    const toggle = (show: boolean) => (e: MouseEvent) => {
        e.stopPropagation();
        setRevealed(show);
    };

    if (!spoilerText) {
        return <p>{post.message}</p>;
    }

    // The host styles the direct child of the post body (for example `.AutoHeight > div`),
    // so keep the root a plain block and put the sized spoiler box inside it.
    return (
        <div className='spoiler-post-container'>
            <div
                className={revealed ? 'spoiler-post spoiler-post--revealed' : 'spoiler-post'}
                data-testid='spoilerPost'
            >
                <div
                    ref={contentRef}
                    className='spoiler-post__content'
                    aria-hidden={!revealed}
                >
                    {content}
                </div>
                {revealed ? (
                    <button
                        type='button'
                        className='spoiler-post__hide'
                        aria-label={LABEL_HIDE}
                        title={LABEL_HIDE}
                        onClick={toggle(false)}
                    >
                        <i className='icon icon-eye-off-outline'/>
                    </button>
                ) : (
                    <button
                        type='button'
                        className='spoiler-post__cover'
                        aria-label={LABEL_REVEAL}
                        onClick={toggle(true)}
                    >
                        <span className='spoiler-post__pill'>
                            <i className='icon icon-eye-outline'/>
                            <span className='spoiler-post__label'>{LABEL_SPOILER}</span>
                            <span className='spoiler-post__hint'>{LABEL_REVEAL_HINT}</span>
                        </span>
                    </button>
                )}
            </div>
        </div>
    );
}
